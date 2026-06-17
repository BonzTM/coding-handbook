package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	widgetv1 "github.com/example/examplegrpc/api/widget/v1"
	grpcapi "github.com/example/examplegrpc/internal/api/grpc"
	"github.com/example/examplegrpc/internal/config"
	"github.com/example/examplegrpc/internal/core"
	"github.com/example/examplegrpc/internal/telemetry"
	"github.com/example/examplegrpc/internal/testutil"
)

const bufSize = 1 << 20

// harness wires an in-process gRPC server over a bufconn pipe with the full
// interceptor chain, returning a connected client.
type harness struct {
	client widgetv1.WidgetServiceClient
	srv    *grpc.Server
}

func newHarness(t *testing.T, authn grpcapi.Authenticator) *harness {
	t.Helper()

	clk := testutil.NewFakeClock(time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))
	svc := core.NewService(core.NewMemory(), clk)
	logger := telemetry.NewLogger(testWriter{t}, 0, false)
	metrics := telemetry.NewPromMetrics("test")

	srv, _ := grpcapi.NewGRPCServer(config.GRPCConfig{
		MaxRecvMsgBytes: 1 << 20,
		HandlerTimeout:  5 * time.Second,
		ConnTimeout:     5 * time.Second,
	}, grpcapi.Deps{Service: svc, Logger: logger, Metrics: metrics, Authn: authn})

	lis := bufconn.Listen(bufSize)
	go func() {
		if err := srv.Serve(lis); err != nil {
			// Serve returns when the listener closes during teardown.
			return
		}
	}()

	//nolint:staticcheck // grpc.DialContext with a bufconn dialer is the documented in-process test pattern.
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufnet: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop()
	})

	return &harness{client: widgetv1.NewWidgetServiceClient(conn), srv: srv}
}

// testWriter routes server logs to the test log.
type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Logf("server: %s", p)
	return len(p), nil
}

func writerPrincipal() core.Principal {
	return core.Principal{Subject: "u", TenantID: "t", Roles: []core.Role{core.RoleReader, core.RoleWriter}}
}

func syntheticAuthn() grpcapi.Authenticator {
	return grpcapi.NewSyntheticAuthenticator(writerPrincipal())
}

func TestCreateGetHappyPath(t *testing.T) {
	h := newHarness(t, syntheticAuthn())
	ctx := context.Background()

	created, err := h.client.CreateWidget(ctx, &widgetv1.CreateWidgetRequest{Id: "w1", Name: "One"})
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	if created.GetWidget().GetId() != "w1" || created.GetWidget().GetName() != "One" {
		t.Errorf("created widget = %+v", created.GetWidget())
	}
	if created.GetWidget().GetCreatedAt() == nil {
		t.Error("created_at must be set on the wire")
	}

	got, err := h.client.GetWidget(ctx, &widgetv1.GetWidgetRequest{Id: "w1"})
	if err != nil {
		t.Fatalf("GetWidget: %v", err)
	}
	if got.GetWidget().GetName() != "One" {
		t.Errorf("got widget name = %q, want One", got.GetWidget().GetName())
	}
}

func TestGetNotFoundMapsToNotFound(t *testing.T) {
	h := newHarness(t, syntheticAuthn())
	_, err := h.client.GetWidget(context.Background(), &widgetv1.GetWidgetRequest{Id: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %v, want NotFound (err=%v)", status.Code(err), err)
	}
}

func TestCreateDuplicateMapsToAlreadyExists(t *testing.T) {
	h := newHarness(t, syntheticAuthn())
	ctx := context.Background()
	if _, err := h.client.CreateWidget(ctx, &widgetv1.CreateWidgetRequest{Id: "dup", Name: "a"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := h.client.CreateWidget(ctx, &widgetv1.CreateWidgetRequest{Id: "dup", Name: "b"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("code = %v, want AlreadyExists (err=%v)", status.Code(err), err)
	}
}

func TestCreateInvalidMapsToInvalidArgument(t *testing.T) {
	h := newHarness(t, syntheticAuthn())
	_, err := h.client.CreateWidget(context.Background(), &widgetv1.CreateWidgetRequest{Id: "", Name: "x"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument (err=%v)", status.Code(err), err)
	}
}

func TestListKeysetPagination(t *testing.T) {
	h := newHarness(t, syntheticAuthn())
	ctx := context.Background()

	for i := range 5 {
		id := string(rune('a' + i))
		if _, err := h.client.CreateWidget(ctx, &widgetv1.CreateWidgetRequest{Id: id, Name: id}); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}

	var ids []string
	token := ""
	pages := 0
	for {
		resp, err := h.client.ListWidgets(ctx, &widgetv1.ListWidgetsRequest{PageSize: 2, PageToken: token})
		if err != nil {
			t.Fatalf("ListWidgets: %v", err)
		}
		pages++
		for _, w := range resp.GetWidgets() {
			ids = append(ids, w.GetId())
		}
		token = resp.GetNextPageToken()
		if token == "" {
			break
		}
		if pages > 10 {
			t.Fatal("pagination did not terminate")
		}
	}
	if len(ids) != 5 {
		t.Fatalf("got %d widgets across pages, want 5 (%v)", len(ids), ids)
	}
	if pages != 3 {
		t.Errorf("pages = %d, want 3", pages)
	}
}

func TestListInvalidTokenMapsToInvalidArgument(t *testing.T) {
	h := newHarness(t, syntheticAuthn())
	_, err := h.client.ListWidgets(context.Background(), &widgetv1.ListWidgetsRequest{PageToken: "!!!bad!!!"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument (err=%v)", status.Code(err), err)
	}
}

func TestAuthInterceptorRejectsBadToken(t *testing.T) {
	authn := grpcapi.NewStaticTokenAuthenticator("s3cret", writerPrincipal())
	h := newHarness(t, authn)

	// No token -> Unauthenticated.
	_, err := h.client.GetWidget(context.Background(), &widgetv1.GetWidgetRequest{Id: "x"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing token code = %v, want Unauthenticated", status.Code(err))
	}

	// Valid token -> NotFound (auth passed, widget absent).
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer s3cret")
	_, err = h.client.GetWidget(ctx, &widgetv1.GetWidgetRequest{Id: "x"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("valid token code = %v, want NotFound (auth should pass)", status.Code(err))
	}
}
