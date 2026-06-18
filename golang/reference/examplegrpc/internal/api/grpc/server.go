// Package grpc is the gRPC transport adapter for the widgets service. Server
// methods are thin: they decode the request message, call exactly one core
// method, map domain errors to codes.* via errors.go, and build the response.
// No domain logic lives here, per golang/services/grpc-services.md.
package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	widgetv1 "github.com/example/examplegrpc/api/widget/v1"
	"github.com/example/examplegrpc/internal/core"
)

// widgetSvc is the subset of *core.Service the server depends on. Defining it at
// the consumer keeps the transport decoupled from the concrete service and makes
// the handler trivially testable with a fake.
type widgetSvc interface {
	CreateWidget(ctx context.Context, id, name string) (core.Widget, error)
	GetWidget(ctx context.Context, id string) (core.Widget, error)
	ListWidgetsPage(ctx context.Context, after core.Cursor, pageSize int) (core.Page, error)
}

// Server implements widgetv1.WidgetServiceServer over the core widgets service.
// It embeds the generated Unimplemented type for forward compatibility so a new
// RPC added to the proto does not break the build before it is implemented.
type Server struct {
	widgetv1.UnimplementedWidgetServiceServer
	svc widgetSvc
}

// NewServer constructs a Server backed by the given core service.
func NewServer(svc widgetSvc) *Server {
	return &Server{svc: svc}
}

// CreateWidget decodes the request, calls core, and maps domain errors to
// codes.*.
func (s *Server) CreateWidget(ctx context.Context, req *widgetv1.CreateWidgetRequest) (*widgetv1.CreateWidgetResponse, error) {
	w, err := s.svc.CreateWidget(ctx, req.GetId(), req.GetName())
	if err != nil {
		return nil, errorFromDomain(err)
	}
	return &widgetv1.CreateWidgetResponse{Widget: toProto(w)}, nil
}

// GetWidget decodes the request, calls core, and maps a missing widget to
// codes.NotFound.
func (s *Server) GetWidget(ctx context.Context, req *widgetv1.GetWidgetRequest) (*widgetv1.GetWidgetResponse, error) {
	w, err := s.svc.GetWidget(ctx, req.GetId())
	if err != nil {
		return nil, errorFromDomain(err)
	}
	return &widgetv1.GetWidgetResponse{Widget: toProto(w)}, nil
}

// ListWidgets decodes the keyset page token, calls core, and renders the page
// plus the next page token. A malformed token is mapped to InvalidArgument by
// errors.go via core.ErrInvalidCursor.
func (s *Server) ListWidgets(ctx context.Context, req *widgetv1.ListWidgetsRequest) (*widgetv1.ListWidgetsResponse, error) {
	after, err := core.DecodeCursor(req.GetPageToken())
	if err != nil {
		return nil, errorFromDomain(err)
	}
	page, err := s.svc.ListWidgetsPage(ctx, after, int(req.GetPageSize()))
	if err != nil {
		return nil, errorFromDomain(err)
	}

	out := &widgetv1.ListWidgetsResponse{
		Widgets:       make([]*widgetv1.Widget, 0, len(page.Widgets)),
		NextPageToken: core.EncodeCursor(page.NextCursor),
	}
	for _, w := range page.Widgets {
		out.Widgets = append(out.Widgets, toProto(w))
	}
	return out, nil
}

// toProto renders a domain widget as its wire DTO. The tenant ID is never put on
// the wire: it is an internal scoping concern resolved from the principal.
func toProto(w core.Widget) *widgetv1.Widget {
	return &widgetv1.Widget{
		Id:        w.ID,
		Name:      w.Name,
		CreatedAt: timestamppb.New(w.CreatedAt),
	}
}
