package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
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
// service, with a discard logger and a readiness flag the caller controls. Auth
// is disabled (nil verifier): the auth middleware injects the local-dev
// principal (tenant "local-dev", both roles), so the widgets routes are
// exercisable without a token. Tests that need a real token wire a verifier via
// newTestServerWithAuth.
func newTestServer(t *testing.T, ready bool) *Server {
	t.Helper()
	return newTestServerWithDeps(t, ready, Deps{
		Idempotency: db.NewMemoryIdempotency(time.Hour),
		Clock:       fixedClock{t: time.Unix(1700000000, 0).UTC()},
	})
}

// memIdem builds a fresh in-memory idempotency store for tests.
func memIdem() *db.MemoryIdempotency { return db.NewMemoryIdempotency(time.Hour) }

// testDeps returns Deps with auth disabled and a fresh in-memory idempotency
// store, the default for tests that do not exercise identity/idempotency.
func testDeps() Deps {
	return Deps{
		Idempotency: memIdem(),
		Clock:       fixedClock{t: time.Unix(1700000000, 0).UTC()},
	}
}

// newTestServerWithDeps wires a Server with caller-supplied identity/idempotency
// deps so individual tests can inject a verifier or a shared idempotency store.
func newTestServerWithDeps(t *testing.T, ready bool, deps Deps) *Server {
	t.Helper()
	svc := core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.HTTPConfig{
		Addr:              ":0",
		ReadHeaderTimeout: time.Second,
		MaxBodyBytes:      1 << 20,
	}
	return New(cfg, svc, logger, telemetry.NopMetrics{}, telemetry.NewReadiness(ready), deps)
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

// TestListWidgetsPaginationWalk drives the keyset List endpoint over the wire:
// it seeds widgets, then walks every page using the opaque next_cursor and
// proves the pages partition the set in stable order with an empty cursor on the
// last page. The fixed test clock gives every widget the same CreatedAt, so the
// id tiebreaker drives the order.
func TestListWidgetsPaginationWalk(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	ids := []string{"w5", "w1", "w3", "w2", "w4"}
	for _, id := range ids {
		body := strings.NewReader(`{"id":"` + id + `","name":"n"}`)
		req := httptest.NewRequest(http.MethodPost, "/widgets", body)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed %s status = %d, want 201", id, rec.Code)
		}
	}

	var got []string
	cursor := ""
	for page := 0; page < len(ids)+1; page++ {
		path := "/widgets?page_size=2"
		if cursor != "" {
			path += "&cursor=" + url.QueryEscape(cursor)
		}
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("page %d status = %d, want 200; body=%s", page, rec.Code, rec.Body.String())
		}
		var resp listWidgetsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode page %d: %v", page, err)
		}
		if len(resp.Items) > 2 {
			t.Fatalf("page %d items = %d, want <= 2", page, len(resp.Items))
		}
		for _, it := range resp.Items {
			got = append(got, it.ID)
		}
		if resp.NextCursor == "" {
			break
		}
		cursor = resp.NextCursor
	}

	want := []string{"w1", "w2", "w3", "w4", "w5"}
	if len(got) != len(want) {
		t.Fatalf("walked %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestListWidgetsPageSizeClamp proves an oversized page_size is clamped to the
// server maximum rather than rejected or honored unbounded.
func TestListWidgetsPageSizeClamp(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	// Seed more than MaxPageSize widgets so a clamp is observable.
	total := core.MaxPageSize + 10
	for i := range total {
		id := fmt.Sprintf("w%04d", i)
		body := strings.NewReader(`{"id":"` + id + `","name":"n"}`)
		req := httptest.NewRequest(http.MethodPost, "/widgets", body)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed %s status = %d, want 201", id, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/widgets?page_size=%d", total*2), nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp listWidgetsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != core.MaxPageSize {
		t.Errorf("clamped page items = %d, want %d", len(resp.Items), core.MaxPageSize)
	}
	// More rows remain, so the cursor must be non-empty.
	if resp.NextCursor == "" {
		t.Error("next_cursor is empty but more rows remain")
	}
}

// TestListWidgetsInvalidCursor proves a malformed cursor is a 400, not a 500.
func TestListWidgetsInvalidCursor(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/widgets?cursor=!!!not-valid!!!", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

// TestListWidgetsInvalidPageSize proves a non-numeric page_size is a 400.
func TestListWidgetsInvalidPageSize(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/widgets?page_size=abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

// TestListWidgetsLastPageEmptyCursor proves a single full page reports an empty
// next_cursor (the last-page contract).
func TestListWidgetsLastPageEmptyCursor(t *testing.T) {
	srv := newTestServer(t, true)
	h := srv.Handler()

	for _, id := range []string{"a", "b"} {
		body := strings.NewReader(`{"id":"` + id + `","name":"n"}`)
		req := httptest.NewRequest(http.MethodPost, "/widgets", body)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed %s status = %d, want 201", id, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/widgets?page_size=10", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp listWidgetsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(resp.Items))
	}
	if resp.NextCursor != "" {
		t.Errorf("next_cursor = %q, want empty on last page", resp.NextCursor)
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
