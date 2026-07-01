// Command examplegrpc is the gRPC service entrypoint. It mirrors the canonical
// thin-main contract: main does nothing but wire process lifecycle. All fallible
// work lives in run so it can return an error and so tests can exercise startup.
// The root context is cancelled on SIGINT/SIGTERM; everything downstream stops
// by observing that single cancellation. Shutdown is ordered and bounded: flip
// health to NOT_SERVING, GracefulStop the gRPC server (bounded, falling back to
// Stop), drain the sidecar, close the store, flush telemetry.
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	grpcapi "github.com/example/examplegrpc/internal/api/grpc"
	"github.com/example/examplegrpc/internal/buildinfo"
	"github.com/example/examplegrpc/internal/config"
	"github.com/example/examplegrpc/internal/core"
	"github.com/example/examplegrpc/internal/telemetry"
)

func main() {
	// One root context for the whole process. Cancelled on the first signal;
	// stop() restores default signal handling so a second signal kills hard.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := run(ctx)
	stop()
	if err != nil {
		// Boundary log: errors are wrapped with %w on the way up and logged
		// exactly once, here, before the process exits non-zero.
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

// systemClock is the production core.Clock, wired here in main. No core package
// reads the wall clock directly.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func run(ctx context.Context) error {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := telemetry.NewLogger(os.Stdout, cfg.Telemetry.LogLevel, cfg.Telemetry.LogJSON)
	logger.Info("starting",
		"service", buildinfo.Name,
		"version", buildinfo.Version,
		"commit", buildinfo.Commit,
	)

	metrics := telemetry.NewPromMetrics(buildinfo.Name)

	tracerProvider, err := telemetry.NewTracerProvider(ctx, cfg.Telemetry, buildinfo.Name, buildinfo.Version)
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}

	readiness := telemetry.NewReadiness(false)

	// Store -> core service.
	st := core.NewMemory()
	svc := core.NewService(st, systemClock{})

	// Auth seam. Local/dev injects a synthetic principal so the service boots
	// offline; AUTH_ENABLED wires the static-token authenticator (a real build
	// swaps in a JWKS verifier behind the same seam).
	var authn grpcapi.Authenticator
	if cfg.Auth.Enabled {
		authn = grpcapi.NewStaticTokenAuthenticator(cfg.Auth.Token, core.Principal{
			Subject:  "static-token",
			TenantID: "default",
			Roles:    []core.Role{core.RoleReader, core.RoleWriter},
		})
		logger.Info("auth enabled (static token)")
	} else {
		authn = grpcapi.NewSyntheticAuthenticator(core.Principal{
			Subject:  "local-dev",
			TenantID: "default",
			Roles:    []core.Role{core.RoleReader, core.RoleWriter},
		})
		logger.Warn("auth disabled (local/dev mode); requests run as a synthetic principal")
	}

	// Transport security seam. ServerTransportCredentials returns nil when no
	// cert/key is configured (local/dev): we then serve insecure and warn loudly.
	// Production MUST configure TLS, and mTLS (client-cert verification) is the
	// default posture for internal service-to-service traffic unless a mesh
	// terminates TLS. A configured-but-unloadable cert is fail-fast here.
	creds, err := grpcapi.ServerTransportCredentials(cfg.GRPC.TLS)
	if err != nil {
		return fmt.Errorf("load tls credentials: %w", err)
	}
	switch {
	case cfg.GRPC.TLS.MutualTLS():
		logger.Info("grpc transport security: mTLS (client certificates required)")
	case cfg.GRPC.TLS.Enabled():
		logger.Info("grpc transport security: server TLS (no client-cert verification)")
	default:
		logger.Warn("grpc transport security DISABLED (insecure listener; local/dev only) — production requires TLS, and mTLS is the default posture for internal service-to-service traffic")
	}

	grpcSrv, healthSrv := grpcapi.NewGRPCServer(cfg.GRPC, grpcapi.Deps{
		Service: svc,
		Logger:  logger,
		Metrics: metrics,
		Authn:   authn,
		Creds:   creds,
	})
	sidecar := grpcapi.NewSidecar(cfg.HTTP, metrics.Handler(), readiness)

	lis, err := grpcapi.Listen(cfg.GRPC.Addr)
	if err != nil {
		return fmt.Errorf("bind gRPC listener: %w", err)
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		readiness.Set(true)
		logger.Info("grpc listening", "addr", cfg.GRPC.Addr)
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("sidecar listening", "addr", sidecar.Addr())
		if err := sidecar.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("sidecar serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		// Shutdown supervisor. Blocks until the root context (signal) or a
		// sibling failure cancels gctx, then drives ordered teardown.
		<-gctx.Done()
		return shutdown(shutdownDeps{
			grpcSrv:        grpcSrv,
			healthSrv:      healthSrv,
			sidecar:        sidecar,
			store:          st,
			readiness:      readiness,
			tracerProvider: tracerProvider,
			logger:         logger,
			grace:          cfg.ShutdownGrace,
		})
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	logger.Info("stopped")
	return nil
}

type shutdownDeps struct {
	grpcSrv        *grpc.Server
	healthSrv      *health.Server
	sidecar        *grpcapi.Sidecar
	store          *core.Memory
	readiness      *telemetry.Readiness
	tracerProvider *telemetry.TracerProvider
	logger         *slog.Logger
	grace          time.Duration
}

// shutdown drains and releases resources in reverse dependency order under a
// bounded grace budget: flip health/readiness to unready, GracefulStop the gRPC
// server (bounded; fall back to a hard Stop), drain the sidecar, close the
// store, then flush telemetry last.
func shutdown(d shutdownDeps) error {
	d.logger.Info("shutting down", "grace", d.grace)

	ctx, cancel := context.WithTimeout(context.Background(), d.grace)
	defer cancel()

	// 1. Stop accepting: flip health to NOT_SERVING and readiness to false so
	//    load balancers and probes stop routing new traffic while RPCs drain.
	d.healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	d.readiness.Set(false)

	// 2. GracefulStop drains in-flight RPCs but can block indefinitely if a
	//    handler hangs, so bound it: race it against the grace deadline and fall
	//    back to a hard Stop that cancels in-flight RPCs.
	done := make(chan struct{})
	go func() {
		d.grpcSrv.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		d.logger.Warn("graceful stop exceeded grace; forcing stop")
		d.grpcSrv.Stop()
		<-done
	}

	// 3. Drain the metrics/probes sidecar.
	if err := d.sidecar.Shutdown(ctx); err != nil {
		return fmt.Errorf("sidecar shutdown: %w", err)
	}

	// 4. Close the store now that no RPC can still be using it.
	if err := d.store.Close(); err != nil {
		return fmt.Errorf("store close: %w", err)
	}

	// 5. Flush telemetry last so the steps above are recorded.
	if err := d.tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("telemetry flush: %w", err)
	}
	return nil
}
