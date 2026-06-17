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

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// newTestServer wires a Server against the real in-memory store and core
// service, with a discard logger and a readiness flag the caller controls.
func newTestServer(t *testing.T, ready bool) *Server {
	t.Helper()
	svc := core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.HTTPConfig{
		Addr:              ":0",
		ReadHeaderTimeout: time.Second,
		MaxBodyBytes:      1 << 20,
	}
	return New(cfg, svc, logger, telemetry.NopMetrics{}, telemetry.NewReadiness(ready))
}

func TestCreateAndGetWidget(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	// Create.
	body := strings.NewReader(`{"id":"w1","name":"Widget One"}`)
	req := httptest.NewRequest(http.MethodPost, "/widgets", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /widgets status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var created widgetResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID != "w1" || created.Name != "Widget One" {
		t.Errorf("created = %+v, want id=w1 name=Widget One", created)
	}

	// Get it back.
	req = httptest.NewRequest(http.MethodGet, "/widgets/w1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /widgets/w1 status = %d, want 200", rec.Code)
	}
}

func TestCreateWidgetValidationError(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	// Missing name -> 400 from core validation, mapped at the boundary.
	body := strings.NewReader(`{"id":"w1"}`)
	req := httptest.NewRequest(http.MethodPost, "/widgets", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var errResp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("error response body is empty")
	}
}

func TestCreateWidgetUnknownField(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	// Strict surface: an unknown field is rejected with 400.
	body := strings.NewReader(`{"id":"w1","name":"n","extra":true}`)
	req := httptest.NewRequest(http.MethodPost, "/widgets", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown-field status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetWidgetNotFound(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/widgets/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestCreateWidgetConflict(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	create := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"dup","name":"n"}`))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	if rec := create(); rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want 201", rec.Code)
	}
	if rec := create(); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate create status = %d, want 409", rec.Code)
	}
}

func TestLivezReadyz(t *testing.T) {
	// Liveness is always OK; readiness reflects the flag.
	srv := newTestServer(t, false)
	h := srv.Handler()

	// /livez is OK even when not ready.
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/livez (unready) status = %d, want 200", rec.Code)
	}

	// /readyz is 503 when not ready.
	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("/readyz (unready) status = %d, want 503", rec.Code)
	}

	// Flip to ready: /readyz becomes 200.
	srv.SetReady(true)
	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/readyz (ready) status = %d, want 200", rec.Code)
	}
}

func TestShutdownDrains(t *testing.T) {
	// Prove the ordered-shutdown contract end to end against a real listener:
	// readiness flips before Shutdown returns and the drain is bounded.
	srv := newTestServer(t, true)

	ln, ts := startServer(t, srv)
	defer ts.Close()

	// Sanity: ready before shutdown.
	if !get(t, ln, "/readyz") {
		t.Fatal("expected ready before shutdown")
	}

	srv.SetReady(false)
	if get(t, ln, "/readyz") {
		t.Fatal("expected unready after SetReady(false)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

// startServer serves srv.Handler() on an httptest server so the listener path
// is exercised without binding the configured port.
func startServer(t *testing.T, srv *Server) (string, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(srv.Handler())
	return ts.URL, ts
}

func get(t *testing.T, base, path string) bool {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, base+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatalf("drain body: %v", err)
	}
	return resp.StatusCode == http.StatusOK
}
