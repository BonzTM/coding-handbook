package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

// migrationsFS embeds the goose-tagged SQL migrations so a built binary carries
// its own schema and can apply it on startup without shipping loose .sql files,
// per golang/services/database.md (migrations live with the code and run as an
// explicit, ordered step). The runner below points goose at this FS.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrationsDialect is the goose SQL dialect. The reference targets Postgres;
// it is a named constant so the one place that must change for another engine
// is obvious.
const migrationsDialect = "postgres"

// Migrate applies all pending goose migrations from the embedded FS to db,
// bringing the schema up to the latest version. It is config-gated by the
// caller (main runs it only when DB_MIGRATE_ON_STARTUP is set) so the default
// in-memory path never touches a database. The driver behind db must already be
// registered; Migrate does not open the pool.
//
// goose's package-level configuration (SetBaseFS, SetDialect) is mutated under
// the hood, so Migrate is not safe to call concurrently with other goose use;
// it is intended to run once during startup before the listener accepts traffic.
func Migrate(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect(migrationsDialect); err != nil {
		return fmt.Errorf("set goose dialect %q: %w", migrationsDialect, err)
	}
	// "migrations" is the path inside the embedded FS that holds the .sql files.
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
