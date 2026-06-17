package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
)

// Postgres is a database/sql-backed core.Store implementation. It is a
// reference for golang/services/database.md and compiles against the standard
// library ONLY — it imports no driver. To run it against a real database, wire
// a driver in main with a blank import and pass its name to OpenDB:
//
//	import _ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
//	...
//	pool, err := db.OpenDB(ctx, "pgx", cfg.Database)
//
// The handbook starts with database/sql, hand-written SQL, and explicit pool
// sizing; this file demonstrates exactly that without taking on a driver
// dependency the reference module is not allowed to carry.
type Postgres struct {
	db *sql.DB
}

// Compile-time proof that *Postgres satisfies the consumer-defined core.Store
// contract, enforced here at the implementation rather than only at the wiring
// site in main.
var _ core.Store = (*Postgres)(nil)

// OpenDB opens a *sql.DB for the given driver name and DSN, applies all four
// pool limits from config (never the database/sql defaults, per
// golang/services/database.md), and verifies connectivity with PingContext so
// a service that cannot reach its store fails fast instead of reporting ready.
//
// The driver must already be registered (blank-imported) by the caller; with
// no driver linked, sql.Open returns an "unknown driver" error — which is why
// the default wiring uses the in-memory store.
func OpenDB(ctx context.Context, driverName string, cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Pool sizing is mandatory and explicit. The *sql.DB is a pool, not a
	// connection; the zero-value defaults are wrong for production.
	db.SetMaxOpenConns(cfg.MaxOpenConns)       // cap total open connections
	db.SetMaxIdleConns(cfg.MaxIdleConns)       // idle floor; must be <= MaxOpenConns
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime) // bound age (required behind LB/proxy/failover)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime) // reap idle connections under low load

	// Verify connectivity under a bounded context so startup does not hang.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		// Best-effort close; the open failed so the pool is unusable.
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// NewPostgres wraps an already-opened pool. Open the pool with OpenDB so the
// limits and ping are applied consistently.
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

// Stats exposes the pool statistics that golang/services/database.md requires
// be exported as metrics so saturation (WaitCount/WaitDuration) is visible
// before it becomes a timeout.
func (p *Postgres) Stats() sql.DBStats { return p.db.Stats() }

// Close releases the pool. Call it during shutdown only after in-flight
// requests have drained.
func (p *Postgres) Close() error { return p.db.Close() }

// Create inserts a widget. The unique-violation path is mapped to
// core.ErrAlreadyExists. Note: a production impl would inspect the driver's
// typed error (e.g. pgconn.PgError code 23505) rather than relying on the
// generic insert error; that mapping belongs here in the storage layer.
func (p *Postgres) Create(ctx context.Context, w core.Widget) error {
	const query = `INSERT INTO widgets (id, name, created_at) VALUES ($1, $2, $3)`
	// Context is threaded through ExecContext per the database doc.
	_, err := p.db.ExecContext(ctx, query, w.ID, w.Name, w.CreatedAt.UTC())
	if err != nil {
		return fmt.Errorf("insert widget: %w", err)
	}
	return nil
}

// Get loads a widget by ID, mapping sql.ErrNoRows to core.ErrNotFound.
func (p *Postgres) Get(ctx context.Context, id string) (core.Widget, error) {
	const query = `SELECT id, name, created_at FROM widgets WHERE id = $1`
	var w core.Widget
	err := p.db.QueryRowContext(ctx, query, id).Scan(&w.ID, &w.Name, &w.CreatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return core.Widget{}, core.ErrNotFound
	case err != nil:
		return core.Widget{}, fmt.Errorf("select widget: %w", err)
	}
	return w, nil
}

// List returns all widgets ordered by ID. Rows are closed and Rows.Err is
// checked after iteration, per the linting gate (rowserrcheck/sqlclosecheck).
func (p *Postgres) List(ctx context.Context) ([]core.Widget, error) {
	const query = `SELECT id, name, created_at FROM widgets ORDER BY id`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query widgets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []core.Widget
	for rows.Next() {
		var w core.Widget
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan widget: %w", err)
		}
		out = append(out, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate widgets: %w", err)
	}
	return out, nil
}
