package core

import (
	"context"
	"errors"
	"slices"
)

// Identity sentinel errors the transport boundary branches on with errors.Is.
// They are part of the package contract: the HTTP layer maps them to 401/403.
var (
	// ErrUnauthenticated is returned when an operation requires an authenticated
	// principal but the context carries none. The transport maps it to 401.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrForbidden is returned when an authenticated principal lacks the role or
	// ownership required for an operation. The transport maps it to 403.
	ErrForbidden = errors.New("forbidden")
)

// Role is a coarse authorization role carried on a Principal. Roles are
// low-cardinality, stable strings; the service requires a role for an operation
// and additionally enforces a per-resource tenant/ownership check at the
// boundary, per golang/services/http-services.md (authorize at the edge, not in
// helpers).
type Role string

const (
	// RoleReader may read widgets within its tenant.
	RoleReader Role = "widgets.reader"
	// RoleWriter may create widgets within its tenant. A writer is also a reader.
	RoleWriter Role = "widgets.writer"
)

// Principal is the authenticated caller resolved from a verified bearer token.
// It is a plain value type with no wire concerns: the auth layer builds it from
// validated JWT claims and threads it through the request context. Subject and
// TenantID are required; Roles may be empty (an authenticated but unprivileged
// caller).
type Principal struct {
	// Subject is the token "sub": the stable caller identity.
	Subject string
	// TenantID is the tenant the caller acts within ("tenant_id" claim). Every
	// store query is scoped by it so one tenant can never read another's rows.
	TenantID string
	// Roles are the caller's authorization roles ("roles" claim).
	Roles []Role
}

// HasRole reports whether the principal carries the given role.
func (p Principal) HasRole(want Role) bool {
	return slices.Contains(p.Roles, want)
}

// principalCtxKey is an unexported context key type so principal values cannot
// collide with keys from other packages (revive: context-keys-type).
type principalCtxKey struct{}

// WithPrincipal returns a child context carrying the authenticated principal.
// The auth middleware calls it after verifying a token; the service reads it
// back with PrincipalFrom. The principal is never stored on a struct; it travels
// only on the request context.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalCtxKey{}, p)
}

// PrincipalFrom returns the principal stored in ctx. The boolean is false when
// no principal is present (an unauthenticated request reaching code that
// expected one), so callers fail closed rather than acting as an empty tenant.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalCtxKey{}).(Principal)
	return p, ok
}
