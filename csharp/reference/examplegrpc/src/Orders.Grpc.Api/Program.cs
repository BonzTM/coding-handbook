// Program.cs - the service entrypoint and composition root, adapted from
// csharp/templates/program-main.cs.txt and the keystone exampleservice module.
// It does nothing but wire the process. Startup fails fast (bad config or an
// unloadable TLS key pair aborts before listeners open); shutdown is bounded
// and drains in-flight calls. There is no --migrate mode: this module has no
// database by design (see README).

using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.Extensions.Options;

using Orders.Grpc.Api;
using Orders.Grpc.Api.Auth;
using Orders.Grpc.Api.Grpc;
using Orders.Grpc.Api.Grpc.Interceptors;
using Orders.Grpc.Api.Telemetry;
using Orders.Grpc.Api.Tls;
using Orders.Grpc.Core.Orders;

var builder = WebApplication.CreateBuilder(args);

// Configuration precedence (last wins): appsettings.json (committed, no
// secrets) -> appsettings.{Environment}.json -> user secrets (Development) ->
// environment variables -> command line. Section:Key maps to Section__Key in
// the environment (csharp/foundations/configuration.md).

// Options with fail-fast validation: a bad value kills the process at deploy
// time, not on the first call at 3am.
builder.Services.AddOptions<ServerOptions>()
    .BindConfiguration(ServerOptions.SectionName)
    .ValidateDataAnnotations()
    .ValidateOnStart();
builder.Services.AddOptions<AuthOptions>()
    .BindConfiguration(AuthOptions.SectionName)
    .ValidateDataAnnotations()
    .ValidateOnStart();
builder.Services.AddSingleton<IValidateOptions<AuthOptions>, AuthOptionsValidator>();
builder.Services.AddOptions<TlsOptions>()
    .BindConfiguration(TlsOptions.SectionName)
    .ValidateDataAnnotations()
    .ValidateOnStart();
builder.Services.AddSingleton<IValidateOptions<TlsOptions>, TlsOptionsValidator>();

// Time is an input: one system TimeProvider at the root, injected everywhere,
// FakeTimeProvider in tests (csharp/foundations/time.md).
builder.Services.AddSingleton(TimeProvider.System);

// Telemetry is one extension - structured logging, OTel traces/metrics/logs
// over OTLP, base health-check registration (csharp/operations/observability.md).
builder.AddServiceTelemetry();

// The domain and its in-memory store (this module's honest scope - see
// README; a database-backed service composes an Infrastructure project here
// instead, exactly like the keystone module).
builder.Services.AddSingleton<IOrderStore, InMemoryOrderStore>();
builder.Services.AddSingleton<OrderService>();

// Auth seam, config-gated exactly like the keystone: enabled wires the
// static-token authenticator (production swaps a JWT/JWKS validator behind
// the same interface); disabled wires the synthetic local-dev principal.
var auth = builder.Configuration.GetSection(AuthOptions.SectionName).Get<AuthOptions>() ?? new AuthOptions();
if (auth.Enabled)
{
    builder.Services.AddSingleton<IAuthenticator, StaticTokenAuthenticator>();
}
else
{
    builder.Services.AddSingleton<IAuthenticator, LocalDevAuthenticator>();
}

// The interceptor chain. Registration order = execution order; the pinned
// contract is logging (outermost, observes the final mapped status) -> auth
// (rejects before any handler work) -> exception mapping (innermost, nothing
// unmapped escapes). Registered globally so every service - including ones
// added later - is authenticated by default; health and reflection are the
// interceptor's two documented exemptions (csharp/services/grpc-services.md).
builder.Services.AddSingleton<RequestLoggingInterceptor>();
builder.Services.AddSingleton<AuthInterceptor>();
builder.Services.AddSingleton<ExceptionMappingInterceptor>();
builder.Services.AddGrpc(options =>
{
    options.Interceptors.Add<RequestLoggingInterceptor>();
    options.Interceptors.Add<AuthInterceptor>();
    options.Interceptors.Add<ExceptionMappingInterceptor>();
});

// grpc.health.v1 over the SAME health-check registrations as /livez//readyz;
// reflection so grpcurl works without local protos (config-gated - internal
// services only).
builder.Services.AddGrpcHealthChecks();
var server = builder.Configuration.GetSection(ServerOptions.SectionName).Get<ServerOptions>() ?? new ServerOptions();
if (server.EnableReflection)
{
    builder.Services.AddGrpcReflection();
}

// Two listeners (gRPC HTTP/2, probes HTTP/1.1) with config-gated TLS/mTLS.
// An unloadable configured key pair throws at bind time and aborts startup -
// never a silent downgrade to plaintext.
var tls = builder.Configuration.GetSection(TlsOptions.SectionName).Get<TlsOptions>() ?? new TlsOptions();
builder.WebHost.ConfigureKestrel(kestrel => kestrel.ConfigureOrdersListeners(server, tls));

// Bounded shutdown: keep the drain budget below the orchestrator's grace
// period (k8s terminationGracePeriodSeconds default 30s). Kestrel drains
// in-flight RPCs; streaming calls observe cancellation and stop.
builder.Services.Configure<HostOptions>(options =>
    options.ShutdownTimeout = TimeSpan.FromSeconds(15));

var app = builder.Build();

if (!tls.HasServerTls)
{
    // The Go reference logs the same loud warning: plaintext is for local/dev
    // only; production requires TLS (csharp/services/grpc-services.md).
    StartupLog.PlaintextListener(app.Logger, server.GrpcPort);
}

// The gRPC surface.
app.MapGrpcService<OrdersGrpcService>();
app.MapGrpcHealthChecksService();
if (server.EnableReflection)
{
    app.MapGrpcReflectionService();
}

// HTTP probes on the probes listener (csharp/operations/observability.md):
// /livez runs no checks - a dead dependency must not restart the pod; /readyz
// runs the "ready"-tagged checks (none in this module - no dependencies) and
// sheds traffic instead.
app.MapHealthChecks("/livez", new HealthCheckOptions { Predicate = _ => false });
app.MapHealthChecks("/readyz", new HealthCheckOptions
{
    Predicate = registration => registration.Tags.Contains("ready"),
});

await app.RunAsync();

/// <summary>Source-generated startup log messages (CA1848 - csharp/quality/linting.md).</summary>
internal static partial class StartupLog
{
    [LoggerMessage(Level = LogLevel.Warning,
        Message = "serving PLAINTEXT gRPC on port {GrpcPort} - local/dev only; production requires TLS (set Tls:CertPath/Tls:KeyPath)")]
    public static partial void PlaintextListener(ILogger logger, int grpcPort);
}

// Expose the entry point to WebApplicationFactory<Program> in tests
// (top-level statements generate an internal Program otherwise).
// CA1515 wants every type in an application assembly internal; Program is the
// ONE deliberate exception - the xUnit test classes hosting
// WebApplicationFactory<Program> must be public, so their fixture type (and
// therefore Program) must be too.
#pragma warning disable CA1515 // Consider making public types internal
public partial class Program;
#pragma warning restore CA1515
