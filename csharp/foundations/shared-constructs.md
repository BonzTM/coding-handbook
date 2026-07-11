# Shared Constructs

Recommended reusable building blocks for mature .NET solutions. These exist to remove repeated wiring without creating a junk drawer.

## Default Approach

Prefer a small number of explicit, named homes that solve common cross-cutting needs well. These are the C# counterparts of a Go repo's `internal/*` shared packages — projects and folders inside the app solution, not extra abstraction layers.

| Construct | Owns | Do not turn it into |
|---|---|---|
| `<App>.Core` | domain types, ports (interfaces), business rules | a layer that knows about HTTP, EF Core, or the broker |
| `<App>.Api/Endpoints/` | minimal-API endpoint groups: route registration, request/response mapping, status mapping | business logic hiding in handlers — a second Core |
| `<App>.Infrastructure/Data/` | `DbContext`, migrations, repositories implementing Core ports | a query layer leaking `IQueryable` or entities to callers, or growing business rules |
| `<App>.Infrastructure/Clients/` | typed outbound HTTP clients (`IHttpClientFactory` + typed client + resilience handler) | a wrapper around every external SDK call, or a zoo of per-call retry policies |
| `<App>.Api/Telemetry/` — `AddServiceTelemetry()` | logging, OpenTelemetry, and health wiring | a wrapper around every `ILogger` or metrics call |
| `<App>.TestUtilities` (only when ≥2 test projects need it) | fakes, builders, `FakeTimeProvider` helpers | hidden production logic or giant assertion DSLs |

For the Orders service these are `Orders.Core`, `Orders.Api/Endpoints/`, `Orders.Infrastructure/Data/`, `Orders.Infrastructure/Clients/`, `Orders.Api/Telemetry/`, and `Orders.TestUtilities`. `AddServiceTelemetry()` may graduate into a shared `ServiceDefaults`-style extension when several services in one repo repeat it — that is the ceiling of sharing, not the start. `<App>.TestUtilities` does not exist until a second test project actually duplicates helpers; until then the helpers live in the one test project that uses them.

Boundary and reference rules between these constructs are in [solution-and-project-design.md](solution-and-project-design.md); the tree they live in is in [project-setup.md](project-setup.md).

### Constructor Pattern

- Constructors make dependencies explicit; registration happens in per-project extension methods (`AddOrdersCore()`, `AddOrdersInfrastructure()`, `AddServiceTelemetry()`) so `Program.cs` stays a readable wiring manifest.
- Long constructor signatures are acceptable when they reveal real dependencies; assembly-scanning registration or an ambient container is not inherently cleaner — it is invisible.
- Group dependencies in a small class only when they naturally belong together; tuning knobs go through validated options, not extra parameters (see [configuration.md](configuration.md)).

### Shutdown Pattern

- The host owns process lifetime: `app.Run()` wires SIGTERM and Ctrl+C through `IHostApplicationLifetime`. Never hand-roll signal handling (see [cross-platform.md](cross-platform.md)).
- Bound the drain with `HostOptions.ShutdownTimeout`; Kestrel drains in-flight requests within it.
- Workers honor `stoppingToken` in `BackgroundService.ExecuteAsync` and exit promptly when it fires (see [cancellation-and-async.md](cancellation-and-async.md)).
- Register disposable resources with the DI container so they are disposed deterministically at shutdown, instead of ad hoc cleanup paths.

## Common Mistakes And Forbidden Patterns

- An `Orders.Common`, `Orders.Shared`, or `Orders.Utils` project that gradually becomes the real architecture.
- Assembly-scanning or convention-based registration added before manual `Add*` wiring is actually painful.
- Hidden dependencies through ambient state: statics, service locator, `HttpContext` reached from Core, or `DateTime.Now` instead of an injected `TimeProvider` (see [time.md](time.md)).
- A `TestUtilities` project created for a single test project, or test helpers that assert too much and make failures harder to read.
- `AddServiceTelemetry()` absorbing business middleware, auth policy, or anything beyond logging/OTel/health wiring.

## Verification And Proof

- A new contributor should be able to read `Program.cs` and understand the dependency graph.
- Shared helpers should reduce duplication without making call sites mysterious.
- If a shared construct cannot state exactly what it owns, it is probably too broad.
