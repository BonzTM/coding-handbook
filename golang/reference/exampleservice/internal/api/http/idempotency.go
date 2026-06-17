package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/exampleservice/internal/core"
)

// idempotencyHeader is the request header carrying the client-supplied key.
const idempotencyHeader = "Idempotency-Key"

// clock is the narrow time seam the idempotency middleware needs so TTL
// stamping is deterministic in tests. *core.Clock-shaped; production wires the
// real clock, tests a fake.
type clock interface {
	Now() time.Time
}

// idempotencyMiddleware implements recipes/add-idempotent-write.md for unsafe
// writes. It keys on (tenant, route, header value), fingerprints the request
// body, and:
//
//   - absent the header, passes through unchanged (idempotency is opt-in);
//   - on first use, processes the request while buffering the response, then
//     persists status+body via the store so a retry replays it;
//   - on a duplicate completed key with the SAME body, replays the stored
//     response byte-identically WITHOUT re-running the handler (no second side
//     effect);
//   - on an in-flight duplicate, returns 409;
//   - on a reused key with a DIFFERENT body, returns 422;
//   - bounded by the store's TTL.
//
// It runs AFTER auth/authz (it needs the principal's tenant) and just before the
// handler, so only authorized writes consume a key. The store write and the
// domain side effect are not in one SQL transaction in the in-memory build; the
// SQL store (behind the integration tag) persists the response in the same
// transaction as the write for true atomicity, per the recipe.
func idempotencyMiddleware(store core.IdempotencyStore, clk clock, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(idempotencyHeader)
			if key == "" {
				// Opt-in: no key means no idempotency tracking.
				next.ServeHTTP(w, r)
				return
			}

			p, ok := core.PrincipalFrom(r.Context())
			if !ok {
				// Should not happen: auth runs first. Fail closed.
				writeError(w, r, logger, core.ErrUnauthenticated)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeError(w, r, logger, core.ErrInvalidWidget)
				return
			}
			// Restore the body so the handler can decode it.
			r.Body = io.NopCloser(bytes.NewReader(body))

			ik := core.IdempotencyKey{TenantID: p.TenantID, Route: routePattern(r), Key: key}
			hash := hashRequest(body)
			now := clk.Now()

			record, leased, err := store.Begin(r.Context(), ik, hash, now)
			if err != nil {
				// In-flight (409) and mismatch (422) map through statusForError.
				writeError(w, r, logger, err)
				return
			}
			if !leased {
				// Completed duplicate with the same body: replay byte-identically.
				replay(w, record)
				return
			}

			// Leased: process the request while capturing the response.
			cw := &captureWriter{ResponseWriter: w, status: http.StatusOK}
			func() {
				// If the handler panics, release the lease so the key is retryable
				// rather than wedged in flight; then re-panic for recoverMiddleware.
				defer func() {
					if rec := recover(); rec != nil {
						releaseQuietly(r.Context(), store, ik, logger)
						panic(rec)
					}
				}()
				next.ServeHTTP(cw, r)
			}()

			// Persist the captured response so retries replay it. The TTL is stamped
			// from the same clock instant used for Begin.
			rec := core.IdempotencyRecord{
				RequestHash:    hash,
				ResponseStatus: cw.status,
				ResponseBody:   cw.body.Bytes(),
			}
			if err := store.Complete(r.Context(), ik, rec, now); err != nil {
				// The response is already written to the client; we cannot change it.
				// Log once so a failed persist (the next retry would re-run the side
				// effect) is observable.
				logger.ErrorContext(r.Context(), "idempotency complete failed",
					"route", ik.Route, "error", err.Error())
			}
		})
	}
}

// captureWriter tees the handler's response into a buffer while writing it
// through to the client, so the exact status and body can be persisted for
// replay. It buffers the full body; idempotent writes are small JSON responses,
// not streams, so this is bounded.
type captureWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func (c *captureWriter) WriteHeader(code int) {
	if c.wroteHeader {
		return
	}
	c.status = code
	c.wroteHeader = true
	c.ResponseWriter.WriteHeader(code)
}

func (c *captureWriter) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	c.body.Write(b)
	n, err := c.ResponseWriter.Write(b)
	if err != nil {
		return n, err //nolint:wrapcheck // pass the writer's error through unchanged
	}
	return n, nil
}

// replay writes a stored response back to the client byte-identically. The
// Content-Type is restored to JSON to match the original write path.
func replay(w http.ResponseWriter, rec *core.IdempotencyRecord) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Idempotency-Replayed", "true")
	w.WriteHeader(rec.ResponseStatus)
	_, _ = w.Write(rec.ResponseBody) //nolint:errcheck // status committed; client cannot be told of a late write error
}

// releaseQuietly releases a leased key, logging a failure rather than masking
// the original panic that triggered it.
func releaseQuietly(ctx context.Context, store core.IdempotencyStore, ik core.IdempotencyKey, logger *slog.Logger) {
	if err := store.Release(ctx, ik); err != nil {
		logger.ErrorContext(ctx, "idempotency release failed",
			"route", ik.Route, "error", err.Error())
	}
}

// hashRequest fingerprints the request body so a reused key with a different
// body is detected (422). SHA-256 hex is stable and collision-resistant.
func hashRequest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
