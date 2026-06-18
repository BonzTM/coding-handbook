package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
	"github.com/example/exampleservice/internal/telemetry"
)

// TestWidgetResponseGoldenJSON pins the EXACT wire shape of a widget response:
// the keys, their snake_case spelling, and the order json.Marshal emits them in
// (struct-field order). A struct-tag rename, a case change, or a field
// add/remove is then caught as a byte-level diff, per the golden-test
// requirement in golang/foundations/serialization.md (Verification And Proof).
func TestWidgetResponseGoldenJSON(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	// The test server uses a fixed clock at unix 1700000000 (2023-11-14T22:13:20Z).
	req := newCreateRequest(`{"id":"w1","name":"Widget One"}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /widgets status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	// json.NewEncoder appends a trailing newline; trim it so the golden is the
	// exact JSON object bytes.
	got := strings.TrimRight(rec.Body.String(), "\n")
	const golden = `{"id":"w1","name":"Widget One","created_at":"2023-11-14T22:13:20Z"}`
	if got != golden {
		t.Errorf("widget response JSON mismatch:\n got: %s\nwant: %s", got, golden)
	}
}

// TestListWidgetsGoldenJSON pins the keyset-pagination envelope: a top-level
// object with an "items" array and an opaque "next_cursor" (never a bare
// array), so the response stays extensible and a wrapper or cursor-field rename
// is caught as a diff. A single seeded widget fits in one page, so next_cursor
// is the empty string.
func TestListWidgetsGoldenJSON(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	createReq := newCreateRequest(`{"id":"w1","name":"Widget One"}`)
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("seed create status = %d, want 201", createRec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /widgets status = %d, want 200", rec.Code)
	}

	got := strings.TrimRight(rec.Body.String(), "\n")
	const golden = `{"items":[{"id":"w1","name":"Widget One","created_at":"2023-11-14T22:13:20Z"}],"next_cursor":""}`
	if got != golden {
		t.Errorf("list response JSON mismatch:\n got: %s\nwant: %s", got, golden)
	}
}

// newCappedServer builds a Server with a tiny request-body cap so the
// MaxBytesReader limit can be exercised with a small payload.
func newCappedServer(t *testing.T, maxBodyBytes int64) *Server {
	t.Helper()
	svc := core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.HTTPConfig{
		Addr:              ":0",
		ReadHeaderTimeout: time.Second,
		MaxBodyBytes:      maxBodyBytes,
	}
	return New(cfg, svc, logger, telemetry.NopMetrics{}, telemetry.NewReadiness(true), testDeps())
}

// TestCreateWidgetBodyTooLarge proves the body-size cap: a request body larger
// than MaxBytesReader's limit is rejected with 400 (a client error), not an OOM
// or 500, per the body-size-cap requirement in
// golang/foundations/serialization.md (Verification And Proof). MaxBytesReader
// turns the over-limit read into a decode error, which the handler maps to 400.
func TestCreateWidgetBodyTooLarge(t *testing.T) {
	const limit = 64 // bytes
	srv := newCappedServer(t, limit)
	h := srv.Handler()

	// A valid-looking JSON object whose "name" alone exceeds the cap. The
	// Idempotency-Key is required on the create route and is checked BEFORE the
	// body is read, so set one here to reach the body-size-cap path under test.
	oversized := `{"id":"w1","name":"` + strings.Repeat("x", limit*4) + `"}`
	req := newCreateRequest(oversized)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized body status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

// TestErrorEnvelopeGoldenJSON pins the structured error envelope on a validation
// failure: the exact keys, the top-level "validation_failed" code, and the
// per-field "fields" shape (field path + per-field code), so a thinning or
// rename of the error contract shows up as a diff, per
// golang/foundations/serialization.md (Verification And Proof). The request_id
// is non-deterministic, so it is replaced with a stable placeholder before the
// byte comparison; its presence is asserted separately.
func TestErrorEnvelopeGoldenJSON(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	// Missing "name" triggers a field-level validation failure.
	req := newCreateRequest(`{"id":"w1"}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	// Pull the (non-deterministic) request_id out and assert it is present, then
	// pin the remaining envelope byte-for-byte.
	var parsed ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if parsed.RequestID == "" {
		t.Fatal("request_id missing from error envelope")
	}
	normalized := strings.Replace(strings.TrimRight(rec.Body.String(), "\n"), parsed.RequestID, "REQID", 1)
	const golden = `{"code":"validation_failed","message":"the request has invalid fields","fields":[{"field":"name","code":"required","message":"name must not be empty"}],"request_id":"REQID"}`
	if normalized != golden {
		t.Errorf("error envelope JSON mismatch:\n got: %s\nwant: %s", normalized, golden)
	}
}

// TestServerErrorIsOpaque proves a 5xx body carries NO internal detail: a
// generic "internal" code and message, a request_id for correlation, and none
// of the panic text, per golang/foundations/serialization.md (a 5xx is opaque;
// detail lives in the log under the request_id). It drives the recovery
// middleware directly with a handler that panics with a secret string.
func TestServerErrorIsOpaque(t *testing.T) {
	const secret = "super-secret-internal-detail"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	boom := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(secret)
	})
	// Mirror the server's wiring: request ID is OUTERMOST so the id is on the
	// context before recovery runs and the recovered 500 can echo it.
	h := requestIDMiddleware(recoverMiddleware(logger)(boom))

	req := httptest.NewRequest(http.MethodGet, "/widgets/x", nil).WithContext(context.Background())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), secret) {
		t.Errorf("5xx body leaks internal detail: %s", rec.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode 5xx envelope: %v", err)
	}
	if errResp.Code != codeInternal {
		t.Errorf("code = %q, want %q", errResp.Code, codeInternal)
	}
	if errResp.Message != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("message = %q, want generic %q", errResp.Message, http.StatusText(http.StatusInternalServerError))
	}
	if errResp.RequestID == "" {
		t.Error("request_id missing from 5xx envelope")
	}
	if len(errResp.Fields) != 0 {
		t.Errorf("5xx must not carry field errors, got %+v", errResp.Fields)
	}
}
