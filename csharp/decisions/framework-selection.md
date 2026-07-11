# Framework Selection

Rules for deciding when a dependency earns its complexity cost.

## Default Approach

Start with the .NET base class library and the `Microsoft.Extensions.*` platform packages, and add third-party NuGet packages only when they clearly improve correctness, interoperability, or operator experience. The runtime and ASP.NET Core cover far more ground than most ecosystems' standard libraries — the burden of proof is on the package, not the platform.

### Approval Questions

Before adding a dependency, answer all of these:

1. What concrete problem does the BCL, ASP.NET Core, or the current stack fail to solve well enough?
2. What maintenance, upgrade, licensing, and security cost does this add?
3. Does the package introduce hidden magic (reflection scanning, assembly weaving, static state), or framework lock-in?
4. Is it widely used, actively maintained, license-compatible, and easy to replace later if needed?

## Default Choices By Concern

| Concern | Default | Acceptable escalation | Avoid by default |
|---|---|---|---|
| runtime | current LTS release only; SDK pinned by `global.json` — copy from [../templates/global.json](../templates/global.json) (see [../foundations/project-setup.md](../foundations/project-setup.md)) | an STS release via ADR when a feature is genuinely blocking | floating SDK versions; builds that pass only on one machine's SDK |
| language posture | latest C# for the pinned SDK; `<Nullable>enable</Nullable>`, `<ImplicitUsings>enable</ImplicitUsings>`, file-scoped namespaces — set once in [../templates/Directory.Build.props](../templates/Directory.Build.props) | — | disabling nullable in new code; per-project language drift |
| solution format | `.slnx` (XML solution format) | `.sln` only for tooling that cannot read slnx | maintaining both formats in parallel |
| solution layout | `src/` + `tests/`; a service is `Orders.Api` (thin host) + `Orders.Core` (domain, no ASP.NET/EF references) + `Orders.Infrastructure` (EF Core, external clients) — see [../foundations/solution-and-project-design.md](../foundations/solution-and-project-design.md) | fewer projects for tiny tools, noted in the project README | a single project mixing transport, domain, and data access; a `Common` dumping-ground project |
| boundary enforcement | project references define the allowed direction: Core references nothing app-level; Api and Infrastructure reference Core; Api references Infrastructure only in the `Program.cs` composition root | `InternalsVisibleTo` only for the matching test project | architecture-test frameworks papering over reference cycles; reflection across project boundaries |
| web framework | ASP.NET Core Minimal APIs with `TypedResults`, route groups, and endpoint filters (see [../services/http-services.md](../services/http-services.md)) | MVC controllers via ADR when filter pipelines or model-binding volume justify them | third-party web frameworks layered over ASP.NET Core; hiding endpoints behind bespoke routing abstractions |
| server-rendered web | Razor Pages on the same host (see [../services/web-apps.md](../services/web-apps.md)) | Blazor via ADR when interactivity genuinely outgrows pages plus progressive enhancement | a SPA by reflex for form-over-data apps |
| HTML templating | Razor (`.cshtml` pages, partials, tag helpers) with automatic HTML encoding | a Razor class library when views are shared across hosts | string-built HTML; `HtmlString`/`Html.Raw` around unencoded input |
| gRPC | `Grpc.AspNetCore`, contract-first from `.proto` files in `api/` (see [../services/grpc-services.md](../services/grpc-services.md)) | — | hand-rolled RPC; gateway/proxy sprawl before it is needed |
| request validation | .NET built-in minimal-API validation (`AddValidation` + DataAnnotations on request DTOs) at the trust boundary (see [../foundations/serialization.md](../foundations/serialization.md), [../foundations/data-modeling.md](../foundations/data-modeling.md)) | FluentValidation via ADR when rule complexity genuinely outgrows attributes | validation deferred until raw payloads have crossed into the domain; duplicate validation stacks |
| CORS | none — service-to-service APIs need no CORS layer (see [../services/http-services.md](../services/http-services.md) ### CORS) | built-in CORS middleware (`AddCors` with a named policy and an explicit origin allowlist) when spec intake identifies browser clients | `AllowAnyOrigin` combined with credentials; policies that reflect arbitrary `Origin` values |
| sessions | ASP.NET Core cookie authentication with server-validated tickets (see [../services/web-apps.md](../services/web-apps.md)) | a server-side ticket store backed by a store the repo already runs (Postgres/Redis) when revocation or ticket size demands it (ADR-level) | hand-rolled session/cookie crypto; JWTs in cookies as a session substitute |
| CSRF | built-in antiforgery — automatic in Razor Pages; `AddAntiforgery` plus per-endpoint antiforgery for minimal-API form posts (see [../services/web-apps.md](../services/web-apps.md)) | — | disabling antiforgery app-wide to accommodate one endpoint; CSRF-exempting state-changing GETs instead of removing them |
| data access | EF Core with one `DbContext` per service, Npgsql/PostgreSQL provider (see [../services/database.md](../services/database.md)) | Dapper for measured hot paths via ADR; SQL Server when the org mandates it | repository frameworks wrapped around `DbContext`; SQL built by string concatenation |
| schema migrations | EF Core migrations, applied by an explicit migration step (`--migrate` flag or init job), never on normal startup (see [../services/database.md](../services/database.md)) | — | `EnsureCreated` or auto-migrate-on-start in production; hand-applied SQL with no migration history |
| config loading | `Microsoft.Extensions.Configuration`: committed `appsettings.json` (no secrets) with env-var overrides winning; options pattern with `ValidateDataAnnotations` + `ValidateOnStart` (see [../foundations/configuration.md](../foundations/configuration.md)) | extra config files for operator workflows via ADR | raw `IConfiguration` indexer reads scattered through business logic; global config frameworks with implicit precedence |
| dependency injection | built-in `Microsoft.Extensions.DependencyInjection` with constructor injection | a third-party container via ADR (see ## Mandated Frameworks) | service locator; property injection; container-specific attributes in Core |
| logging | `ILogger<T>` with source-generated `[LoggerMessage]` and structured templates; log once at the boundary that can act (see [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)) | Serilog via ADR only when its sinks are required — still consumed through `ILogger<T>` | bespoke logging abstractions; `Console.WriteLine`; string-interpolated log messages |
| metrics / tracing | OpenTelemetry over `System.Diagnostics` `Meter`/`ActivitySource`, OTLP exporter (see [../operations/observability.md](../operations/observability.md)) | the OTel Prometheus exporter when the org scrapes; org-mandated backends behind the OTel API | vendor SDK calls in Core; ad hoc trace systems |
| health | `Microsoft.Extensions.Diagnostics.HealthChecks` with `/livez` (always cheap) and `/readyz` (dependency checks) — see [../operations/operability.md](../operations/operability.md) | — | liveness probes that touch dependencies; health endpoints that fan out on every scrape without caching |
| serialization | `System.Text.Json` with source generators (`JsonSerializerContext`) at trust boundaries; strict number handling; explicit unknown-field policy (see [../foundations/serialization.md](../foundations/serialization.md)) | — | Newtonsoft.Json in new code (forbidden below); two serializers with divergent date/null/unknown-field behavior in one service |
| money / decimal | `decimal` in domain types with an explicit currency code; integer minor units on wire and storage contracts where exactness must survive other consumers (see [../foundations/data-modeling.md](../foundations/data-modeling.md)) | a dedicated money library (ADR-level) only for genuine multi-currency arithmetic such as allocation or compounding rules | `double`/`float` for money anywhere; culture-implicit `Parse`/`ToString` on amounts at trust boundaries |
| time | injected `TimeProvider`; `DateTimeOffset`/`DateOnly`/`TimeOnly` for storage and wire; `FakeTimeProvider` in tests (see [../foundations/time.md](../foundations/time.md)) | — | `DateTime.Now`/`DateTime.UtcNow` read directly in logic; Windows time-zone IDs on the wire |
| retries / circuit breakers | `Microsoft.Extensions.Http.Resilience` standard resilience handler on every typed outbound `HttpClient`; hand-rolled bounded exponential backoff with full jitter behind an injected `TimeProvider` for non-HTTP work (see [../operations/resilience.md](../operations/resilience.md)) | a custom Polly pipeline (ADR-level) when policy complexity genuinely outgrows the standard handler; the pipeline must expose attempt and breaker-state metrics | unbounded or zero-jitter retry loops; a breaker on every client by reflex; resilience wrappers that hide attempt state |
| messaging | the repo-owned seam — publisher/consumer interfaces in Core plus inbox/outbox in Infrastructure; broker client (NATS JetStream or RabbitMQ when the spec is silent) confined behind the seam (see [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)) | MassTransit or Wolverine via ADR after idempotency, ordering, retry, and DLQ expectations are documented | frameworks that obscure ack, retry, DLQ, or partition behavior; broker types leaking into Core |
| in-process caching | `HybridCache` with explicit entry options — bounded size and TTL; its stampede protection collapses duplicate loads (see [../services/caching.md](../services/caching.md)) | Redis/Valkey as the distributed backend only when the working set or cross-instance sharing demands it | unbounded `MemoryCache` or static dictionaries used as caches |
| feature flags | static typed flags on the options pattern in [../foundations/configuration.md](../foundations/configuration.md) — plain config-backed booleans read through a typed accessor | `Microsoft.FeatureManagement` via ADR when percentage rollout or audience targeting is genuinely needed | a managed flag/experimentation service before targeting needs it; scattered raw config lookups; long-lived flags left as debt |
| job scheduling | `PeriodicTimer` inside a `BackgroundService` that honors `stoppingToken` and guards against overlapping runs | Quartz.NET via ADR for calendar/cron schedules; a distributed lock or leader election for multi-replica singletons | fire-and-forget `Task.Delay` loops with no overlap guard; unobserved background tasks |
| API deprecation signaling | `[Obsolete]` on the shipped contract surface plus a `Sunset` header (RFC 8594) and a documented `Deprecation` header form, recorded in an ADR (see [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md)) | an org-standard deprecation registry or policy | removing a contract with no deprecation signal or window |
| CLI parsing | `System.CommandLine` (GA `SetAction` API) | — | hand-rolled `args` parsing beyond a single flag; abandoned CLI frameworks |
| testing | xUnit v3 on Microsoft.Testing.Platform with plain `Assert`; `WebApplicationFactory` for in-proc HTTP tests; Testcontainers for DB/broker integration tests (see [../quality/testing.md](../quality/testing.md)) | — | Moq and FluentAssertions (forbidden below); assertion DSLs that obscure behavior |
| test doubles | hand-rolled fakes at Core-owned interfaces (see [../quality/testing.md](../quality/testing.md) ### Test Doubles) | NSubstitute when a fake is too costly to write or maintain | Moq (forbidden below); over-specified call expectations that pin implementation details |
| benchmark comparison | BenchmarkDotNet with repeated runs and exported artifacts | — | `Stopwatch` one-offs or single-run before/after deltas presented as proof |
| analyzers / format | built-in .NET analyzers at `AnalysisLevel=latest-all`, `TreatWarningsAsErrors=true`, `.editorconfig` as the single severity source, `dotnet format --verify-no-changes` as the gate (see [../quality/linting.md](../quality/linting.md)) | extra analyzer packages via ADR | blanket suppressions without justification; per-project severity drift |
| package hygiene | Central Package Management (`Directory.Packages.props`), `packages.lock.json` with locked-mode restore in CI, NuGetAudit failing on high/critical (see [../foundations/project-setup.md](../foundations/project-setup.md)) | — | per-`.csproj` floating versions; wildcard version ranges; unreviewed lock-file churn |
| build entrypoint | `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit; `Makefile` is a one-line shim (see [../foundations/project-setup.md](../foundations/project-setup.md)) | — | bash-only scripts in the verify path; CI steps that diverge from the script |
| containers | multi-stage Dockerfile: SDK image builds, chiseled ASP.NET runtime image runs, non-root — copy from [../templates/Dockerfile](../templates/Dockerfile) (see [../operations/deployment.md](../operations/deployment.md)) | Native AOT or ReadyToRun via ADR | shipping the SDK image to production; running as root; `latest` tags |
| shutdown | `IHostApplicationLifetime` + `HostOptions.ShutdownTimeout`; Kestrel drains in-flight requests; workers honor `stoppingToken` (see [../operations/operability.md](../operations/operability.md)) | — | `Environment.Exit` in app code; workers that ignore cancellation and die mid-write |
| secrets manager | env vars or mounted files injected by an external manager; `dotnet user-secrets` for local dev only (see [../operations/security.md](../operations/security.md)) | the org-provided manager (Key Vault, Vault, cloud secrets manager) behind a seam in Infrastructure | secrets in `appsettings.json`, source, images, or build args; the app fetching and caching long-lived plaintext itself |
| audit / log sink | structured `ILogger` events on a dedicated audit category routed to a stream the platform collects (see [../operations/security.md](../operations/security.md)) | a SIEM, managed audit service, or append-only store behind an Infrastructure seam when compliance requires tamper-evidence | mixing audit events into the shared application log; no retention or access control on the sink |
| object mapping | map by hand — explicit constructors, factory methods, or extension methods per boundary (see [../foundations/data-modeling.md](../foundations/data-modeling.md)) | Mapperly via ADR when mapping volume is real: it is source-generated, so every mapping is inspectable, compile-time-checked C# | AutoMapper (forbidden below); mapping by reflection at trust boundaries |
| in-process dispatch | plain DI — interfaces in Core, constructor injection, direct calls | — | MediatR for DI-solvable dispatch (forbidden below); message buses simulated inside one process |

## Mandated Frameworks

Sometimes the spec or the requester mandates a framework this table would not choose — MVC controllers everywhere, Autofac, Serilog, an org-standard stack. A mandate is honored, not fought, and not silently absorbed:

- Record an ADR stating the framework was **mandated by the requester**, which solution area may depend on it, and what the handbook default would have been. The Approval Questions still get written answers; "mandated" answers question 1.
- The framework-independent invariants survive unchanged: the `src/`+`tests/` layout with Core referencing nothing app-level, a thin `Program.cs` composition root, `CancellationToken` flowing through every async path, `ILogger<T>` as the logging surface (through an adapter or provider if the framework insists on its own logger — mandated Serilog is still consumed as an `ILogger<T>` provider), the RFC 9457 ProblemDetails error contract, the testing bar, and the full `pwsh ./verify.ps1` gate.
- Confine framework types to the host and infrastructure projects (`Orders.Api`, `Orders.Infrastructure`). Endpoints or controllers translate framework types into plain arguments for `Orders.Core`; Core never references the framework. This keeps the mandate reversible and the domain testable without it.
- Flag the cost of mandates that shift the handbook's guidance explicitly in the ADR. Mandated MVC controllers replace the minimal-API surface: endpoint filters, `TypedResults`, and route-group guidance in [../services/http-services.md](../services/http-services.md) apply only in spirit, and their controller equivalents (action filters, `ActionResult<T>`, model binding) must be named in the ADR. Mandated Autofac changes registration and lifetime semantics: registrations stay in `IServiceCollection`-shaped extension methods where possible, and no Autofac attribute or module type may appear in Core. That cost belongs in writing before code starts.
- A mandate covers the named framework only — it is not a license to relax the rest of this table.

## Common Mistakes And Forbidden Patterns

Forbidden by default — each requires an ADR that answers why the stated reason does not apply:

- **Newtonsoft.Json** — `System.Text.Json` with source generation is the trust-boundary default; a second serializer forks date, null, and unknown-field behavior inside one service and silently bypasses the source-generated contracts.
- **Moq** — the SponsorLink incident demonstrated the project will ship surprising behavior in a patch release; NSubstitute covers the rare case a hand-rolled fake cannot.
- **FluentAssertions** — moved to a paid commercial license; xUnit's `Assert` is sufficient and keeps the dependency and license surface flat.
- **AutoMapper** — moved to a commercial license, and reflection-based mapping hides field drift until runtime. Map by hand; adopt Mapperly via ADR when volume justifies it, because its source-generated mappings are visible, reviewable C# checked at compile time.
- **MediatR for simple DI-solvable dispatch** — moved to a commercial license, and for plain request→handler wiring it is an indirection layer that hides the call graph from readers, analyzers, and refactoring tools. A constructor-injected interface does the same job traceably.

And in general:

- No floating or wildcard package versions; every version lives in `Directory.Packages.props` and `packages.lock.json`, with no per-project overrides.
- No dependency added only because it is familiar from another language ecosystem.
- No repository framework, third-party DI container, or web micro-framework just to avoid writing explicit ASP.NET Core code.
- No tool dependency in runtime packages when it belongs in the tool manifest (`dotnet tool restore` from `.config/dotnet-tools.json`).
- No messaging library adopted before the repo documents idempotency, ordering, retry, and DLQ expectations.
- No dependency added without the Approval Questions answered in writing and the `Directory.Packages.props`/`packages.lock.json` diff understood line by line.
- No package shipping native binaries when a managed equivalent exists; native dependencies complicate chiseled images, cross-OS CI, and any future AOT decision, and need an ADR.
- No returning a bare `{"error":"..."}` string for failures; emit RFC 9457 ProblemDetails with the `errors` field map for validation (see [../foundations/serialization.md](../foundations/serialization.md) ### Error Responses).
- No exception to a default in this doc without an ADR recorded.

## Verification And Proof

A dependency choice is proven, not asserted. Before a dependency lands, demonstrate all of:

- The Approval Questions above are answered in writing, in the PR description or the ADR — not left implicit.
- The `Directory.Packages.props` and `packages.lock.json` diff is reviewed and understood: every added direct and transitive package is accounted for, and the size of the blast radius is acceptable.
- The restore audit is clean against the new dependency set: NuGetAudit reports no high or critical advisories — this runs in the restore (locked) stage of `pwsh ./verify.ps1`.
- An ADR is recorded for any choice that departs from the Default Choices By Concern table, cross-linking [architecture-decision-records.md](architecture-decision-records.md).

### Decision Record

When a repo chooses an exception, the ADR (see [architecture-decision-records.md](architecture-decision-records.md)) must write down:

- the package name and why the default was insufficient
- which solution area is allowed to depend on it
- the operational risk, license, or lock-in tradeoff accepted
- what would trigger re-evaluation or removal later
