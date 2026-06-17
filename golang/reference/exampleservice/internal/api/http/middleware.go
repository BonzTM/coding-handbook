package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
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

// routePattern returns the matched ServeMux route pattern for low-cardinality
// logging and metrics labels (never the raw path). Go 1.22+ exposes it on the
// request once routing has matched.
func routePattern(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	return "unmatched"
}

// statusRecorder captures the response status for access logging without
// buffering the body.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// recoverMiddleware is the outermost layer: it converts a panic in any inner
// handler or middleware into a 500 so a single bad request cannot crash the
// process. It logs the panic once with the request context.
func recoverMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"method", r.Method,
						"route", routePattern(r),
						"panic", rec,
					)
					writeJSON(w, r, logger, http.StatusInternalServerError,
						errorResponse{Error: http.StatusText(http.StatusInternalServerError)})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// requestIDMiddleware attaches a request ID (from the inbound header or freshly
// generated) to the context and echoes it in the response, so logs and clients
// can correlate a request.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware emits one access log line per request with method, route
// pattern, status, and duration, plus the request ID. It records metrics with
// low-cardinality labels (route pattern + status class).
func loggingMiddleware(logger *slog.Logger, metrics requestMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			route := routePattern(r)
			logger.InfoContext(r.Context(), "request",
				"method", r.Method,
				"route", route,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestIDFrom(r.Context()),
			)
			metrics.IncRequest(route, statusClass(rec.status))
		})
	}
}

// requestMetrics is the subset of telemetry.Metrics the logging middleware
// needs. Defined at the consumer to keep the dependency narrow.
type requestMetrics interface {
	IncRequest(routePattern, statusClass string)
}

// statusClass collapses a status code into a low-cardinality class label
// ("2xx", "4xx", ...) so metric cardinality stays bounded.
func statusClass(status int) string {
	switch {
	case status < 200:
		return "1xx"
	case status < 300:
		return "2xx"
	case status < 400:
		return "3xx"
	case status < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

func newRequestID() string {
	var b [16]byte
	// crypto/rand.Read never returns an error on supported platforms; on the
	// off chance it does, fall back to a fixed marker rather than panicking in
	// a hot path.
	if _, err := rand.Read(b[:]); err != nil {
		return "req-unknown"
	}
	return hex.EncodeToString(b[:])
}
