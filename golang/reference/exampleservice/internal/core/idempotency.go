package core

import (
	"context"
	"errors"
	"time"
)

// Idempotency sentinel errors the transport boundary branches on with
// errors.Is. They implement recipes/add-idempotent-write.md: a duplicate
// in-flight key is a 409, and a reused key with a different request body is a
// 422.
var (
	// ErrIdempotencyInFlight is returned when a request reuses an Idempotency-Key
	// whose original request is still being processed (no stored response yet).
	// The transport maps it to 409 Conflict so the client retries later.
	ErrIdempotencyInFlight = errors.New("idempotency key in flight")
	// ErrIdempotencyKeyMismatch is returned when a request reuses a completed
	// Idempotency-Key but with a different request fingerprint (body/route). The
	// transport maps it to 422 Unprocessable Entity: the key is being misused.
	ErrIdempotencyKeyMismatch = errors.New("idempotency key reused with a different request")
)

// IdempotencyKey identifies a stored idempotent operation. It is scoped to the
// tenant and route so the same client-supplied key under a different tenant or
// endpoint is a distinct operation — keys are never global.
type IdempotencyKey struct {
	// TenantID scopes the key to one tenant.
	TenantID string
	// Route is the matched low-cardinality route pattern, e.g. "POST /widgets".
	Route string
	// Key is the client-supplied Idempotency-Key header value.
	Key string
}

// IdempotencyRecord is the persisted outcome of a completed idempotent request:
// the request fingerprint that produced it plus the exact response to replay.
// Storing the response (status + body) lets a duplicate completed key replay
// byte-identically without re-running the side effect.
type IdempotencyRecord struct {
	// RequestHash is a fingerprint of the original request body. A later request
	// with the same key but a different hash is a misuse (ErrIdempotencyKeyMismatch).
	RequestHash string
	// ResponseStatus is the HTTP status the original request returned.
	ResponseStatus int
	// ResponseBody is the exact response body bytes to replay.
	ResponseBody []byte
}

// IdempotencyStore persists the result of idempotent writes so a retried request
// replays the original response instead of repeating the side effect. It is the
// consumer-defined contract from recipes/add-idempotent-write.md; the in-memory
// implementation lives in internal/db and the SQL implementation behind the
// integration build tag. The store owns TTL expiry so abandoned in-flight keys
// do not wedge a slot forever.
type IdempotencyStore interface {
	// Begin reserves key for an in-flight request fingerprinted by requestHash.
	//
	//   - If key is unused (or its TTL has lapsed), it reserves the slot and
	//     returns leased=true with a nil record: the caller processes the request
	//     and MUST call Complete or Release.
	//   - If key is reserved but not yet completed, it returns
	//     ErrIdempotencyInFlight (a concurrent duplicate).
	//   - If key is completed with the SAME requestHash, it returns leased=false
	//     and the stored record to replay.
	//   - If key is completed with a DIFFERENT requestHash, it returns
	//     ErrIdempotencyKeyMismatch.
	Begin(ctx context.Context, key IdempotencyKey, requestHash string, now time.Time) (record *IdempotencyRecord, leased bool, err error)
	// Complete records the final response for a leased key so future duplicates
	// replay it. now stamps the record for TTL expiry.
	Complete(ctx context.Context, key IdempotencyKey, rec IdempotencyRecord, now time.Time) error
	// Release abandons a leased key without recording a response (e.g. the handler
	// panicked or the request was canceled before completion) so the key can be
	// retried rather than staying wedged in flight.
	Release(ctx context.Context, key IdempotencyKey) error
}
