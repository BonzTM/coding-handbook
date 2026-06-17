package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db/sqlcgen"
)

// PostgresIdempotency is a database/sql-backed core.IdempotencyStore. Like
// Postgres it compiles against the standard library only (no driver linked); a
// DB-backed build blank-imports a driver and wires it in main. It persists the
// idempotency record in the same database as the widget write so the recipe's
// atomicity guarantee (response stored with the side effect) holds; see
// recipes/add-idempotent-write.md.
//
// The TTL is enforced in SQL via the expires_at column: Begin reclaims an
// expired row on conflict and Get filters on expires_at, so an abandoned
// in-flight key does not wedge the slot forever.
type PostgresIdempotency struct {
	db  *sql.DB
	ttl time.Duration
}

// Compile-time proof that *PostgresIdempotency satisfies the consumer-defined
// core.IdempotencyStore contract.
var _ core.IdempotencyStore = (*PostgresIdempotency)(nil)

// NewPostgresIdempotency wraps an already-opened pool with the given TTL. A
// non-positive ttl panics: an unbounded store is a configuration error.
func NewPostgresIdempotency(db *sql.DB, ttl time.Duration) *PostgresIdempotency {
	if ttl <= 0 {
		panic("db: idempotency TTL must be positive")
	}
	return &PostgresIdempotency{db: db, ttl: ttl}
}

func (p *PostgresIdempotency) queries() *sqlcgen.Queries { return sqlcgen.New(p.db) }

// Begin reserves the key (or reclaims an expired row) with an atomic
// insert-or-conditional-update CAS. When the CAS affects a row this caller owns
// a fresh lease; otherwise a live row already exists and its stored state
// (read with GetIdempotency) decides replay vs. in-flight vs. mismatch.
func (p *PostgresIdempotency) Begin(ctx context.Context, key core.IdempotencyKey, requestHash string, now time.Time) (*core.IdempotencyRecord, bool, error) {
	q := p.queries()
	affected, err := q.InsertIdempotency(ctx, sqlcgen.InsertIdempotencyParams{
		TenantID:       key.TenantID,
		Route:          key.Route,
		IdempotencyKey: key.Key,
		RequestHash:    requestHash,
		CreatedAt:      now.UTC(),
		ExpiresAt:      now.Add(p.ttl).UTC(),
	})
	if err != nil {
		return nil, false, fmt.Errorf("begin idempotency: %w", err)
	}
	if affected > 0 {
		// Fresh lease (new row or reclaimed-expired row): process the request.
		return nil, true, nil
	}

	// A live row already exists. Read it to decide the outcome.
	row, err := q.GetIdempotency(ctx, sqlcgen.GetIdempotencyParams{
		TenantID:       key.TenantID,
		Route:          key.Route,
		IdempotencyKey: key.Key,
		ExpiresAt:      now.UTC(),
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Raced with an expiry/release between the insert and the read; report
		// in-flight so the client retries rather than double-processing.
		return nil, false, core.ErrIdempotencyInFlight
	case err != nil:
		return nil, false, fmt.Errorf("read idempotency: %w", err)
	case !row.ResponseStatus.Valid:
		// Reserved but not completed: a concurrent duplicate is in flight.
		return nil, false, core.ErrIdempotencyInFlight
	case row.RequestHash != requestHash:
		// Completed under a different request body: misuse.
		return nil, false, core.ErrIdempotencyKeyMismatch
	default:
		status, convErr := int32ToStatus(row.ResponseStatus.Int32)
		if convErr != nil {
			return nil, false, convErr
		}
		return &core.IdempotencyRecord{
			RequestHash:    row.RequestHash,
			ResponseStatus: status,
			ResponseBody:   row.ResponseBody,
		}, false, nil
	}
}

// Complete records the final response for a leased key.
func (p *PostgresIdempotency) Complete(ctx context.Context, key core.IdempotencyKey, rec core.IdempotencyRecord, _ time.Time) error {
	if rec.ResponseStatus < 0 || rec.ResponseStatus > math.MaxInt32 {
		return fmt.Errorf("complete idempotency: status %d out of range", rec.ResponseStatus)
	}
	err := p.queries().CompleteIdempotency(ctx, sqlcgen.CompleteIdempotencyParams{
		TenantID:       key.TenantID,
		Route:          key.Route,
		IdempotencyKey: key.Key,
		ResponseStatus: sql.NullInt32{Int32: int32(rec.ResponseStatus), Valid: true},
		ResponseBody:   rec.ResponseBody,
	})
	if err != nil {
		return fmt.Errorf("complete idempotency: %w", err)
	}
	return nil
}

// Release abandons a leased key so a failed request is retryable.
func (p *PostgresIdempotency) Release(ctx context.Context, key core.IdempotencyKey) error {
	err := p.queries().ReleaseIdempotency(ctx, sqlcgen.ReleaseIdempotencyParams{
		TenantID:       key.TenantID,
		Route:          key.Route,
		IdempotencyKey: key.Key,
	})
	if err != nil {
		return fmt.Errorf("release idempotency: %w", err)
	}
	return nil
}

// int32ToStatus narrows a stored status to an int, rejecting an impossible
// negative value rather than silently wrapping.
func int32ToStatus(v int32) (int, error) {
	if v < 0 {
		return 0, errors.New("idempotency: stored response_status is negative")
	}
	return int(v), nil
}
