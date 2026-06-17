package auth

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/example/exampleservice/internal/core"
)

// allowedAlgs is the explicit signature-algorithm allowlist. Pinning it closes
// the classic JWT pitfalls: "alg":"none" (unsigned tokens) and the RS/HS
// confusion attack where an attacker submits an HMAC token signed with the
// public key. Only asymmetric RSA/ECDSA signatures from the JWKS are accepted.
var allowedAlgs = []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}

// JWKSVerifier validates JWTs against a JWKS key source, pinning issuer and
// audience from config. It is the production Verifier. The keyfunc.Keyfunc owns
// the JWKS cache and (for a remote URL) the background refresh; construct it
// once in main and share it.
type JWKSVerifier struct {
	keyfunc  keyfunc.Keyfunc
	issuer   string
	audience string
}

// Compile-time proof that *JWKSVerifier satisfies the Verifier seam.
var _ Verifier = (*JWKSVerifier)(nil)

// NewJWKSVerifier builds a verifier that fetches keys from jwksURL and validates
// iss == issuer and aud contains audience. The JWKS is fetched and refreshed in
// the background under ctx (cancel it to stop the refresh goroutine). A failure
// to reach the JWKS at construction time is fail-fast: the caller (main) aborts
// startup rather than booting an auth layer that rejects every token.
func NewJWKSVerifier(ctx context.Context, jwksURL, issuer, audience string) (*JWKSVerifier, error) {
	if jwksURL == "" {
		return nil, fmt.Errorf("%w: empty JWKS URL", ErrInvalidToken)
	}
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("init JWKS keyfunc: %w", err)
	}
	return NewJWKSVerifierFromKeyfunc(kf, issuer, audience), nil
}

// NewJWKSVerifierFromKeyfunc builds a verifier from an already-constructed
// keyfunc.Keyfunc. It is the seam NewJWKSVerifier uses internally and that
// offline tests use to wire a Keyfunc backed by local in-memory JWKS storage,
// exercising the production parse path without any network.
func NewJWKSVerifierFromKeyfunc(kf keyfunc.Keyfunc, issuer, audience string) *JWKSVerifier {
	return &JWKSVerifier{keyfunc: kf, issuer: issuer, audience: audience}
}

// Verify implements Verifier against the JWKS key source.
func (v *JWKSVerifier) Verify(_ context.Context, raw string) (core.Principal, error) {
	return parseAndMap(raw, v.keyfunc.Keyfunc, v.issuer, v.audience)
}

// StaticVerifier validates JWTs with a fixed jwt.Keyfunc (e.g. a single public
// key) instead of a remote JWKS. It is the offline test seam: a test signs a
// token with a local private key and wires the matching public key here, so
// unit tests exercise the exact same parse/validate path without any network.
type StaticVerifier struct {
	keyfunc  jwt.Keyfunc
	issuer   string
	audience string
}

// Compile-time proof that *StaticVerifier satisfies the Verifier seam.
var _ Verifier = (*StaticVerifier)(nil)

// NewStaticVerifier builds a verifier from a fixed jwt.Keyfunc. The keyfunc
// returns the verification key for a token; tests typically close over a single
// public key.
func NewStaticVerifier(kf jwt.Keyfunc, issuer, audience string) *StaticVerifier {
	return &StaticVerifier{keyfunc: kf, issuer: issuer, audience: audience}
}

// Verify implements Verifier against the static key.
func (v *StaticVerifier) Verify(_ context.Context, raw string) (core.Principal, error) {
	return parseAndMap(raw, v.keyfunc, v.issuer, v.audience)
}

// parseAndMap is the shared validation path for both verifiers: parse with the
// algorithm allowlist, enforce iss/aud, then map claims to a Principal. Every
// failure is reported as ErrInvalidToken so the boundary maps a uniform 401 and
// no detail about which check failed leaks to the client (it is kept for logs).
func parseAndMap(raw string, kf jwt.Keyfunc, issuer, audience string) (core.Principal, error) {
	if raw == "" {
		return core.Principal{}, fmt.Errorf("%w: empty token", ErrInvalidToken)
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods(allowedAlgs),
		jwt.WithExpirationRequired(),
	}
	if issuer != "" {
		opts = append(opts, jwt.WithIssuer(issuer))
	}
	if audience != "" {
		opts = append(opts, jwt.WithAudience(audience))
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, kf, opts...)
	if err != nil {
		return core.Principal{}, fmt.Errorf("%w: %s", ErrInvalidToken, err.Error())
	}
	if !token.Valid {
		return core.Principal{}, fmt.Errorf("%w: token failed validation", ErrInvalidToken)
	}
	return principalFromClaims(claims)
}
