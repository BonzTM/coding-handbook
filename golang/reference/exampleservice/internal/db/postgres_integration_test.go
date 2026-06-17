//go:build integration

// Package db's integration test runs the Postgres store and the embedded goose
// migrations against a LIVE database. It is gated behind the `integration` build
// tag so the default `make verify` stays green offline; run it with:
//
//	TEST_DATABASE_DSN='postgres://user:pass@localhost:5432/exampleservice?sslmode=disable' \
//		go test -tags=integration ./internal/db/...
//
// The pgx stdlib driver is blank-imported so sql.Open("pgx", dsn) works; the
// reference's non-test build deliberately links no driver.
package db_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

// newLivePool opens the pool, applies the embedded migrations, and truncates the
// tables so each test starts from a clean, migrated schema.
func newLivePool(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set; skipping live Postgres integration test")
	}

	ctx := context.Background()
	pool, err := db.OpenDB(ctx, "pgx", config.DatabaseConfig{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
	})
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if _, err := pool.ExecContext(ctx, "TRUNCATE TABLE widgets, idempotency_keys"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

// newLiveStore wraps newLivePool in the Postgres widget store.
func newLiveStore(t *testing.T) *db.Postgres {
	t.Helper()
	return db.NewPostgres(newLivePool(t))
}

// TestPostgresCreateGet exercises the round-trip including the tenant_id column.
func TestPostgresCreateGet(t *testing.T) {
	ctx := context.Background()
	store := newLiveStore(t)

	w := core.Widget{ID: "w1", TenantID: "t1", Name: "Widget One", CreatedAt: time.Now().UTC().Truncate(time.Microsecond)}
	if err := store.Create(ctx, w); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := store.Get(ctx, "t1", "w1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != w.ID || got.TenantID != w.TenantID || got.Name != w.Name || !got.CreatedAt.Equal(w.CreatedAt) {
		t.Errorf("Get = %+v, want %+v", got, w)
	}
}

// TestPostgresKeysetPagination proves the SQL keyset query partitions the set in
// (created_at, id) order across pages with no overlap or gaps, matching the
// in-memory store's behavior.
func TestPostgresKeysetPagination(t *testing.T) {
	ctx := context.Background()
	store := newLiveStore(t)

	base := time.Now().UTC().Truncate(time.Microsecond)
	ids := []string{"w3", "w1", "w5", "w2", "w4"}
	for i, id := range ids {
		w := core.Widget{ID: id, TenantID: "t1", Name: "n", CreatedAt: base.Add(time.Duration(i) * time.Second)}
		if err := store.Create(ctx, w); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}

	var got []string
	after := core.Cursor{}
	for page := 0; page < len(ids)+1; page++ {
		batch, err := store.ListPage(ctx, "t1", after, 2)
		if err != nil {
			t.Fatalf("ListPage: %v", err)
		}
		if len(batch) == 0 {
			break
		}
		if len(batch) > 2 {
			t.Fatalf("page size = %d, want <= 2", len(batch))
		}
		for _, b := range batch {
			got = append(got, b.ID)
		}
		last := batch[len(batch)-1]
		after = core.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	// created_at increases with insertion index, so the keyset order is the
	// insertion order — NOT the alphabetical id order — proving created_at is the
	// primary sort key and id is only the tiebreaker.
	want := []string{"w3", "w1", "w5", "w2", "w4"}
	if len(got) != len(want) {
		t.Fatalf("walked %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestPostgresKeysetTiebreak proves the id tiebreaker in the SQL row-comparison:
// rows that share a created_at are ordered by id and the cursor resumes
// correctly across the tie boundary.
func TestPostgresKeysetTiebreak(t *testing.T) {
	ctx := context.Background()
	store := newLiveStore(t)

	ts := time.Now().UTC().Truncate(time.Microsecond)
	for _, id := range []string{"c", "a", "b"} {
		if err := store.Create(ctx, core.Widget{ID: id, TenantID: "t1", Name: "n", CreatedAt: ts}); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}

	first, err := store.ListPage(ctx, "t1", core.Cursor{}, 2)
	if err != nil {
		t.Fatalf("ListPage page 1: %v", err)
	}
	if len(first) != 2 || first[0].ID != "a" || first[1].ID != "b" {
		t.Fatalf("page 1 = %v, want [a b]", ids(first))
	}
	last := first[len(first)-1]
	second, err := store.ListPage(ctx, "t1", core.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}, 2)
	if err != nil {
		t.Fatalf("ListPage page 2: %v", err)
	}
	if len(second) != 1 || second[0].ID != "c" {
		t.Fatalf("page 2 = %v, want [c]", ids(second))
	}
}

func ids(ws []core.Widget) []string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = w.ID
	}
	return out
}

// TestPostgresTenantIsolation proves the SQL store scopes reads by tenant_id: a
// widget written by t1 is not readable by t2, and the same id may coexist across
// tenants under the composite primary key.
func TestPostgresTenantIsolation(t *testing.T) {
	ctx := context.Background()
	store := newLiveStore(t)

	ts := time.Now().UTC().Truncate(time.Microsecond)
	if err := store.Create(ctx, core.Widget{ID: "shared", TenantID: "t1", Name: "t1", CreatedAt: ts}); err != nil {
		t.Fatalf("create t1: %v", err)
	}
	if err := store.Create(ctx, core.Widget{ID: "shared", TenantID: "t2", Name: "t2", CreatedAt: ts}); err != nil {
		t.Fatalf("create t2 (same id, other tenant): %v", err)
	}

	// Cross-tenant read returns the row owned by the asking tenant only.
	if got, err := store.Get(ctx, "t1", "shared"); err != nil || got.Name != "t1" {
		t.Fatalf("Get(t1) = %+v, %v; want t1", got, err)
	}
	// A tenant with no such row gets not-found, never another tenant's data.
	if _, err := store.Get(ctx, "t3", "shared"); !errors.Is(err, core.ErrNotFound) {
		t.Errorf("Get(t3) = %v, want ErrNotFound", err)
	}
	list, err := store.ListPage(ctx, "t3", core.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListPage(t3): %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListPage(t3) len = %d, want 0 (cross-tenant isolation)", len(list))
	}
}

// TestPostgresIdempotency proves the SQL idempotency store: first use leases,
// a same-hash duplicate replays the stored response, an in-flight duplicate is
// reported as in-flight, and a different-hash reuse is a mismatch.
func TestPostgresIdempotency(t *testing.T) {
	ctx := context.Background()
	pool := newLivePool(t)
	store := db.NewPostgresIdempotency(pool, time.Hour)

	now := time.Now().UTC().Truncate(time.Microsecond)
	key := core.IdempotencyKey{TenantID: "t1", Route: "POST /widgets", Key: "abc"}

	// First use leases.
	rec, leased, err := store.Begin(ctx, key, "hash-1", now)
	if err != nil || !leased || rec != nil {
		t.Fatalf("first Begin = (%v, %v, %v), want (nil, true, nil)", rec, leased, err)
	}

	// In-flight duplicate before Complete.
	if _, _, inflightErr := store.Begin(ctx, key, "hash-1", now); !errors.Is(inflightErr, core.ErrIdempotencyInFlight) {
		t.Fatalf("in-flight Begin = %v, want ErrIdempotencyInFlight", inflightErr)
	}

	// Complete records the response.
	body := []byte(`{"id":"w1"}`)
	if completeErr := store.Complete(ctx, key, core.IdempotencyRecord{RequestHash: "hash-1", ResponseStatus: 201, ResponseBody: body}, now); completeErr != nil {
		t.Fatalf("Complete: %v", completeErr)
	}

	// Same-hash duplicate replays.
	rec, leased, err = store.Begin(ctx, key, "hash-1", now)
	if err != nil || leased || rec == nil {
		t.Fatalf("replay Begin = (%v, %v, %v), want (record, false, nil)", rec, leased, err)
	}
	if rec.ResponseStatus != 201 || string(rec.ResponseBody) != string(body) {
		t.Errorf("replay record = %+v, want status 201 body %s", rec, body)
	}

	// Different-hash reuse is a mismatch.
	if _, _, err := store.Begin(ctx, key, "hash-2", now); !errors.Is(err, core.ErrIdempotencyKeyMismatch) {
		t.Fatalf("mismatch Begin = %v, want ErrIdempotencyKeyMismatch", err)
	}
}
