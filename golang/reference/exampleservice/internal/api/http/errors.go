// Package http is the HTTP transport adapter for the widgets service. It is a
// translation layer only: it decodes and validates input, calls one core
// method, maps domain errors to HTTP status codes, and encodes the response. It
// depends on internal/core, internal/config, and internal/telemetry; it never
// touches SQL directly, per golang/services/http-services.md.
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/example/exampleservice/internal/core"
)

// errorResponse is the wire shape for an error. It is a dedicated DTO with
// explicit json tags, per golang/foundations/serialization.md.
type errorResponse struct {
	Error string `json:"error"`
}

// statusForError maps a domain error to an HTTP status code. This is the single
// place domain semantics become transport semantics; handlers do not branch on
// errors themselves beyond calling this.
func statusForError(err error) int {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, core.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, core.ErrInvalidWidget):
		return http.StatusBadRequest
	case errors.Is(err, core.ErrInvalidCursor):
		return http.StatusBadRequest
	case errors.Is(err, core.ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, core.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, core.ErrIdempotencyInFlight):
		return http.StatusConflict
	case errors.Is(err, core.ErrIdempotencyKeyMismatch):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// unauthenticated wraps an auth-verification failure as core.ErrUnauthenticated
// so statusForError yields 401. The underlying detail (which check failed) is
// preserved for the boundary log but never reaches the client: the response
// body for a 401 is the generic status text.
func unauthenticated(err error) error {
	return fmt.Errorf("%w: %s", core.ErrUnauthenticated, err.Error())
}

// forbidden builds a core.ErrForbidden-wrapping error naming the missing role,
// so statusForError yields 403 and the boundary log records what was required.
func forbidden(subject string, want core.Role) error {
	return fmt.Errorf("%w: %s lacks role %s", core.ErrForbidden, subject, want)
}

// writeError maps err to a status, logs it once at the boundary, and encodes a
// safe error body. Internal errors are logged with detail but the client only
// sees a generic message so implementation details do not leak.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status := statusForError(err)

	// Log once, here at the boundary, with stable fields. 5xx is the
	// unexpected class and gets error level; client errors are info.
	attrs := []any{
		"method", r.Method,
		"route", routePattern(r),
		"status", status,
		"error", err.Error(),
	}
	if status >= http.StatusInternalServerError {
		logger.ErrorContext(r.Context(), "request failed", attrs...)
	} else {
		logger.InfoContext(r.Context(), "request rejected", attrs...)
	}

	msg := http.StatusText(status)
	// For client-class errors the error message is safe and actionable; for
	// server-class errors we hide the detail behind the generic status text. A
	// 401 is deliberately also generic: leaking which token check failed
	// (expired vs. wrong audience vs. missing claim) helps an attacker, so the
	// detail stays in the boundary log only.
	if status < http.StatusInternalServerError && status != http.StatusUnauthorized {
		msg = err.Error()
	}
	writeJSON(w, r, logger, status, errorResponse{Error: msg})
}

// writeJSON encodes v as JSON with the given status. Encoding errors are logged
// by the caller's recovery middleware if they panic; a failed write to an
// already-started response cannot be recovered, so we ignore the encode error
// deliberately after headers are sent.
func writeJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// The status and headers are already committed; a late encode error cannot
	// change the response. We cannot tell the client, but we log it once so a
	// broken connection or marshal bug is observable rather than silent.
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.WarnContext(r.Context(), "write response body failed",
			"route", routePattern(r),
			"status", status,
			"error", err.Error(),
		)
	}
}
