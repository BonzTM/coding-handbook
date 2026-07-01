// Command exampleservice is the service entrypoint. It mirrors the canonical
// thin-main template at golang/templates/cmd-app-main.go.txt.
//
// Contract: main does nothing but wire process lifecycle. All fallible work
// lives in run so it can return an error and so tests can exercise startup. The
// root context is cancelled on SIGINT/SIGTERM; everything downstream stops by
// observing that single cancellation. Shutdown is ordered and bounded: flip
// readiness, drain HTTP, close the store, flush telemetry.
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

	// Blank-import the pure-Go pgx stdlib driver so sql.Open("pgx", dsn) works.
	// pgx/v5/stdlib registers the "pgx" database/sql driver via its init(); it is
	// pure Go (no cgo), so the binary stays a static CGO_ENABLED=0 build. The
	// reference default (no DSN) never opens it; it is linked so a DB-backed
	// deployment can run without any further wiring.
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/sync/errgroup"

	httpapi "github.com/example/exampleservice/internal/api/http"
	"github.com/example/exampleservice/internal/auth"
	"github.com/example/exampleservice/internal/buildinfo"
	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
	"github.com/example/exampleservice/internal/telemetry"
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

// systemClock is the production core.Clock, wired here in main. No core package
// reads the wall clock directly.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func run(ctx context.Context) error {
	// Fail fast: a malformed or incomplete config must abort before we open
	// listeners or external connections. config.Load validates fully.
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Logger is built once and threaded explicitly; no global logger in
	// reusable packages.
	logger := telemetry.NewLogger(os.Stdout, cfg.Telemetry)
	logger.Info("starting",
		"service", buildinfo.Name,
		"version", buildinfo.Version,
		"commit", buildinfo.Commit,
	)

	// One-shot migration mode (-migrate): apply the embedded goose migrations
	// against DB_DSN and exit. This is the production path for schema changes —
	// a deployment runs the SAME image as a migration Job (args: ["-migrate"])
	// ahead of the rollout — so it short-circuits here, before any server
	// wiring. Config validation has already required a DSN.
	// DB_MIGRATE_ON_STARTUP below remains the dev/CI self-migrate convenience.
	if cfg.Migrate {
		return migrate(ctx, cfg, logger)
	}

	// Metrics seam: a Prometheus-backed adapter behind the telemetry.Metrics
	// interface. It owns a private registry and publishes GET /metrics (mounted
	// ahead of the heavy middleware so scrapes are not access-logged or counted).
	// Swapping NopMetrics/ExpvarMetrics in here is a one-line wiring change.
	metrics := telemetry.NewPromMetrics(buildinfo.Name)

	// Tracing: a config-gated OpenTelemetry pipeline. With no OTLP endpoint it
	// installs a never-export provider so the service runs offline; with one it
	// batches spans to the collector. It also sets the global W3C propagator.
	// Flushed last in the ordered shutdown.
	tracerProvider, err := telemetry.NewTracerProvider(ctx, cfg.Telemetry, buildinfo.Name, buildinfo.Version)
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}

	// Readiness starts false and flips true once dependencies are wired and the
	// listener is about to serve; shutdown flips it back to false first.
	readiness := telemetry.NewReadiness(false)

	// Store selection. The reference default is the in-memory store so the
	// service boots offline with no external dependency. A DSN selects the
	// database/sql path: the pure-Go pgx driver is blank-imported above, so
	// sql.Open("pgx", dsn) resolves and db.OpenDB applies the four pool limits and
	// pings before the listener accepts traffic. closeStore is nil for the
	// in-memory store (it holds no resources) and the pool's Close for the
	// DB-backed build, so shutdown releases the pool only when one was opened.
	var st core.Store
	var closeStore func() error
	if cfg.Database.DSN != "" {
		pool, derr := db.OpenDB(ctx, "pgx", cfg.Database)
		if derr != nil {
			return fmt.Errorf("open database: %w", derr)
		}
		// Apply the embedded goose migrations before serving when enabled. This is
		// an explicit, ordered step gated by DB_MIGRATE_ON_STARTUP; it runs before
		// readiness flips so the schema is current when the first request lands.
		if cfg.Database.MigrateOnStartup {
			if merr := db.Migrate(ctx, pool); merr != nil {
				_ = pool.Close()
				return fmt.Errorf("migrate: %w", merr)
			}
		}
		pg := db.NewPostgres(pool)
		st = pg
		closeStore = pg.Close
		logger.Info("using postgres store", "migrate_on_startup", cfg.Database.MigrateOnStartup)
	} else {
		st = db.NewMemory()
		logger.Info("using in-memory store (no DB_DSN)")
	}

	// Identity layer. Auth is config-gated: with AUTH_ENABLED=true we build the
	// JWKS-backed verifier (fail-fast if the JWKS is unreachable); otherwise the
	// verifier is nil and the HTTP layer runs in local/dev mode with a synthetic
	// principal so the service boots offline without an identity provider. The
	// JWKS refresh goroutine is bound to the root context.
	var verifier auth.Verifier
	if cfg.Auth.Enabled {
		v, verr := auth.NewJWKSVerifier(ctx, cfg.Auth.JWKSURL, cfg.Auth.Issuer, cfg.Auth.Audience)
		if verr != nil {
			return fmt.Errorf("init auth verifier: %w", verr)
		}
		verifier = v
		logger.Info("auth enabled", "issuer", cfg.Auth.Issuer, "audience", cfg.Auth.Audience)
	} else {
		logger.Warn("auth disabled (local/dev mode); requests run as a synthetic principal")
	}

	// Idempotency store for unsafe writes. The reference uses the in-memory store;
	// a DB-backed build would wire the SQL store (see internal/db) so the response
	// persists in the same transaction as the write.
	idem := db.NewMemoryIdempotency(cfg.Idempotency.TTL)

	// Audit logger: a SEPARATE slog handler routed to its own sink for
	// security-relevant events (authn failure, authz denial, data-mutating write),
	// per operations/security.md ### Audit Logging. It is deliberately NOT the app
	// logger: audit records are evidence with their own retention and access
	// controls. The reference routes it to stderr so it is a distinct stream from
	// the app logger on stdout; a deployment points it at the org's audit sink.
	auditLogger := telemetry.NewAuditLogger(os.Stderr, systemClock{})

	// Explicit constructor wiring: store -> core service -> HTTP server.
	svc := core.NewService(st, systemClock{})
	srv := httpapi.New(cfg.HTTP, svc, logger, metrics, readiness, httpapi.Deps{
		Verifier:    verifier,
		Idempotency: idem,
		Clock:       systemClock{},
		Audit:       auditLogger,
	})

	// errgroup bound to the root context: if any member returns, gctx is
	// cancelled and siblings observe it. g.Wait reports the first error.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		// Dependencies are wired and the listener is about to accept: ready.
		readiness.Set(true)
		logger.Info("http listening", "addr", srv.Addr())
		// ErrServerClosed is the expected outcome of a clean Shutdown.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		// Shutdown supervisor. Blocks until the root context (signal) or a
		// sibling failure cancels gctx, then drives ordered teardown.
		<-gctx.Done()
		return shutdown(srv, closeStore, tracerProvider, logger, cfg.ShutdownGrace)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	logger.Info("stopped")
	return nil
}

// migrate is the one-shot -migrate mode: open the pool, apply all pending
// embedded goose migrations, close the pool. main maps a nil return to exit 0
// and any error to a logged failure with exit 1, which is exactly the contract
// a migration Job needs. It reuses the same db.OpenDB/db.Migrate pair as the
// DB_MIGRATE_ON_STARTUP path, so the two modes cannot drift.
func migrate(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := db.OpenDB(ctx, "pgx", cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	if merr := db.Migrate(ctx, pool); merr != nil {
		_ = pool.Close()
		return fmt.Errorf("migrate: %w", merr)
	}
	if err := pool.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	logger.Info("migrations applied")
	return nil
}

// shutdown drains and releases resources in reverse dependency order under a
// bounded grace budget: flip readiness, drain HTTP with a FRESH deadline, then
// (for a DB-backed build) close the store, then flush telemetry last. The
// in-memory store needs no close; the ordering comment documents where a real
// st.Close() would go.
func shutdown(srv *httpapi.Server, closeStore func() error, tracerProvider *telemetry.TracerProvider, logger *slog.Logger, grace time.Duration) error {
	logger.Info("shutting down", "grace", grace)

	// Detach from the cancelled root context: shutdown gets its own deadline.
	ctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()

	// 1. Flip readiness to unready so load balancers stop routing new traffic
	//    while existing requests drain. Liveness stays green.
	srv.SetReady(false)

	// 2. Stop accepting connections and wait for in-flight requests to finish,
	//    bounded by the fresh grace deadline.
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	// 3. Close the store now that no request can still be using it. closeStore is
	//    nil for the in-memory store (it holds no resources) and the pool's Close
	//    for the DB-backed build, which releases all pooled connections.
	if closeStore != nil {
		if err := closeStore(); err != nil {
			return fmt.Errorf("close store: %w", err)
		}
	}

	// 4. Flush telemetry last so the steps above are recorded. The Prometheus
	//    registry is pull-based and needs no flush; the tracer provider drains
	//    its batch span processor here so buffered spans reach the collector.
	if err := tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("telemetry flush: %w", err)
	}
	return nil
}
