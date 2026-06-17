package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/exampleservice/internal/auth"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/telemetry"
)

// Audit action labels. They are low-cardinality, stable strings naming the
// security-relevant operation, per golang/operations/security.md ### Audit
// Logging. Kept as constants so the audit-event schema is a single source of
// truth shared by the middleware and handlers.
const (
	auditActionAuthenticate = "authenticate"
	auditActionAuthorize    = "authorize"
	auditActionWidgetCreate = "widget.create"
)

// authMiddleware validates the inbound Bearer token and attaches the resolved
// principal to the request context. It is the AUTHN layer: a missing or invalid
// token is a 401 and the request never reaches the handler. It runs after
// request-ID/trace wiring (so a rejected request is still correlated) and before
// authz, logging, and the handler.
//
// When verifier is nil the service is in local/dev mode (AUTH_ENABLED=false):
// the middleware injects a fixed local-dev principal so the tenant-scoped
// service still functions offline without an identity provider. This is the
// config gate described in the README security section.
func authMiddleware(verifier auth.Verifier, logger *slog.Logger, audit *telemetry.AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				// Local/dev mode: a deterministic principal with both roles so the
				// service is exercisable end to end without a token. NEVER reached
				// when AUTH_ENABLED=true (the verifier is non-nil then).
				ctx := core.WithPrincipal(r.Context(), localDevPrincipal())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			raw, ok := bearerToken(r)
			if !ok {
				// Authentication failure: no principal is established, so the audit
				// record carries no actor/tenant — only the action, result, route, and
				// request id. The raw token is NEVER placed on the record.
				auditAuthnFailure(r, audit)
				writeError(w, r, logger, core.ErrUnauthenticated)
				return
			}
			principal, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				// Verification detail (which check failed) is logged at the boundary
				// but the client sees a uniform 401; map the auth error to the core
				// sentinel so statusForError yields 401 without leaking detail. The
				// audit record records the failure fact only — never the token or the
				// specific verification error (which can hint at a valid claim shape).
				auditAuthnFailure(r, audit)
				writeError(w, r, logger, unauthenticated(err))
				return
			}
			ctx := core.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// auditAuthnFailure emits the audit event for an authentication failure. No
// principal exists, so actor/tenant are empty. The authn check runs BEFORE the
// API mux matches a route, so r.Pattern is not yet set; the resource is left
// empty rather than recording the raw request path (which is high-cardinality
// and may carry sensitive identifiers). No credential material is recorded.
func auditAuthnFailure(r *http.Request, audit *telemetry.AuditLogger) {
	audit.Emit(r.Context(), telemetry.AuditEvent{
		Action:    auditActionAuthenticate,
		Result:    telemetry.AuditFailure,
		RequestID: requestIDFrom(r.Context()),
	})
}

// requireRole is the route-scoped AUTHZ layer: it enforces that the
// authenticated principal carries the role a route requires BEFORE the handler
// runs, returning 403 on denial. The complementary per-resource tenant
// ownership check is enforced in the tenant-scoped store (a cross-tenant read is
// a 404), so authorization is checked at the boundary, not buried in a helper.
func requireRole(want core.Role, logger *slog.Logger, audit *telemetry.AuditLogger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := core.PrincipalFrom(r.Context())
		if !ok {
			// No principal reached the authz layer: treat as an authentication
			// failure for the audit trail (no actor to attribute the denial to).
			auditAuthnFailure(r, audit)
			writeError(w, r, logger, core.ErrUnauthenticated)
			return
		}
		if !p.HasRole(want) {
			// Authorization denial: the caller is known (actor + tenant) but lacks the
			// required role. The required role is low-cardinality metadata, safe to
			// record; no token or payload is included.
			audit.Emit(r.Context(), telemetry.AuditEvent{
				Actor:     p.Subject,
				Tenant:    p.TenantID,
				Action:    auditActionAuthorize,
				Resource:  routePattern(r),
				Result:    telemetry.AuditDenied,
				RequestID: requestIDFrom(r.Context()),
			})
			writeError(w, r, logger, forbidden(p.Subject, want))
			return
		}
		next(w, r)
	}
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
// The scheme match is case-insensitive per RFC 7235; a missing header or other
// scheme yields ok=false.
func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	const prefix = "bearer "
	if len(h) < len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(h[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}

// localDevPrincipal is the synthetic principal used when auth is disabled. It is
// a single tenant with both roles so the in-memory store and authz checks are
// exercisable locally; it is never produced when AUTH_ENABLED=true.
func localDevPrincipal() core.Principal {
	return core.Principal{
		Subject:  "local-dev",
		TenantID: "local-dev",
		Roles:    []core.Role{core.RoleReader, core.RoleWriter},
	}
}
