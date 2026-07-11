// Program.cs - the worker entrypoint and composition root, adapted from
// csharp/templates/program-main.cs.txt. It does nothing but wire the process:
// this is the ONLY file in Orders.Worker that touches
// Orders.Worker.Infrastructure; the delivery pipeline depends on
// Orders.Worker.Core ports. Startup fails fast (bad config aborts before the
// consumer subscribes); shutdown is a bounded graceful drain.

using Microsoft.AspNetCore.Diagnostics.HealthChecks;

using Orders.Worker.Infrastructure;
using Orders.Worker.Infrastructure.Health;
using Orders.Worker.Telemetry;

var builder = WebApplication.CreateBuilder(args);

// Configuration precedence (last wins): appsettings.json (committed, no
// secrets) -> appsettings.{Environment}.json -> user secrets (Development) ->
// environment variables -> command line. Section:Key maps to Section__Key in
// the environment (csharp/foundations/configuration.md).

// Time is an input: one system TimeProvider at the root, injected everywhere,
// FakeTimeProvider in tests (csharp/foundations/time.md).
builder.Services.AddSingleton(TimeProvider.System);

// Telemetry is one extension - structured logging, OTel traces/metrics/logs
// over OTLP, base health-check registration (csharp/operations/observability.md).
builder.AddServiceTelemetry();

// Messaging composition: the in-memory broker behind the Core ports, the
// inbox/outbox/DLQ stores, the consumer + outbox-relay BackgroundServices,
// metrics, and the "ready" broker health check. This one call is the host ->
// Infrastructure boundary crossing.
builder.Services.AddOrdersWorkerMessaging(builder.Configuration);

// Bounded shutdown: StopAsync cancels the consume loop's stoppingToken, then
// waits at most this long for in-flight messages to settle and the outbox's
// final flush. Keep the budget below the orchestrator's grace period (k8s
// terminationGracePeriodSeconds default 30s).
builder.Services.Configure<HostOptions>(options =>
    options.ShutdownTimeout = TimeSpan.FromSeconds(15));

var app = builder.Build();

// The worker serves probes ONLY - no application routes, no auth stack, no
// middleware pipeline beyond what the probes need (the counterpart of the Go
// reference's bare probe sidecar). /livez runs no checks: a dead broker must
// not restart the pod. /readyz runs the "ready"-tagged broker check and sheds
// work instead (csharp/operations/observability.md).
app.MapHealthChecks("/livez", new HealthCheckOptions { Predicate = _ => false });
app.MapHealthChecks("/readyz", new HealthCheckOptions
{
    Predicate = registration => registration.Tags.Contains("ready"),
});

// Flip readiness off the moment shutdown begins - before the drain - so the
// platform stops counting on this replica while in-flight work finishes.
var readiness = app.Services.GetRequiredService<WorkerReadiness>();
app.Lifetime.ApplicationStopping.Register(() => readiness.SetReady(false));

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
