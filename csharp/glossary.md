# Glossary

> **Lookup aid, not required reading.** Consult a single entry when a handbook term is unclear; nothing here needs to be read front-to-back to build.

Canonical vocabulary for this handbook. Every term below has exactly one meaning across all repos; use the word for nothing else, and define it nowhere else. Each entry is one or two sentences plus the doc that owns the full rule — read that doc before relying on the term. Start from [README.md](README.md) for how these pieces fit together.

Terms are alphabetical.

## ADR (Architecture Decision Record)
A short, immutable document stating one decision, the context that forced it, and the consequences accepted; once accepted it is frozen and only superseded, never edited. Owned by [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).

## Chiseled image
The distroless-style `aspnet` runtime container base — no shell, no SDK, no package manager, non-root — that production services run on, built by the multi-stage Dockerfile. Owned by [operations/deployment.md](operations/deployment.md); the committed copy is [templates/Dockerfile](templates/Dockerfile).

## Composition root (thin `Program.cs`)
The one place — `Program.cs` — where configuration is composed, implementations are bound to Core ports, and the host lifetime is wired; it is the only spot where `<App>.Api` may touch `<App>.Infrastructure`, and it holds no business logic. Owned by [foundations/project-setup.md](foundations/project-setup.md) and [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md).

## Contract (wire / data contract)
The externally observable shape another system depends on: a wire contract is the request/response or message payload; a data contract is the persisted schema. Both are first-class engineering surfaces, versioned and compatibility-checked like code. Owned by [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md).

## Core / domain (`<App>.Core`)
The project that holds business rules, domain types, and ports, referencing only the BCL and abstraction-only packages. It must not reference ASP.NET Core, EF Core, broker SDKs, or any other project in the solution; everything else points inward at it. Owned by [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md).

## CPM (Central Package Management)
The MSBuild feature that puts every package version in one `Directory.Packages.props` file; `.csproj` files reference packages by name only, so versions cannot drift per project. Owned by [foundations/project-setup.md](foundations/project-setup.md); the committed copy is [templates/Directory.Packages.props](templates/Directory.Packages.props).

## DTO (data transfer object)
A wire record defined next to the endpoint group and mapped explicitly to and from domain types at the boundary; it is never the domain or `<App>.Core` type. The DTO decouples the public wire contract from internal refactors. Owned by [foundations/serialization.md](foundations/serialization.md).

## Endpoint filter
The minimal-API counterpart of middleware scoped to a route group or endpoint — validation, idempotency, and cross-cutting request logic run here instead of inside handler bodies. Owned by [services/http-services.md](services/http-services.md).

## Endpoint group
A route group (`MapGroup`) plus its handlers and request/response DTOs, registered as one unit in `src/<App>.Api/Endpoints/` — the handbook's unit of HTTP transport organization. Owned by [services/http-services.md](services/http-services.md) and [foundations/shared-constructs.md](foundations/shared-constructs.md).

## Envelope
A stable metadata wrapper around a message payload (event ID, type, source, time, correlation/trace context), CloudEvents-style even where CloudEvents is not adopted literally. It carries the fields delivery and dedupe logic key on, separate from the business payload. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Error budget
The agreed allowance of failure: `(1 - target) × valid events` over an SLO window. It is spent deliberately on releases and risk, and its named owner decides what happens when it is exhausted. Owned by [operations/operability.md](operations/operability.md).

## Expand/contract migration
The staged pattern for destructive schema changes across releases: add the new shape and dual-write, backfill, switch reads, then drop the old shape only in a later release once no running version references it. Never drop, rename, or narrow in the same release as code that still reads the old shape. Owned by [services/database.md](services/database.md).

## Fail-fast config
Configuration bound to sealed options classes with `ValidateDataAnnotations` plus `ValidateOnStart`, so a missing or malformed value kills the process at startup — before any listener opens — rather than failing later under load. Owned by [foundations/configuration.md](foundations/configuration.md).

## Idempotency
The property that processing the same message or request more than once yields the same effect as processing it once. Consumers are made idempotent (typically via an inbox/dedupe table keyed by event ID) before retries are added; HTTP writes use an `Idempotency-Key`. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md) and [recipes/add-idempotent-write.md](recipes/add-idempotent-write.md).

## Inbox
A durable dedupe table on the consumer side, keyed by event ID, that records which messages have already been handled so duplicate delivery is harmless. The consumer-side counterpart to the outbox. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Liveness vs readiness
Liveness (`/livez`) answers "is the process up and not deadlocked?" and is always cheap; readiness (`/readyz`) answers "can it serve traffic right now?" and is dependency-aware. They are distinct endpoints: a service can be live but not ready when a critical dependency is down. Owned by [operations/observability.md](operations/observability.md).

## Locked restore
`dotnet restore --locked-mode` against the committed `packages.lock.json`: restore fails when the lockfile disagrees with the project graph instead of silently rewriting it, so the dependency graph CI proves is the one that was reviewed. Owned by [foundations/project-setup.md](foundations/project-setup.md) and [operations/ci-and-release.md](operations/ci-and-release.md).

## MTP (Microsoft.Testing.Platform)
The test runner platform xUnit v3 executes on — test projects build as self-contained executables rather than loading into an external runner host. The handbook's test projects copy the MTP configuration from the templates. Owned by [quality/testing.md](quality/testing.md).

## Options pattern
Typed configuration: each config section binds to one sealed `<Section>Options` class consumed via `IOptions<T>`, with validation at startup; nothing outside the composition root touches `IConfiguration`. Owned by [foundations/configuration.md](foundations/configuration.md).

## Outbox
A transactional outbox table written in the same DB transaction as the business state, so committing state and recording the intent to publish is one atomic action; the row is relayed to the broker asynchronously. Used by default when a service both commits state and emits an event. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Port
An interface declared in `<App>.Core` and named for what the consumer needs (`IOrderRepository`, `IPaymentGateway`), implemented by `<App>.Infrastructure`. Ports are the substitution point for hand-rolled fakes and the boundary keeping EF Core and SDKs out of the domain. Owned by [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md).

## ProblemDetails
The RFC 9457 `application/problem+json` error envelope — ASP.NET Core's native shape via `AddProblemDetails` — extended with `requestId` and an `errors` field→messages map for validation. It is the only wire error contract; a bare error string is forbidden. Owned by [foundations/serialization.md](foundations/serialization.md) and [foundations/errors-and-logging.md](foundations/errors-and-logging.md).

## Relay
The asynchronous worker that reads committed outbox rows and publishes them to the broker, decoupling the publish from the original transaction. Its failure semantics are explicit, not hidden behind a repository abstraction. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Settlement
The point at which a consumer durably finalizes a message — acknowledge or commit offset — performed only after durable side effects, never after mere parsing. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## SLI (Service Level Indicator)
A ratio of good events to valid events, computed entirely from telemetry the service already emits; one SLI measures one user-visible promise. Owned by [operations/operability.md](operations/operability.md).

## SLO (Service Level Objective)
A target for an SLI over a rolling window, always stated as `target over window` (e.g. "99.9% over 30 rolling days"); both halves are load-bearing. Owned by [operations/operability.md](operations/operability.md).

## Telemetry seam (`AddServiceTelemetry()`)
The `src/<App>.Api/Telemetry/` extension that owns logging, OpenTelemetry, and health wiring in one call — the construction point observability and SLIs build on, and nothing more. Owned by [operations/observability.md](operations/observability.md) and [foundations/shared-constructs.md](foundations/shared-constructs.md).

## Typed client
An outbound HTTP client class registered through `IHttpClientFactory` with the standard resilience handler attached, living in `src/<App>.Infrastructure/Clients/` and consumed through a Core port. Owned by [foundations/shared-constructs.md](foundations/shared-constructs.md); the procedure is [recipes/add-external-client.md](recipes/add-external-client.md).

## Verify gate (`pwsh ./verify.ps1`)
The single committed proof gate: restore (locked), format-check, build (warnings-as-errors), test, audit, run identically locally and in CI on an ubuntu/windows/macos matrix. Integration tests sit behind the explicit `-Integration` switch because Docker is not guaranteed on every dev machine. A change is not done until the gate is green. Owned by [quality/testing.md](quality/testing.md) and [operations/ci-and-release.md](operations/ci-and-release.md); the fast-path summary is in [AGENTS.md](AGENTS.md).

## Related

- [README.md](README.md) — how the layers, defaults, and proof gate fit together.
- [AGENTS.md](AGENTS.md) — the fast-path contract that uses this vocabulary.
- [AGENTS.md](AGENTS.md) (## Change Routing) — which doc owns each change surface.
