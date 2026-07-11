// Program.cs - the service entrypoint and composition root, adapted from
// csharp/templates/program-main.cs.txt. It does nothing but wire the process:
// this is the ONLY file in Orders.Api that touches Orders.Infrastructure;
// everything else depends on Orders.Core interfaces. Startup fails fast (bad
// config aborts before listeners open); shutdown is bounded and drains
// in-flight requests.

using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.AspNetCore.Http.Timeouts;

using Orders.Api;
using Orders.Api.Auth;
using Orders.Api.Contracts;
using Orders.Api.Endpoints;
using Orders.Api.ErrorHandling;
using Orders.Api.Middleware;
using Orders.Api.RateLimiting;
using Orders.Api.Telemetry;
using Orders.Infrastructure;

var builder = WebApplication.CreateBuilder(args);

// Configuration precedence (last wins): appsettings.json (committed, no
// secrets) -> appsettings.{Environment}.json -> user secrets (Development) ->
// environment variables -> command line. Section:Key maps to Section__Key in
// the environment (csharp/foundations/configuration.md).

// Options with fail-fast validation: a bad value kills the process at deploy
// time, not on the first request at 3am.
builder.Services.AddOptions<HttpServerOptions>()
    .BindConfiguration(HttpServerOptions.SectionName)
    .ValidateDataAnnotations()
    .ValidateOnStart();

// Time is an input: one system TimeProvider at the root, injected everywhere,
// FakeTimeProvider in tests (csharp/foundations/time.md).
builder.Services.AddSingleton(TimeProvider.System);

// Telemetry is one extension - structured logging, OTel traces/metrics/logs
// over OTLP, base health-check registration (csharp/operations/observability.md).
builder.AddServiceTelemetry();

// RFC 9457 ProblemDetails is the ONLY error shape on the wire. requestId joins
// every problem body to logs and traces; domain exceptions map to statuses in
// one handler (csharp/foundations/errors-and-logging.md).
builder.Services.AddProblemDetails(options =>
    options.CustomizeProblemDetails = context =>
        context.ProblemDetails.Extensions["requestId"] = context.HttpContext.TraceIdentifier);
builder.Services.AddExceptionHandler<DomainExceptionHandler>();

// Built-in minimal-API validation: DataAnnotations on the wire DTOs run before
// handlers and fail as 400 ValidationProblemDetails.
builder.Services.AddValidation();

// The wire surface serializes through the source-generated context - a closed,
// reviewable set of types (csharp/foundations/serialization.md).
builder.Services.ConfigureHttpJsonOptions(options =>
    options.SerializerOptions.TypeInfoResolverChain.Insert(0, OrdersJsonContext.Default));

// AuthN (JWT/JWKS, config-gated local-dev mode) + deny-by-default authZ.
builder.AddOrdersSecurity();

// Per-client rate limiting: 429 + Retry-After, never the 503 default.
builder.AddOrdersRateLimiting();

// Server hardening: request timeout policy + body-size cap from config.
var server = builder.Configuration.GetSection(HttpServerOptions.SectionName).Get<HttpServerOptions>()
    ?? new HttpServerOptions();
builder.Services.AddRequestTimeouts(options =>
    options.DefaultPolicy = new RequestTimeoutPolicy { Timeout = server.RequestTimeout });
builder.WebHost.ConfigureKestrel(kestrel =>
    kestrel.Limits.MaxRequestBodySize = server.MaxRequestBodyBytes);

// Infrastructure composition: pooled DbContext (Npgsql), repositories,
// idempotency runner, domain service, and the "ready" database health check.
// This one call is the Api -> Infrastructure boundary crossing.
builder.Services.AddOrdersInfrastructure(builder.Configuration);

// Bounded shutdown: keep the drain budget below the orchestrator's grace
// period (k8s terminationGracePeriodSeconds default 30s).
builder.Services.Configure<HostOptions>(options =>
    options.ShutdownTimeout = TimeSpan.FromSeconds(15));

var app = builder.Build();

// Middleware order is a contract (csharp/services/http-services.md): request
// id outermost so even an unhandled-exception 500 carries requestId; authN
// before rate limiting so partitions key on identity; timeouts innermost.
app.UseMiddleware<RequestIdMiddleware>();
app.UseExceptionHandler();
app.UseStatusCodePages();
app.UseRouting();
app.UseAuthentication();
app.UseAuthorization();
app.UseRateLimiter();
app.UseRequestTimeouts();

// Probes (csharp/operations/observability.md): /livez runs no checks - a dead
// database must not restart the pod. /readyz runs the "ready"-tagged checks
// (the database) and sheds traffic instead. Anonymous on purpose: the
// deny-by-default fallback policy would otherwise 401 the kubelet.
app.MapHealthChecks("/livez", new HealthCheckOptions { Predicate = _ => false })
    .AllowAnonymous();
app.MapHealthChecks("/readyz", new HealthCheckOptions
{
    Predicate = registration => registration.Tags.Contains("ready"),
}).AllowAnonymous();

// Endpoint groups - one Map*Endpoints extension per resource.
app.MapOrderEndpoints();

// Explicit migration step: `dotnet Orders.Api.dll --migrate` applies pending
// migrations and exits. Never a side effect of normal startup
// (csharp/services/database.md).
if (args.Contains("--migrate", StringComparer.Ordinal))
{
    await app.MigrateOrdersDatabaseAsync();
    return;
}

await app.RunAsync();

// Expose the entry point to WebApplicationFactory<Program> in tests
// (top-level statements generate an internal Program otherwise).
// CA1515 wants every type in an application assembly internal; Program is the
// ONE deliberate exception - the xUnit test classes hosting
// WebApplicationFactory<Program> must be public, so their fixture type (and
// therefore Program) must be too.
#pragma warning disable CA1515 // Consider making public types internal
public partial class Program;
#pragma warning restore CA1515
