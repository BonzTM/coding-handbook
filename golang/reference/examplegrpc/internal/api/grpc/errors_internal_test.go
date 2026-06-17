package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/example/examplegrpc/internal/core"
)

func TestStatusFromDomainMapping(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode codes.Code
		expected bool
	}{
		{"nil", nil, codes.OK, true},
		{"not found", core.ErrNotFound, codes.NotFound, true},
		{"already exists", core.ErrAlreadyExists, codes.AlreadyExists, true},
		{"invalid widget", fmt.Errorf("wrap: %w", core.ErrInvalidWidget), codes.InvalidArgument, true},
		{"invalid cursor", core.ErrInvalidCursor, codes.InvalidArgument, true},
		{"unauthenticated", core.ErrUnauthenticated, codes.Unauthenticated, true},
		{"forbidden", core.ErrForbidden, codes.PermissionDenied, true},
		{"canceled", context.Canceled, codes.Canceled, true},
		{"deadline", context.DeadlineExceeded, codes.DeadlineExceeded, true},
		{"unexpected", errors.New("disk on fire"), codes.Internal, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st, expected := statusFromDomain(tc.err)
			if st.Code() != tc.wantCode {
				t.Errorf("code = %v, want %v", st.Code(), tc.wantCode)
			}
			if expected != tc.expected {
				t.Errorf("expected flag = %v, want %v", expected, tc.expected)
			}
			// Unexpected errors must not leak their internal message.
			if !tc.expected && st.Message() == "disk on fire" {
				t.Error("internal error message leaked")
			}
		})
	}
}

func TestErrorFromDomainPreservesCode(t *testing.T) {
	err := errorFromDomain(core.ErrNotFound)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %v, want NotFound", status.Code(err))
	}
	if errorFromDomain(nil) != nil {
		t.Error("nil error must map to nil")
	}
}
