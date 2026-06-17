// Package auth verifies inbound bearer tokens and resolves the authenticated
// core.Principal the rest of the service authorizes against. It is the one
// default the handbook ships for identity: a JWT validated against a JWKS key
// source, with issuer/audience pinned from config.
//
// The SCHEME choice (JWT + JWKS, bearer in the Authorization header) is an ADR;
// this package is the single copyable implementation of that decision. The
// Verifier seam keeps the transport layer independent of the JWT library and
// lets offline tests sign with a local key instead of reaching a real JWKS URL.
//
// It depends only on internal/core (for the Principal type) and the JWT/JWKS
// libraries; it imports no transport package, so the dependency direction in
// golang/foundations/package-design.md holds.
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/exampleservice/internal/core"
)

// ErrInvalidToken is returned for any token that fails verification: missing,
// malformed, expired, wrong issuer/audience, unverifiable signature, or missing
// required claims. The transport maps it to 401. A single sentinel keeps the
// boundary mapping simple and avoids leaking which check failed to the client.
var ErrInvalidToken = errors.New("invalid token")

// Verifier validates a raw bearer token string and returns the authenticated
// principal. It is the consumer-defined seam the auth middleware depends on:
// production wires JWKSVerifier (JWT + remote JWKS); tests wire a StaticVerifier
// backed by a local key so no network is touched. Any verification failure is
// reported as ErrInvalidToken (wrapped with detail for logs).
type Verifier interface {
	// Verify parses and validates raw and returns the resolved principal, or an
	// ErrInvalidToken-wrapping error. It honors ctx for cancellation/timeouts.
	Verify(ctx context.Context, raw string) (core.Principal, error)
}

// Claims is the JWT claim set the service requires beyond the registered
// claims (iss/aud/exp validated by the parser). tenant_id and roles are the
// custom claims that drive multi-tenancy and RBAC.
type Claims struct {
	jwt.RegisteredClaims

	// TenantID scopes every request to one tenant. A token without it cannot be
	// authorized (the service would not know which tenant's data to touch), so
	// it is required.
	TenantID string `json:"tenant_id"`
	// Roles are the caller's RBAC roles. May be empty (authenticated but
	// unprivileged); the boundary authorization check rejects missing roles.
	Roles []string `json:"roles"`
}

// principalFromClaims maps validated claims to a core.Principal. It requires sub
// and tenant_id; a token missing either is unusable for a tenant-scoped,
// subject-identified operation and is rejected as an invalid token.
func principalFromClaims(c *Claims) (core.Principal, error) {
	if c.Subject == "" {
		return core.Principal{}, fmt.Errorf("%w: missing sub claim", ErrInvalidToken)
	}
	if c.TenantID == "" {
		return core.Principal{}, fmt.Errorf("%w: missing tenant_id claim", ErrInvalidToken)
	}
	roles := make([]core.Role, 0, len(c.Roles))
	for _, r := range c.Roles {
		roles = append(roles, core.Role(r))
	}
	return core.Principal{
		Subject:  c.Subject,
		TenantID: c.TenantID,
		Roles:    roles,
	}, nil
}
