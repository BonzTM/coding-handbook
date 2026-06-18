package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	widgetv1 "github.com/example/examplegrpc/api/widget/v1"
	grpcapi "github.com/example/examplegrpc/internal/api/grpc"
	"github.com/example/examplegrpc/internal/config"
	"github.com/example/examplegrpc/internal/core"
	"github.com/example/examplegrpc/internal/telemetry"
)

// panicSvc panics on every call to exercise the recovery interceptor.
type panicSvc struct{}

func (panicSvc) CreateWidget(context.Context, string, string) (core.Widget, error) {
	panic("boom")
}

func (panicSvc) GetWidget(context.Context, string) (core.Widget, error) { panic("boom") }

func (panicSvc) ListWidgetsPage(context.Context, core.Cursor, int) (core.Page, error) {
	panic("boom")
}

func TestRecoveryInterceptorMapsPanicToInternal(t *testing.T) {
	logger := telemetry.NewLogger(testWriter{t}, 0, false)
	srv, _ := grpcapi.NewGRPCServer(config.GRPCConfig{
		MaxRecvMsgBytes: 1 << 20,
		HandlerTimeout:  5 * time.Second,
		ConnTimeout:     5 * time.Second,
	}, grpcapi.Deps{
		Service: panicSvc{},
		Logger:  logger,
		Metrics: telemetry.NewPromMetrics("test"),
		Authn:   grpcapi.NewSyntheticAuthenticator(writerPrincipal()),
	})

	lis := bufconn.Listen(bufSize)
	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("serve stopped: %v", err)
		}
	}()
	t.Cleanup(srv.Stop)

	//nolint:staticcheck // documented in-process bufconn dial pattern.
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := widgetv1.NewWidgetServiceClient(conn)
	_, err = client.GetWidget(context.Background(), &widgetv1.GetWidgetRequest{Id: "x"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("code = %v, want Internal (recovered panic)", status.Code(err))
	}
	// The internal message must not leak the panic value.
	if msg := status.Convert(err).Message(); msg == "boom" {
		t.Errorf("internal panic value leaked to client: %q", msg)
	}
}

func TestStaticTokenAuthenticator(t *testing.T) {
	authn := grpcapi.NewStaticTokenAuthenticator("tok", writerPrincipal())

	if _, err := authn.Authenticate(context.Background(), "tok"); err != nil {
		t.Errorf("valid token rejected: %v", err)
	}
	if _, err := authn.Authenticate(context.Background(), "wrong"); err == nil {
		t.Error("wrong token accepted")
	}
	if _, err := authn.Authenticate(context.Background(), ""); err == nil {
		t.Error("empty token accepted")
	}
}
