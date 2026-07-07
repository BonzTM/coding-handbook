package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

// TestIdempotencySameKeyReplaysAndRunsOnce drives the create endpoint over the
// wire twice with the same Idempotency-Key and identical body, then proves the
// side effect ran ONCE (one widget) and the replayed response is byte-identical
// to the first (the recipe's core guarantee).
func TestIdempotencySameKeyReplaysAndRunsOnce(t *testing.T) {
	srv := newTestServer(t, true) // auth disabled -> local-dev principal (writer)
	h := srv.Handler()

	post := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"n"}`))
		req.Header.Set(idempotencyHeader, "key-1")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	first := post()
	if first.Code != http.StatusCreated {
		t.Fatalf("first POST status = %d, want 201; body=%s", first.Code, first.Body.String())
	}
	second := post()
	if second.Code != http.StatusCreated {
		t.Fatalf("replay POST status = %d, want 201 (replayed, not 409 conflict); body=%s", second.Code, second.Body.String())
	}
	if first.Body.String() != second.Body.String() {
		t.Errorf("replay body differs:\n first=%s\nsecond=%s", first.Body.String(), second.Body.String())
	}
	if second.Header().Get("Idempotency-Replayed") != "true" {
		t.Errorf("replay missing Idempotency-Replayed header")
	}

	// Exactly one widget exists: the side effect ran once.
	list := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	lrec := httptest.NewRecorder()
	h.ServeHTTP(lrec, list)
	if n := strings.Count(lrec.Body.String(), `"id":"w1"`); n != 1 {
		t.Errorf("widget count = %d, want 1 (single side effect); body=%s", n, lrec.Body.String())
	}
}

// TestIdempotencyDifferentBodyIs422 proves a reused completed key with a
// different request body is rejected with 422.
func TestIdempotencyDifferentBodyIs422(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	first := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"n"}`))
	first.Header.Set(idempotencyHeader, "key-1")
	frec := httptest.NewRecorder()
	h.ServeHTTP(frec, first)
	if frec.Code != http.StatusCreated {
		t.Fatalf("first POST status = %d, want 201", frec.Code)
	}

	// Same key, different body -> 422.
	second := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w2","name":"other"}`))
	second.Header.Set(idempotencyHeader, "key-1")
	srec := httptest.NewRecorder()
	h.ServeHTTP(srec, second)
	if srec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("different-body status = %d, want 422; body=%s", srec.Code, srec.Body.String())
	}
}

// TestIdempotencyInFlightIs409 drives the middleware directly with a handler
// that blocks until signaled, so two concurrent requests with the same key race:
// the first holds the lease while the second observes it in flight and gets 409.
func TestIdempotencyInFlightIs409(t *testing.T) {
	store := db.NewMemoryIdempotency(time.Hour)
	logger := slog.New(slog.DiscardHandler)
	clk := fixedClock{t: time.Unix(1700000000, 0).UTC()}

	release := make(chan struct{})
	started := make(chan struct{}, 1)
	var calls atomic.Int32
	blocking := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		started <- struct{}{}
		<-release // hold the lease open
		w.WriteHeader(http.StatusCreated)
		if _, werr := io.WriteString(w, `{"ok":true}`); werr != nil {
			t.Errorf("write body: %v", werr)
		}
	})
	mw := idempotencyMiddleware(store, clk, logger)(blocking)

	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1"}`))
		req.Header.Set(idempotencyHeader, "key-1")
		// The middleware needs a principal (auth runs first in the real chain).
		ctx := core.WithPrincipal(req.Context(), core.Principal{Subject: "s", TenantID: "t1"})
		return req.WithContext(ctx)
	}

	// Fire the first request; it blocks inside the handler holding the lease.
	rec1 := httptest.NewRecorder()
	var wg sync.WaitGroup
	wg.Go(func() {
		mw.ServeHTTP(rec1, newReq())
	})
	<-started // ensure the lease is held before the duplicate runs

	// Second request with the same key while the first is in flight -> 409.
	rec2 := httptest.NewRecorder()
	mw.ServeHTTP(rec2, newReq())
	if rec2.Code != http.StatusConflict {
		t.Errorf("in-flight duplicate status = %d, want 409", rec2.Code)
	}

	close(release)
	wg.Wait()

	if rec1.Code != http.StatusCreated {
		t.Errorf("first request status = %d, want 201", rec1.Code)
	}
	if calls.Load() != 1 {
		t.Errorf("handler ran %d times, want 1 (in-flight duplicate must not run the handler)", calls.Load())
	}
}

// TestIdempotencyMissingKeyIs400 proves the create route REQUIRES an
// Idempotency-Key: a POST without the header is rejected with 400 before the
// handler runs (no side effect), per recipes/add-idempotent-write.md. The
// rejection uses the structured envelope with a machine-readable code and the
// request_id in the body, not a bare string.
func TestIdempotencyMissingKeyIs400(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"n"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing-key POST = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Code != codeInvalidArgument {
		t.Errorf("code = %q, want %q", errResp.Code, codeInvalidArgument)
	}
	if errResp.RequestID == "" {
		t.Error("request_id missing from error body")
	}

	// No widget was created: the missing-key request never reached the handler.
	list := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	lrec := httptest.NewRecorder()
	h.ServeHTTP(lrec, list)
	if strings.Contains(lrec.Body.String(), `"id":"w1"`) {
		t.Errorf("widget was created despite missing key; body=%s", lrec.Body.String())
	}
}
