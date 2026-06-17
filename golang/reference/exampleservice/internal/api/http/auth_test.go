package http

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/exampleservice/internal/auth"
)

// tokenSigner signs RS256 JWTs with a local key for the HTTP auth tests and
// exposes a matching StaticVerifier — no network, no real JWKS.
type tokenSigner struct {
	key *rsa.PrivateKey
}

func newTokenSigner(t *testing.T) *tokenSigner {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return &tokenSigner{key: key}
}

func (s *tokenSigner) verifier() auth.Verifier {
	kf := func(*jwt.Token) (any, error) { return &s.key.PublicKey, nil }
	return auth.NewStaticVerifier(kf, "https://idp.test", "exampleservice")
}

func (s *tokenSigner) token(t *testing.T, tenant string, roles []string) string {
	t.Helper()
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://idp.test",
			Subject:   "user-1",
			Audience:  jwt.ClaimStrings{"exampleservice"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		TenantID: tenant,
		Roles:    roles,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	raw, err := tok.SignedString(s.key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return raw
}

func (s *tokenSigner) authServer(t *testing.T) http.Handler {
	t.Helper()
	srv := newTestServerWithDeps(t, true, Deps{
		Verifier:    s.verifier(),
		Idempotency: memIdem(),
		Clock:       fixedClock{t: time.Unix(1700000000, 0).UTC()},
	})
	return srv.Handler()
}

func TestAuthMissingTokenIs401(t *testing.T) {
	h := newTokenSigner(t).authServer(t)

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthInvalidTokenIs401(t *testing.T) {
	h := newTokenSigner(t).authServer(t)

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad token status = %d, want 401", rec.Code)
	}
	// The 401 body must not leak which check failed.
	if strings.Contains(rec.Body.String(), "tenant") || strings.Contains(rec.Body.String(), "claim") {
		t.Errorf("401 body leaks detail: %s", rec.Body.String())
	}
}

func TestAuthValidTokenAllowsRead(t *testing.T) {
	s := newTokenSigner(t)
	h := s.authServer(t)

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	req.Header.Set("Authorization", "Bearer "+s.token(t, "tenant-1", []string{"widgets.reader"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid read status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthzReaderCannotWriteIs403(t *testing.T) {
	s := newTokenSigner(t)
	h := s.authServer(t)

	// A reader-only token may not POST: 403 from the route-scoped role check.
	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"n"}`))
	req.Header.Set("Authorization", "Bearer "+s.token(t, "tenant-1", []string{"widgets.reader"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("reader POST status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthzWriterCanWrite(t *testing.T) {
	s := newTokenSigner(t)
	h := s.authServer(t)

	req := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"n"}`))
	req.Header.Set("Authorization", "Bearer "+s.token(t, "tenant-1", []string{"widgets.writer"}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("writer POST status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
}

// TestAuthzCrossTenantReadIsNotFound proves the per-resource ownership check:
// tenant A creates a widget; tenant B (same roles) cannot read it — it is a 404,
// never tenant A's data.
func TestAuthzCrossTenantReadIsNotFound(t *testing.T) {
	s := newTokenSigner(t)
	h := s.authServer(t)

	// Tenant A creates w1.
	create := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"id":"w1","name":"a-widget"}`))
	create.Header.Set("Authorization", "Bearer "+s.token(t, "tenant-a", []string{"widgets.writer"}))
	crec := httptest.NewRecorder()
	h.ServeHTTP(crec, create)
	if crec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", crec.Code)
	}

	// Tenant B reads w1 -> 404 (cross-tenant isolation in the store).
	read := httptest.NewRequest(http.MethodGet, "/widgets/w1", nil)
	read.Header.Set("Authorization", "Bearer "+s.token(t, "tenant-b", []string{"widgets.reader"}))
	rrec := httptest.NewRecorder()
	h.ServeHTTP(rrec, read)
	if rrec.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant read status = %d, want 404", rrec.Code)
	}
}
