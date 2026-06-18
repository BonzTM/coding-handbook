package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

func idemKey() core.IdempotencyKey {
	return core.IdempotencyKey{TenantID: "t1", Route: "POST /widgets", Key: "k1"}
}

func TestMemoryIdempotencyLeaseAndReplay(t *testing.T) {
	ctx := context.Background()
	store := db.NewMemoryIdempotency(time.Hour)
	now := time.Unix(1700000000, 0).UTC()
	key := idemKey()

	// First use leases.
	rec, leased, err := store.Begin(ctx, key, "hash-1", now)
	if err != nil || !leased || rec != nil {
		t.Fatalf("first Begin = (%v, %v, %v), want (nil, true, nil)", rec, leased, err)
	}

	// In-flight duplicate before Complete -> 409 condition.
	if _, _, inflightErr := store.Begin(ctx, key, "hash-1", now); !errors.Is(inflightErr, core.ErrIdempotencyInFlight) {
		t.Fatalf("in-flight Begin = %v, want ErrIdempotencyInFlight", inflightErr)
	}

	// Complete records the response.
	body := []byte(`{"id":"w1"}`)
	if completeErr := store.Complete(ctx, key, core.IdempotencyRecord{RequestHash: "hash-1", ResponseStatus: 201, ResponseBody: body}, now); completeErr != nil {
		t.Fatalf("Complete: %v", completeErr)
	}

	// Same-hash duplicate replays the stored response.
	rec, leased, err = store.Begin(ctx, key, "hash-1", now)
	if err != nil || leased || rec == nil {
		t.Fatalf("replay Begin = (%v, %v, %v), want (record, false, nil)", rec, leased, err)
	}
	if rec.ResponseStatus != 201 || string(rec.ResponseBody) != string(body) {
		t.Errorf("replay = %+v, want status 201 body %s", rec, body)
	}

	// Mutating the returned body must not corrupt the stored record.
	rec.ResponseBody[0] = 'X'
	again, _, err := store.Begin(ctx, key, "hash-1", now)
	if err != nil {
		t.Fatalf("Begin after mutation: %v", err)
	}
	if string(again.ResponseBody) != string(body) {
		t.Errorf("stored body mutated to %s, want %s", again.ResponseBody, body)
	}
}

func TestMemoryIdempotencyMismatch(t *testing.T) {
	ctx := context.Background()
	store := db.NewMemoryIdempotency(time.Hour)
	now := time.Unix(1700000000, 0).UTC()
	key := idemKey()

	if _, _, err := store.Begin(ctx, key, "hash-1", now); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := store.Complete(ctx, key, core.IdempotencyRecord{RequestHash: "hash-1", ResponseStatus: 201}, now); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	// Same key, different request hash -> 422 condition.
	if _, _, err := store.Begin(ctx, key, "hash-2", now); !errors.Is(err, core.ErrIdempotencyKeyMismatch) {
		t.Errorf("mismatch Begin = %v, want ErrIdempotencyKeyMismatch", err)
	}
}

func TestMemoryIdempotencyRelease(t *testing.T) {
	ctx := context.Background()
	store := db.NewMemoryIdempotency(time.Hour)
	now := time.Unix(1700000000, 0).UTC()
	key := idemKey()

	if _, _, err := store.Begin(ctx, key, "hash-1", now); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	// Release an in-flight key so it can be retried.
	if err := store.Release(ctx, key); err != nil {
		t.Fatalf("Release: %v", err)
	}
	// A fresh Begin now leases again rather than reporting in-flight.
	_, leased, err := store.Begin(ctx, key, "hash-1", now)
	if err != nil || !leased {
		t.Fatalf("Begin after Release = (leased %v, %v), want leased true", leased, err)
	}
}

func TestMemoryIdempotencyTTL(t *testing.T) {
	ctx := context.Background()
	ttl := time.Minute
	store := db.NewMemoryIdempotency(ttl)
	now := time.Unix(1700000000, 0).UTC()
	key := idemKey()

	if _, _, err := store.Begin(ctx, key, "hash-1", now); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := store.Complete(ctx, key, core.IdempotencyRecord{RequestHash: "hash-1", ResponseStatus: 201}, now); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	// After the TTL lapses the key is reusable: a different-hash request that
	// would otherwise be a mismatch now leases a fresh slot.
	later := now.Add(ttl + time.Second)
	_, leased, err := store.Begin(ctx, key, "hash-2", later)
	if err != nil || !leased {
		t.Fatalf("Begin after TTL = (leased %v, %v), want leased true", leased, err)
	}
}

func TestNewMemoryIdempotencyRejectsNonPositiveTTL(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewMemoryIdempotency(0) did not panic")
		}
	}()
	_ = db.NewMemoryIdempotency(0)
}
