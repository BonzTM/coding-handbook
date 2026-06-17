package grpc

import (
	"context"
	"errors"

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
	case errors.Is(err, core.ErrInvalidWidget), errors.Is(err, core.ErrInvalidCursor):
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

// errorFromDomain is the convenience form returning the error directly for a
// server method's return statement.
func errorFromDomain(err error) error {
	if err == nil {
		return nil
	}
	st, _ := statusFromDomain(err)
	return st.Err()
}
