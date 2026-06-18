package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/example/exampleservice/internal/auth"
	"github.com/example/exampleservice/internal/core"
)

const (
	testIssuer   = "https://idp.test"
	testAudience = "exampleservice"
	testKID      = "test-key-1"
)

// signer holds a local RSA key pair for signing test tokens offline. No network
// or real JWKS endpoint is touched.
type signer struct {
	key *rsa.PrivateKey
}

func newSigner(t *testing.T) *signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return &signer{key: key}
}

// sign builds and signs a token with the given claims using RS256 and the test
// key id, matching what an identity provider would issue.
func (s *signer) sign(t *testing.T, claims auth.Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = testKID
	raw, err := tok.SignedString(s.key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return raw
}

// validClaims returns a claim set that passes every check: correct iss/aud, a
// future expiry, a subject, a tenant, and both roles.
func validClaims() auth.Claims {
	now := time.Now()
	return auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuer,
			Subject:   "user-1",
			Audience:  jwt.ClaimStrings{testAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		TenantID: "tenant-1",
		Roles:    []string{"widgets.reader", "widgets.writer"},
	}
}

// staticVerifier wires a StaticVerifier whose keyfunc returns the signer's
// public key, the simplest offline production-path exerciser.
func (s *signer) staticVerifier() auth.Verifier {
	kf := func(*jwt.Token) (any, error) { return &s.key.PublicKey, nil }
	return auth.NewStaticVerifier(kf, testIssuer, testAudience)
}

// jwksVerifier wires a JWKSVerifier backed by in-memory JWKS storage holding the
// signer's public key, so the production JWKS parse path runs with no network.
func (s *signer) jwksVerifier(t *testing.T) auth.Verifier {
	t.Helper()
	ctx := context.Background()
	store := jwkset.NewMemoryStorage()
	jwk, err := jwkset.NewJWKFromKey(s.key, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{KID: testKID},
	})
	if err != nil {
		t.Fatalf("new jwk: %v", err)
	}
	if writeErr := store.KeyWrite(ctx, jwk); writeErr != nil {
		t.Fatalf("key write: %v", writeErr)
	}
	kf, err := keyfunc.New(keyfunc.Options{Storage: store})
	if err != nil {
		t.Fatalf("new keyfunc: %v", err)
	}
	return auth.NewJWKSVerifierFromKeyfunc(kf, testIssuer, testAudience)
}

func TestVerifyValidToken(t *testing.T) {
	s := newSigner(t)
	verifiers := map[string]auth.Verifier{
		"static": s.staticVerifier(),
		"jwks":   s.jwksVerifier(t),
	}
	for name, v := range verifiers {
		t.Run(name, func(t *testing.T) {
			raw := s.sign(t, validClaims())
			p, err := v.Verify(context.Background(), raw)
			if err != nil {
				t.Fatalf("Verify: %v", err)
			}
			if p.Subject != "user-1" || p.TenantID != "tenant-1" {
				t.Errorf("principal = %+v, want sub user-1 tenant tenant-1", p)
			}
			if !p.HasRole(core.RoleReader) || !p.HasRole(core.RoleWriter) {
				t.Errorf("principal roles = %v, want reader+writer", p.Roles)
			}
		})
	}
}

func TestVerifyRejects(t *testing.T) {
	s := newSigner(t)
	v := s.staticVerifier()

	tests := []struct {
		name  string
		token func() string
	}{
		{"empty", func() string { return "" }},
		{"garbage", func() string { return "not.a.jwt" }},
		{"expired", func() string {
			c := validClaims()
			c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
			return s.sign(t, c)
		}},
		{"wrong issuer", func() string {
			c := validClaims()
			c.Issuer = "https://evil.test"
			return s.sign(t, c)
		}},
		{"wrong audience", func() string {
			c := validClaims()
			c.Audience = jwt.ClaimStrings{"other-service"}
			return s.sign(t, c)
		}},
		{"missing expiry", func() string {
			c := validClaims()
			c.ExpiresAt = nil
			return s.sign(t, c)
		}},
		{"missing subject", func() string {
			c := validClaims()
			c.Subject = ""
			return s.sign(t, c)
		}},
		{"missing tenant", func() string {
			c := validClaims()
			c.TenantID = ""
			return s.sign(t, c)
		}},
		{"alg none", func() string {
			c := validClaims()
			tok := jwt.NewWithClaims(jwt.SigningMethodNone, c)
			raw, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
			if err != nil {
				t.Fatalf("sign none: %v", err)
			}
			return raw
		}},
		{"wrong key", func() string {
			other := newSigner(t)
			return other.sign(t, validClaims())
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := v.Verify(context.Background(), tt.token())
			if !errors.Is(err, auth.ErrInvalidToken) {
				t.Errorf("Verify(%s) = %v, want ErrInvalidToken", tt.name, err)
			}
		})
	}
}
