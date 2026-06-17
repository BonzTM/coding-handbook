package grpc

import (
	"log/slog"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	widgetv1 "github.com/example/examplegrpc/api/widget/v1"
	"github.com/example/examplegrpc/internal/config"
)

// Deps are the collaborators NewGRPCServer needs beyond config. Defining them in
// one struct keeps the constructor signature stable as the service grows.
type Deps struct {
	// Service implements the widgets RPCs.
	Service widgetSvc
	// Logger is threaded into the recovery and access-log interceptors.
	Logger *slog.Logger
	// Metrics records per-RPC counters/latency exposed via the HTTP sidecar.
	Metrics rpcMetrics
	// Authn verifies bearer tokens; SyntheticAuthenticator in local/dev,
	// StaticTokenAuthenticator (or a JWKS verifier) when auth is enabled.
	Authn Authenticator
}

// NewGRPCServer builds the configured *grpc.Server with the interceptor chain,
// the otelgrpc stats handler for tracing/propagation, the standard health
// service, and server reflection. The returned *health.Server lets main flip
// the service to NOT_SERVING first in the ordered shutdown so load balancers
// stop routing before GracefulStop drains in-flight RPCs.
func NewGRPCServer(cfg config.GRPCConfig, deps Deps) (*grpc.Server, *health.Server) {
	// The chain runs recovery (outermost) -> request-id -> access log ->
	// deadline guard -> auth (innermost). otelgrpc runs as a stats handler, so
	// the span context is in ctx before any interceptor logs.
	chain := chainUnary(
		recoveryUnary(deps.Logger),
		requestIDUnary(),
		accessLogUnary(deps.Logger, deps.Metrics),
		deadlineGuardUnary(cfg.HandlerTimeout),
		authUnary(deps.Authn),
	)

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(chain),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgBytes),
		grpc.ConnectionTimeout(cfg.ConnTimeout),
	)

	widgetv1.RegisterWidgetServiceServer(srv, NewServer(deps.Service))

	// Standard health service for deployability. main sets per-service and
	// overall status; it starts SERVING and is flipped to NOT_SERVING on shutdown.
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus(widgetv1.WidgetService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	// Reflection so grpcurl and other tooling can introspect the service.
	reflection.Register(srv)

	return srv, healthSrv
}
