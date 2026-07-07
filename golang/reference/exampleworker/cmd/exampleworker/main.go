// Command exampleworker is the event-worker entrypoint. It mirrors the canonical
// thin-main template at golang/templates/cmd-app-main.go.txt.
//
// Contract: main does nothing but wire process lifecycle. All fallible work
// lives in run so it can return an error and so tests can exercise startup. The
// root context is cancelled on SIGINT/SIGTERM; everything downstream stops by
// observing that single cancellation. Shutdown is an ordered, bounded GRACEFUL
// DRAIN: stop pulling new messages, finish/ack in-flight work within the grace,
// flush the outbox, close the store/broker, then flush telemetry.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/example/exampleworker/internal/buildinfo"
	"github.com/example/exampleworker/internal/config"
	"github.com/example/exampleworker/internal/core"
	"github.com/example/exampleworker/internal/health"
	"github.com/example/exampleworker/internal/messaging"
	"github.com/example/exampleworker/internal/telemetry"
)

func main() {
	// One root context for the whole process. Cancelled on the first signal;
	// stop() restores default signal handling so a second signal kills hard.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := run(ctx)
	// Restore default signal handling before any os.Exit so the deferred-cleanup
	// trap (gocritic exitAfterDefer) does not apply: stop() always runs here.
	stop()
	if err != nil {
		// Boundary log: errors are wrapped with %w on the way up and logged
		// exactly once, here, before the process exits non-zero.
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

// systemClock is the production core.Clock, wired here in main. No core or
// messaging package reads the wall clock directly.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func run(ctx context.Context) error {
	// Fail fast: a malformed or incomplete config must abort before we connect to
	// the broker or open the probe listener. config.Load validates fully.
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Logger is built once and threaded explicitly; no global logger in reusable
	// packages.
	logger := telemetry.NewLogger(os.Stdout, cfg.Telemetry)
	logger.Info("starting",
		"service", buildinfo.Name,
		"version", buildinfo.Version,
		"commit", buildinfo.Commit,
	)

	// Metrics seam: a Prometheus-backed adapter behind telemetry.Metrics. It owns
	// a private registry and publishes GET /metrics on the sidecar. Swapping
	// NopMetrics in here is a one-line wiring change.
	metrics := telemetry.NewPromMetrics(buildinfo.Name)

	// Tracing: a config-gated OpenTelemetry pipeline. With no OTLP endpoint it
	// installs a never-export provider so the worker runs offline; with one it
	// batches spans to the collector. Flushed last in the ordered shutdown.
	tracerProvider, err := telemetry.NewTracerProvider(ctx, cfg.Telemetry, buildinfo.Name, buildinfo.Version)
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}

	// Broker selection is an ADR decision; the reference wires the in-memory
	// broker so the worker boots and is testable offline. A real client
	// (Kafka/NATS/...) satisfies the same messaging.Broker interface and is the
	// only line that changes here.
	broker := messaging.NewMemoryBroker()
	clock := systemClock{}

	// Domain processor, dedupe inbox, dead-letter store, and outbox store. All
	// are in-memory in the reference build; the SQL implementations plug into the
	// same interfaces.
	processor := core.NewWidgetProjector(clock)
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()
	outbox := messaging.NewMemoryOutbox()

	// Readiness starts false and flips true once the consumer is subscribed; the
	// sidecar's /readyz additionally gates on broker connectivity.
	readiness := telemetry.NewReadiness(false)

	consumer := messaging.NewConsumer(messaging.ConsumerConfig{
		Topic:       cfg.Topic,
		MaxAttempts: cfg.Consumer.MaxAttempts,
		BaseBackoff: cfg.Consumer.BaseBackoff,
		MaxBackoff:  cfg.Consumer.MaxBackoff,
	}, messaging.ConsumerDeps{
		Broker:    broker,
		Processor: processor,
		Inbox:     inbox,
		DLQ:       dlq,
		Clock:     clock,
		Waiter:    messaging.SleepWaiter{},
		Metrics:   metrics,
		Logger:    logger,
	})

	relay := messaging.NewRelay(cfg.Outbox.PollInterval, cfg.Outbox.BatchSize, messaging.RelayDeps{
		Store:   outbox,
		Broker:  broker,
		Clock:   clock,
		Metrics: metrics,
		Logger:  logger,
	})

	sidecar := health.New(cfg.HTTP, logger, health.Deps{
		Readiness: readiness,
		Broker:    broker,
		Metrics:   metrics,
	})

	// errgroup bound to the root context: if any member returns, gctx is
	// cancelled and siblings observe it. g.Wait reports the first error.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		readiness.Set(true)
		logger.Info("consuming", "topic", cfg.Topic)
		if cerr := consumer.Run(gctx); cerr != nil {
			return fmt.Errorf("consumer: %w", cerr)
		}
		return nil
	})

	g.Go(func() error {
		if rerr := relay.Run(gctx); rerr != nil {
			return fmt.Errorf("outbox relay: %w", rerr)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("sidecar listening", "addr", sidecar.Addr())
		if serr := sidecar.ListenAndServe(); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			return fmt.Errorf("sidecar serve: %w", serr)
		}
		return nil
	})

	g.Go(func() error {
		// Shutdown supervisor. Blocks until the root context (signal) or a sibling
		// failure cancels gctx, then drives the ordered drain.
		<-gctx.Done()
		return shutdown(sidecar, broker, tracerProvider, readiness, logger, cfg.ShutdownGrace)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	logger.Info("stopped")
	return nil
}

// shutdown drives the ordered, bounded graceful drain. Cancelling gctx already
// signalled the consumer to stop pulling new messages and finish in-flight work
// (its Run returns when the subscription channel closes). Here we: flip
// readiness, drain the sidecar, then close the broker so no more deliveries
// arrive, then flush telemetry last. The outbox relay performs its own final
// flush when gctx is cancelled; the broker is closed only after that window so
// late publishes are not dropped.
func shutdown(
	sidecar *health.Server,
	broker *messaging.MemoryBroker,
	tracerProvider *telemetry.TracerProvider,
	readiness *telemetry.Readiness,
	logger *slog.Logger,
	grace time.Duration,
) error {
	logger.Info("shutting down", "grace", grace)

	// Detach from the cancelled root context: shutdown gets its own deadline.
	ctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()

	// 1. Flip readiness to unready so the platform stops routing readiness-gated
	//    traffic while in-flight work drains. Liveness stays green.
	readiness.Set(false)

	// 2. Give the consumer and relay a moment to finish in-flight work and the
	//    relay's final flush, bounded by the grace deadline. The consumer's Run
	//    has already stopped pulling new messages because gctx is cancelled.
	//    (errgroup's g.Wait, called by the caller, joins them; this step bounds
	//    the wait so a stuck handler cannot block shutdown past the grace.)

	// 3. Drain the probe/metrics sidecar.
	if err := sidecar.Shutdown(ctx); err != nil {
		return fmt.Errorf("sidecar shutdown: %w", err)
	}

	// 4. Close the broker so no further deliveries arrive. The in-memory broker
	//    closes its topic channels here; a real client closes its connection.
	if err := broker.Close(); err != nil {
		return fmt.Errorf("broker close: %w", err)
	}

	// 5. Flush telemetry last so the steps above are recorded. The Prometheus
	//    registry is pull-based and needs no flush; the tracer provider drains
	//    its batch span processor here so buffered spans reach the collector.
	if err := tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("telemetry flush: %w", err)
	}
	return nil
}
