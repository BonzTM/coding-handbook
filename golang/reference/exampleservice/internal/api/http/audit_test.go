package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/telemetry"
)

// auditServer wires a Server with auth enabled and a capturing audit sink so
// tests can assert the audit events emitted by the authn, authz, and write
// paths. The buffer is the dedicated audit stream, separate from the (discarded)
// application logger.
func auditServer(t *testing.T) (http.Handler, *auditSink, *tokenSigner) {
	t.Helper()
	signer := newTokenSigner(t)
	sink := &auditSink{}
	clock := fixedClock{t: time.Unix(1700000000, 0).UTC()}
	srv := newTestServerWithDeps(t, true, Deps{
		Verifier:    signer.verifier(),
		Idempotency: memIdem(),
		Clock:       clock,
		Audit:       telemetry.NewAuditLogger(sink, clock),
	})
	return srv.Handler(), sink, signer
}

// auditSink is a concurrency-safe io.Writer that records each JSON audit line.
type auditSink struct {
	mu    sync.Mutex
	lines [][]byte
}

func (s *auditSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// slog writes one record per Write call; copy because slog reuses its buffer.
	cp := make([]byte, len(p))
	copy(cp, p)
	s.lines = append(s.lines, cp)
	return len(p), nil
}

// records decodes every captured audit line into a generic map.
func (s *auditSink) records(t *testing.T) []map[string]any {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]map[string]any, 0, len(s.lines))
	for _, ln := range s.lines {
		var m map[string]any
		if err := json.Unmarshal(ln, &m); err != nil {
			t.Fatalf("decode audit record %q: %v", ln, err)
		}
		out = append(out, m)
	}
	return out
}

// only returns the single audit record, failing if there is not exactly one.
func (s *auditSink) only(t *testing.T) map[string]any {
	t.Helper()
	recs := s.records(t)
	if len(recs) != 1 {
		t.Fatalf("audit records = %d, want exactly 1: %v", len(recs), recs)
	}
	return recs[0]
}

func wantField(t *testing.T, rec map[string]any, key, want string) {
	t.Helper()
	got, ok := rec[key].(string)
	if !ok {
		t.Fatalf("audit record missing string field %q: %v", key, rec)
	}
	if got != want {
		t.Errorf("audit %q = %q, want %q", key, got, want)
	}
}

// TestAuditAuthnFailureEmitsEvent proves an authentication failure (invalid
// token) emits an audit record with action=authenticate, result=failure, the
// route as resource, and NO actor/tenant (no principal was established) — and
// that the raw token never appears in the audit stream.
func TestAuditAuthnFailureEmitsEvent(t *testing.T) {
	h, sink, _ := auditServer(t)

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	r := sink.only(t)
	wantField(t, r, "action", auditActionAuthenticate)
	wantField(t, r, "result", string(telemetry.AuditFailure))
	// The authn check runs before route matching, so resource is empty (the raw
	// path is not recorded) and there is no principal to attribute.
	wantField(t, r, "resource", "")
	wantField(t, r, "actor", "")
	wantField(t, r, "tenant", "")
	if r["request_id"] == "" {
		t.Errorf("audit record missing request_id: %v", r)
	}
	if _, ok := r["time"].(string); !ok {
		t.Errorf("audit record missing time: %v", r)
	}
	// The credential must never reach the audit stream.
	for _, ln := range sink.lines {
		if bytes.Contains(ln, []byte("not-a-real-token")) {
			t.Fatalf("audit stream leaked the bearer token: %s", ln)
		}
	}
}

// TestAuditMissingTokenEmitsAuthnFailure proves a missing Authorization header
// (no token at all) is audited as an authentication failure too.
func TestAuditMissingTokenEmitsAuthnFailure(t *testing.T) {
	h, sink, _ := auditServer(t)

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	r := sink.only(t)
	wantField(t, r, "action", auditActionAuthenticate)
	wantField(t, r, "result", string(telemetry.AuditFailure))
}

// TestAuditAuthzDenialEmitsEvent proves an authorization denial (a reader trying
// to write) emits action=authorize, result=denied, WITH the known actor and
// tenant, and the route as resource.
func TestAuditAuthzDenialEmitsEvent(t *testing.T) {
	h, sink, signer := auditServer(t)

	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"n"}`))
	req.Header.Set("Authorization", "Bearer "+signer.token(t, "tenant-1", []string{"widgets.reader"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	r := sink.only(t)
	wantField(t, r, "action", auditActionAuthorize)
	wantField(t, r, "result", string(telemetry.AuditDenied))
	wantField(t, r, "resource", "POST /widgets")
	wantField(t, r, "actor", "user-1")
	wantField(t, r, "tenant", "tenant-1")
}

// TestAuditWidgetCreateEmitsEvent proves a successful data-mutating write emits
// action=widget.create, result=success, with the actor/tenant and the created
// resource id — and that the widget NAME (application payload) never appears in
// the audit record.
func TestAuditWidgetCreateEmitsEvent(t *testing.T) {
	h, sink, signer := auditServer(t)

	req := newCreateRequest(`{"id":"w1","name":"secret-widget-name"}`)
	req.Header.Set("Authorization", "Bearer "+signer.token(t, "tenant-1", []string{"widgets.writer"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	r := sink.only(t)
	wantField(t, r, "action", auditActionWidgetCreate)
	wantField(t, r, "result", string(telemetry.AuditSuccess))
	wantField(t, r, "resource", "w1")
	wantField(t, r, "actor", "user-1")
	wantField(t, r, "tenant", "tenant-1")
	// The application payload (the widget name) must not be audited.
	for _, ln := range sink.lines {
		if bytes.Contains(ln, []byte("secret-widget-name")) {
			t.Fatalf("audit stream leaked the widget name payload: %s", ln)
		}
	}
}
