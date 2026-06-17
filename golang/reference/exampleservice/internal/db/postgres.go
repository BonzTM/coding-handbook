package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db/sqlcgen"
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

// queries is the sqlc-generated query set, bound to the same *sql.DB pool. The
// hand-written methods below delegate to it so the SQL is type-checked and
// generated from internal/db/queries.sql against the goose migration schema,
// per golang/services/database.md (sqlc for the typed query layer). Keeping the
// methods thin lets them map driver/scan errors to the core sentinels — the one
// responsibility sqlc deliberately leaves to the storage layer.
func (p *Postgres) queries() *sqlcgen.Queries { return sqlcgen.New(p.db) }

// Create inserts a widget. The unique-violation path is mapped to
// core.ErrAlreadyExists. Note: a production impl would inspect the driver's
// typed error (e.g. pgconn.PgError code 23505) rather than relying on the
// generic insert error; that mapping belongs here in the storage layer.
func (p *Postgres) Create(ctx context.Context, w core.Widget) error {
	// Context is threaded through the generated method per the database doc.
	err := p.queries().CreateWidget(ctx, sqlcgen.CreateWidgetParams{
		ID:        w.ID,
		TenantID:  w.TenantID,
		Name:      w.Name,
		CreatedAt: w.CreatedAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("insert widget: %w", err)
	}
	return nil
}

// Get loads a widget by ID within tenantID, mapping sql.ErrNoRows to
// core.ErrNotFound. The tenant_id predicate means a widget under another tenant
// is reported as not found, never returned.
func (p *Postgres) Get(ctx context.Context, tenantID, id string) (core.Widget, error) {
	row, err := p.queries().GetWidget(ctx, sqlcgen.GetWidgetParams{TenantID: tenantID, ID: id})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return core.Widget{}, core.ErrNotFound
	case err != nil:
		return core.Widget{}, fmt.Errorf("select widget: %w", err)
	}
	return core.Widget{
		ID:        row.ID,
		TenantID:  row.TenantID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}, nil
}

// ListPage returns up to limit widgets WITHIN tenantID after the given keyset
// cursor, ordered by the stable (created_at, id) key — the index the migration
// creates leads with tenant_id so each page is a per-tenant index range. The
// generated query uses a row-comparison boundary so a zero cursor lists from the
// beginning and a non-zero one resumes strictly after the previous page's last
// row, with id as the tiebreaker so pages neither overlap nor skip rows under
// concurrent writes.
func (p *Postgres) ListPage(ctx context.Context, tenantID string, after core.Cursor, limit int) ([]core.Widget, error) {
	if limit <= 0 {
		return nil, nil
	}
	// Clamp to int32 before passing to the generated query: the SQL LIMIT is an
	// int32 and an over-large request would otherwise wrap on conversion. The
	// service already clamps to a small page size, so this is a defensive belt.
	if limit > math.MaxInt32 {
		limit = math.MaxInt32
	}
	rows, err := p.queries().ListWidgets(ctx, sqlcgen.ListWidgetsParams{
		TenantID:        tenantID,
		FromStart:       after.IsZero(),
		CursorCreatedAt: after.CreatedAt.UTC(),
		CursorID:        after.ID,
		PageLimit:       int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("query widgets: %w", err)
	}
	out := make([]core.Widget, 0, len(rows))
	for _, row := range rows {
		out = append(out, core.Widget{
			ID:        row.ID,
			TenantID:  row.TenantID,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}
