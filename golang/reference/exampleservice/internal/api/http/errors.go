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

// ErrorResponse is the single structured error envelope returned by EVERY
// endpoint, per golang/foundations/serialization.md ### Error Responses. It is a
// dedicated DTO with explicit snake_case json tags. A bare {"error":"..."}
// string is forbidden (decisions/framework-selection.md): the client gets one
// human sentence and nothing to branch on. The envelope carries a
// machine-readable code, a safe human message, optional per-field validation
// failures, and the correlation request_id in the body.
type ErrorResponse struct {
	// Code is a machine-readable string enum the client may branch on. It is NOT
	// the HTTP status: two 404s with different codes are different failures.
	Code string `json:"code"`
	// Message is human-readable and safe to surface. For 5xx it is generic and
	// never carries internal detail.
	Message string `json:"message"`
	// Fields carries one entry per offending input on a validation failure; it is
	// omitted (omitzero) when empty.
	Fields []fieldError `json:"fields,omitzero"`
	// RequestID is the correlation id, echoed in the body (not only the header) so
	// a client can quote it. Omitted when absent.
	RequestID string `json:"request_id,omitzero"`
}

// fieldError carries one validation failure against one input field, per the
// serialization doc's FieldError shape.
type fieldError struct {
	// Field is a dotted path into the request body, e.g. "name" or "items.0.qty".
	Field string `json:"field"`
	// Code is a small machine-readable enum: "required", "out_of_range", etc.
	Code string `json:"code"`
	// Message is the human-readable detail for this field.
	Message string `json:"message"`
}

// Machine-readable top-level error codes. They are a documented part of the wire
// contract and evolve additively, per
// golang/foundations/contracts-and-compatibility.md: add codes, never repurpose
// or silently drop one. They are deliberately NOT the HTTP status.
const (
	codeNotFound         = "not_found"
	codeAlreadyExists    = "already_exists"
	codeInvalidArgument  = "invalid_argument"
	codeValidationFailed = "validation_failed"
	codeUnauthenticated  = "unauthenticated"
	codeForbidden        = "forbidden"
	codeConflict         = "conflict"
	codeInternal         = "internal"
)

// Note: the wire contract also documents codes for failures this reference does
// not (yet) produce — e.g. "rate_limited" for a 429. They are added to this enum
// (and to errorClass) when the corresponding sentinel is introduced; the code
// set evolves additively, never by repurposing an existing value.

// errorClass is the boundary mapping from a domain error to its documented
// (status, code) pair. This is the single place domain semantics become
// transport semantics; handlers do not branch on errors themselves beyond
// calling writeError, which calls this.
func errorClass(err error) (status int, code string) {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return http.StatusNotFound, codeNotFound
	case errors.Is(err, core.ErrAlreadyExists):
		return http.StatusConflict, codeAlreadyExists
	case errors.Is(err, core.ErrInvalidWidget):
		// A widget validation failure is a 400. When the error carries field-level
		// detail (see fieldErrorsFor) writeError upgrades the code to
		// validation_failed; the plain case is invalid_argument.
		return http.StatusBadRequest, codeInvalidArgument
	case errors.Is(err, core.ErrInvalidCursor):
		return http.StatusBadRequest, codeInvalidArgument
	case errors.Is(err, core.ErrMissingIdempotencyKey):
		// A required Idempotency-Key was absent on an unsafe write: a client bug,
		// mapped to 400, per recipes/add-idempotent-write.md.
		return http.StatusBadRequest, codeInvalidArgument
	case errors.Is(err, core.ErrUnauthenticated):
		return http.StatusUnauthorized, codeUnauthenticated
	case errors.Is(err, core.ErrForbidden):
		return http.StatusForbidden, codeForbidden
	case errors.Is(err, core.ErrIdempotencyInFlight):
		return http.StatusConflict, codeConflict
	case errors.Is(err, core.ErrIdempotencyKeyMismatch):
		return http.StatusUnprocessableEntity, codeValidationFailed
	default:
		return http.StatusInternalServerError, codeInternal
	}
}

// unauthenticated wraps an auth-verification failure as core.ErrUnauthenticated
// so errorClass yields (401, "unauthenticated"). The underlying detail (which
// check failed) is preserved for the boundary log but never reaches the client:
// the response body for a 401 is the generic status text.
func unauthenticated(err error) error {
	return fmt.Errorf("%w: %s", core.ErrUnauthenticated, err.Error())
}

// forbidden builds a core.ErrForbidden-wrapping error naming the missing role,
// so errorClass yields (403, "forbidden") and the boundary log records what was
// required.
func forbidden(subject string, want core.Role) error {
	return fmt.Errorf("%w: %s lacks role %s", core.ErrForbidden, subject, want)
}

// writeError maps err to a (status, code) pair, logs it once at the boundary,
// and encodes the single structured envelope. Internal errors are logged with
// detail but the client only sees a generic message so implementation details
// do not leak. The request_id is echoed into the body (not only the header) so a
// client can quote it for correlation.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status, code := errorClass(err)

	// Log once, here at the boundary, with stable fields. 5xx is the
	// unexpected class and gets error level; client errors are info.
	attrs := []any{
		"method", r.Method,
		"route", routePattern(r),
		"status", status,
		"code", code,
		"error", err.Error(),
	}
	if status >= http.StatusInternalServerError {
		logger.ErrorContext(r.Context(), "request failed", attrs...)
	} else {
		logger.InfoContext(r.Context(), "request rejected", attrs...)
	}

	resp := ErrorResponse{
		Code:      code,
		Message:   safeMessage(status, err),
		RequestID: requestIDFrom(r.Context()),
	}
	// A validation failure carries per-field detail so a client can attach
	// messages to form fields without parsing prose. When present, the top-level
	// code is the dedicated validation_failed enum.
	if status < http.StatusInternalServerError {
		if fields := fieldErrorsFor(err); len(fields) > 0 {
			resp.Fields = fields
			resp.Code = codeValidationFailed
			resp.Message = "the request has invalid fields"
		}
	}
	writeJSON(w, r, logger, status, resp)
}

// safeMessage returns a client-safe human message. Client-class (4xx) errors
// carry an actionable message; server-class (5xx) errors are opaque — the detail
// goes to the boundary log under the request_id, never the body. A 401 is also
// deliberately generic: leaking which token check failed (expired vs. wrong
// audience vs. missing claim) helps an attacker.
func safeMessage(status int, err error) string {
	if status >= http.StatusInternalServerError || status == http.StatusUnauthorized {
		return http.StatusText(status)
	}
	return err.Error()
}

// fieldErrorsFor extracts per-field validation failures from err for the
// envelope's fields array. It returns entries only for a core widget-validation
// failure that carries a named field; other client errors yield none and the
// envelope reports only the top-level code/message.
func fieldErrorsFor(err error) []fieldError {
	if fe, ok := errors.AsType[core.FieldValidationError](err); ok {
		return []fieldError{{Field: fe.Field, Code: fe.Code, Message: fe.Reason}}
	}
	return nil
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
