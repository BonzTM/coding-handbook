package grpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/example/examplegrpc/internal/core"
)

// ctxKey is an unexported context key type so values cannot collide with keys
// from other packages (revive: context-keys-type).
type ctxKey int

const requestIDKey ctxKey = iota

// requestIDFrom returns the request ID stored in ctx, or "" if absent.
func requestIDFrom(ctx context.Context) string {
	id, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return id
}

// metadataRequestIDKey is the inbound/outbound metadata key correlating a call.
const metadataRequestIDKey = "x-request-id"

// authMetadataKey carries the bearer token in call metadata.
const authMetadataKey = "authorization"

// chainUnary composes interceptors so the first runs outermost. The order here
// is deliberate: recovery (outermost, catches panics in everything inside),
// request-id (so every inner layer and the handler can log it), access logging
// (times the inner call), deadline guard (bounds handler runtime), then auth
// (closest to the handler, after logging so rejected calls are still logged).
func chainUnary(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Build the chain from the inside out so interceptors[0] is outermost.
		chained := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chained = wrap(interceptors[i], info, chained)
		}
		return chained(ctx, req)
	}
}

// wrap binds one interceptor around a handler for the chain builder.
func wrap(interceptor grpc.UnaryServerInterceptor, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		return interceptor(ctx, req, info, handler)
	}
}

// recoveryUnary converts a panic in any inner interceptor or the handler into a
// codes.Internal status so a single bad request cannot crash the process. It
// logs the panic once with the stack and the request context.
func recoveryUnary(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.ErrorContext(ctx, "panic recovered",
					"method", info.FullMethod,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// requestIDUnary attaches a request ID (from inbound metadata or freshly
// generated) to the context so logs and downstream calls can correlate a
// request.
func requestIDUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		id := metadataFirst(ctx, metadataRequestIDKey)
		if id == "" {
			id = newRequestID()
		}
		return handler(context.WithValue(ctx, requestIDKey, id), req)
	}
}

// accessLogUnary emits one access log line per RPC with the full method, status
// code, and duration, plus the request ID and (when present) the W3C trace and
// span IDs pulled from the request's span context (set by the otelgrpc stats
// handler). It also records the RPC into Prometheus under low-cardinality labels
// (full method, status code).
func accessLogUnary(logger *slog.Logger, metrics rpcMetrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)

		code := status.Code(err)
		attrs := []any{
			"method", info.FullMethod,
			"code", code.String(),
			"duration_ms", elapsed.Milliseconds(),
			"request_id", requestIDFrom(ctx),
		}
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			attrs = append(attrs, "trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
		}
		// Expected client errors log at info; server-side failures at error.
		if code == codes.Internal || code == codes.Unknown || code == codes.DataLoss {
			logger.ErrorContext(ctx, "rpc", attrs...)
		} else {
			logger.InfoContext(ctx, "rpc", attrs...)
		}

		metrics.ObserveRPC(info.FullMethod, code.String(), elapsed.Seconds())
		return resp, err
	}
}

// rpcMetrics is the subset of telemetry.PromMetrics the access-logging
// interceptor needs. Defined at the consumer to keep the dependency narrow.
type rpcMetrics interface {
	ObserveRPC(method, code string, seconds float64)
}

// deadlineGuardUnary self-protects against unbounded work: when a client sends
// no deadline, it imposes a server-side ceiling so a single RPC cannot run
// forever. A client deadline that is already tighter is left untouched.
func deadlineGuardUnary(ceiling time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := ctx.Deadline(); ok {
			return handler(ctx, req)
		}
		ctx, cancel := context.WithTimeout(ctx, ceiling)
		defer cancel()
		return handler(ctx, req)
	}
}

// Authenticator verifies a bearer token and returns the resolved principal. It
// is the seam a production build implements with a JWKS verifier; the reference
// provides a static-token and a synthetic (local/dev) implementation.
type Authenticator interface {
	// Authenticate validates the bearer token and returns the principal. An
	// invalid or missing token returns core.ErrUnauthenticated.
	Authenticate(ctx context.Context, token string) (core.Principal, error)
}

// authUnary verifies the bearer token via the Authenticator and injects the
// resolved principal into the context for the handler/core to read. A missing
// or invalid token is mapped to codes.Unauthenticated.
func authUnary(authn Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token := bearerToken(metadataFirst(ctx, authMetadataKey))
		p, err := authn.Authenticate(ctx, token)
		if err != nil {
			return nil, errorFromDomain(err)
		}
		return handler(core.WithPrincipal(ctx, p), req)
	}
}

// bearerToken strips a leading "Bearer " (case-insensitive) prefix from an
// authorization metadata value.
func bearerToken(v string) string {
	const prefix = "bearer "
	if len(v) >= len(prefix) && equalFold(v[:len(prefix)], prefix) {
		return v[len(prefix):]
	}
	return v
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// metadataFirst returns the first value for key in the inbound metadata, or "".
func metadataFirst(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-unknown"
	}
	return hex.EncodeToString(b[:])
}

// StaticTokenAuthenticator accepts a single configured bearer token. It is the
// dependency-free reference Authenticator; a production build swaps in a JWKS
// verifier behind the same seam. The token comparison is constant-time-ish via a
// length check then byte compare; for a real secret use crypto/subtle.
type StaticTokenAuthenticator struct {
	token     string
	principal core.Principal
}

// NewStaticTokenAuthenticator returns an Authenticator that accepts exactly
// token and resolves it to the given principal.
func NewStaticTokenAuthenticator(token string, principal core.Principal) *StaticTokenAuthenticator {
	return &StaticTokenAuthenticator{token: token, principal: principal}
}

// Authenticate accepts the configured token and rejects everything else with
// core.ErrUnauthenticated.
func (a *StaticTokenAuthenticator) Authenticate(_ context.Context, token string) (core.Principal, error) {
	if token == "" || token != a.token {
		return core.Principal{}, core.ErrUnauthenticated
	}
	return a.principal, nil
}

// SyntheticAuthenticator is the local/dev Authenticator: it ignores the token
// and always returns a fixed synthetic principal so the service runs offline
// without an identity provider. It must never be wired when AUTH_ENABLED=true.
type SyntheticAuthenticator struct {
	principal core.Principal
}

// NewSyntheticAuthenticator returns an Authenticator that always resolves to p.
func NewSyntheticAuthenticator(p core.Principal) *SyntheticAuthenticator {
	return &SyntheticAuthenticator{principal: p}
}

// Authenticate always succeeds with the synthetic principal.
func (a *SyntheticAuthenticator) Authenticate(_ context.Context, _ string) (core.Principal, error) {
	return a.principal, nil
}
