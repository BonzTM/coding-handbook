package grpc

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/example/examplegrpc/internal/core"
)

// statusFromDomain maps a domain error to a gRPC *status.Status, mapping once at
// the transport boundary per golang/recipes/add-grpc-method.md. Core wraps its
// sentinels with %w; this is the single place those sentinels become codes.*.
// A raw internal error becomes codes.Internal with a generic message so we never
// leak internals to clients. A nil error maps to codes.OK.
//
// The returned bool reports whether err was an expected (mapped) domain error;
// the access-logging interceptor uses it to log unexpected Internal errors at a
// higher level.
func statusFromDomain(err error) (*status.Status, bool) {
	if err == nil {
		return status.New(codes.OK, ""), true
	}
	switch {
	case errors.Is(err, core.ErrNotFound):
		return status.New(codes.NotFound, err.Error()), true
	case errors.Is(err, core.ErrAlreadyExists):
		return status.New(codes.AlreadyExists, err.Error()), true
	case errors.Is(err, core.ErrInvalidWidget):
		// A field-level validation failure becomes InvalidArgument WITH a
		// google.rpc.BadRequest detail so clients get machine-readable, per-field
		// violations (which field, why) rather than only a flat message. core
		// carries the violations structurally (core.FieldViolations); the proto
		// detail is built here at the transport boundary.
		return invalidArgumentWithDetails(err), true
	case errors.Is(err, core.ErrInvalidCursor):
		return status.New(codes.InvalidArgument, err.Error()), true
	case errors.Is(err, core.ErrUnauthenticated):
		return status.New(codes.Unauthenticated, err.Error()), true
	case errors.Is(err, core.ErrForbidden):
		return status.New(codes.PermissionDenied, err.Error()), true
	case errors.Is(err, context.Canceled):
		return status.New(codes.Canceled, "request canceled"), true
	case errors.Is(err, context.DeadlineExceeded):
		return status.New(codes.DeadlineExceeded, "deadline exceeded"), true
	default:
		// Unexpected: do not leak the internal message to the client.
		return status.New(codes.Internal, "internal error"), false
	}
}

// invalidArgumentWithDetails builds a codes.InvalidArgument status and, when the
// domain error carries structured field violations, attaches a
// google.rpc.BadRequest detail listing each {field, description}. If attaching
// the detail fails (it cannot in practice for this fixed message type) or there
// are no violations, it falls back to the plain status so a client always gets a
// well-formed InvalidArgument.
func invalidArgumentWithDetails(err error) *status.Status {
	st := status.New(codes.InvalidArgument, err.Error())

	violations := core.FieldViolations(err)
	if len(violations) == 0 {
		return st
	}

	br := &errdetails.BadRequest{
		FieldViolations: make([]*errdetails.BadRequest_FieldViolation, 0, len(violations)),
	}
	for _, v := range violations {
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       v.Field,
			Description: v.Description,
		})
	}

	withDetails, derr := st.WithDetails(br)
	if derr != nil {
		// WithDetails only fails if the detail cannot be marshaled, which cannot
		// happen for this fixed proto; keep the plain status rather than drop the
		// error entirely.
		return st
	}
	return withDetails
}

// errorFromDomain is the convenience form returning the error directly for a
// server method's return statement.
func errorFromDomain(err error) error {
	if err == nil {
		return nil
	}
	st, _ := statusFromDomain(err)
	return st.Err()
}
