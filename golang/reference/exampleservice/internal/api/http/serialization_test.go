package http

import (
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
	body := strings.NewReader(`{"id":"w1","name":"Widget One"}`)
	req := httptest.NewRequest(http.MethodPost, "/widgets", body)
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

// TestListWidgetsGoldenJSON pins the collection wrapper shape: a top-level
// object with a "widgets" array (never a bare array), so the response stays
// extensible and a wrapper rename is caught as a diff.
func TestListWidgetsGoldenJSON(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	create := strings.NewReader(`{"id":"w1","name":"Widget One"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/widgets", create)
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
	const golden = `{"widgets":[{"id":"w1","name":"Widget One","created_at":"2023-11-14T22:13:20Z"}]}`
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
	return New(cfg, svc, logger, telemetry.NopMetrics{}, telemetry.NewReadiness(true))
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

	// A valid-looking JSON object whose "name" alone exceeds the cap.
	oversized := `{"id":"w1","name":"` + strings.Repeat("x", limit*4) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(oversized))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized body status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}
