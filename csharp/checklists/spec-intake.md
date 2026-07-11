# Spec Intake Checklist

Pre-flight checklist run BEFORE an agent writes any code. The handbook supplies the HOW — solution layout, error model, logging, migrations, the `pwsh ./verify.ps1` gate — and this checklist covers the WHAT decisions the spec must supply. Each item names the handbook doc or ADR that consumes the answer. A box is "answered" only when the answer is concrete enough to wire (a named scheme, a named store, a number), not "TBD" or "probably".

## How To Resolve Each Box

Resolve every applicable box in this order. Intake never stalls the build; it decides *where each answer comes from* and makes that traceable.

1. **From the spec.** If the request already answers it, record the answer and move on.
2. **Ask.** Asking beats inferring. When the requester is reachable, ask the open questions that materially change the build — batched, once, up front. Do not interrogate box-by-box, and do not ask about boxes the defaults table below already covers unless the answer would change the architecture (see step 3 for the two that always deserve a question).
3. **Default.** When the requester is unreachable, has said "just build it", or the box is low-stakes for an MVP: take the entry from [Defaults When The Spec Is Silent](#defaults-when-the-spec-is-silent), record it as a stated assumption, and proceed. Two boxes are irreversible-grade — **tenancy** and **compliance posture** — ask about them whenever interaction is possible; when it is not, take the default and flag the assumption at the top of the delivery notes, not buried in an ADR appendix.
4. **Skip.** Sections marked as not applying to the project's shape are skipped, not answered. A CLI tool has no tenancy model; a library has no deploy target. Do not ask questions the shape makes meaningless.

Never invent an answer that is neither in the spec, nor from the requester, nor in the defaults table. Every defaulted answer is disclosed in the delivery summary (project README or the baseline ADR) so the requester can veto it cheaply; ADR-grade defaults (marked below) get a real ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

## Section Applicability By Shape

| Section | HTTP/web service | gRPC service | Event-driven worker | CLI | Library |
|---|---|---|---|---|---|
| Shape & Scope | yes | yes | yes | yes | yes |
| Identity & Access | yes | yes | broker-credential posture only | skip (none-by-design) | skip |
| Tenancy | if it persists data for external principals | same | same | skip | skip |
| Data | if it persists data | same | same | if it writes files/state | skip |
| Integration | if it has async boundaries or external dependencies | same | yes | external calls only | skip |
| Runtime & Deploy | yes | yes | yes | release/registry rows only | skip |
| Compliance & SLOs | yes | yes | yes | compliance row only | skip |

## Shape & Scope

- [ ] Shape is one of HTTP/API service, server-rendered web app, gRPC service, worker, CLI, library, or a named combination — this fixes the `src/` project split and entrypoint count ([new-project.md](new-project.md), [../foundations/solution-and-project-design.md](../foundations/solution-and-project-design.md); web apps additionally follow [../services/web-apps.md](../services/web-apps.md)).
- [ ] The MVP's bounded feature set is written down: what ships in v1 and, explicitly, what does not. Scope creep mid-build is the most common one-shot killer.
- [ ] Each boundary is classified sync (request/response) or async (queued/event-driven); async boundaries pull in the Integration section below.
- [ ] Whether a browser client calls the API is decided; if so, the allowed CORS origins and credentials policy are listed — this shapes the middleware pipeline ([../services/http-services.md](../services/http-services.md), [../operations/security.md](../operations/security.md)).
- [ ] Boundaries needing an explicit contract (`api/` protos, OpenAPI document, schema source, transport doc) are identified ([new-project.md](new-project.md)).

## Identity & Access

- [ ] Authentication scheme is named: OIDC, bearer JWT, mTLS, API key, or none-by-design (e.g. an internal-only worker). This is an ADR-worthy auth model ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)) and shapes the request boundary ([../operations/security.md](../operations/security.md)).
- [ ] Authorization model is named: RBAC, ABAC, resource ownership, or a combination — including where the decision is enforced (authorization policy on the route group vs endpoint filter vs query).
- [ ] Token/credential issuer and validator are identified: who mints credentials, who validates them, and where the keys/JWKS come from. Records to the auth ADR.
- [ ] Sensitive auth events that must be auditable are listed for the audit-logging path ([../operations/security.md](../operations/security.md)).

## Tenancy

- [ ] Single-tenant or multi-tenant is decided — this is irreversible-grade and an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)); ask when you can, and flag the default loudly when you cannot.
- [ ] If multi-tenant, the isolation model is chosen: Postgres row-level security (RLS) vs application-scoped `tenant_id` filtering (an EF Core global query filter) vs database/schema-per-tenant — with the tradeoff recorded.
- [ ] Tenant resolution is defined: how the tenant is derived from the authenticated principal (claim, header, subdomain) and threaded through to every query.

## Data

- [ ] Primary store is PostgreSQL via EF Core/Npgsql unless an ADR says otherwise ([../services/database.md](../services/database.md)); any additional datastore (cache, search, object store, queue) is named and justified in an ADR ([../decisions/framework-selection.md](../decisions/framework-selection.md), [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- [ ] Every entity/field is classified for sensitivity and PII per [../operations/data-handling.md](../operations/data-handling.md), so encryption, redaction, and logging rules are known before the schema is written.
- [ ] Retention and deletion expectations are stated per data class (how long, hard vs soft delete, deletion-on-request) ([../operations/data-handling.md](../operations/data-handling.md)).

## Integration

- [ ] If event-driven, the message broker is named and its delivery semantics are pinned: at-least-once vs exactly-once-effective, ordering guarantees, retry limits, and DLQ behavior. ADR-worthy ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md), [../decisions/framework-selection.md](../decisions/framework-selection.md)).
- [ ] At-least-once consumers have an idempotency strategy chosen — the inbox/dedupe model for broker-delivered duplicates ([../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)); HTTP write retries use the keyed-write recipe instead ([../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md)).
- [ ] External dependencies are listed with their SLAs, timeout/retry posture, and failure mode (degrade vs fail-closed) ([../operations/resilience.md](../operations/resilience.md)).

## Runtime & Deploy

- [ ] Target platform is named (Kubernetes, a specific PaaS, bare VM, serverless) — it drives health/readiness and shutdown wiring ([../operations/deployment.md](../operations/deployment.md)).
- [ ] Secrets manager / source of injected secrets is named (e.g. cloud secrets manager, mounted files, env from orchestrator) — never committed ([../operations/security.md](../operations/security.md)).
- [ ] Container registry / image-publish target is named (e.g. `ghcr.io`, ECR, Artifact Registry, an internal registry) — the release pipeline pushes there on a `v*` tag ([../operations/ci-and-release.md](../operations/ci-and-release.md)).
- [ ] Observability backend is named: OTLP collector vs Prometheus scrape for metrics/traces, and the log sink ([../operations/observability.md](../operations/observability.md)).
- [ ] The multi-environment config source is decided (env vars, mounted config, parameter store) and how dev/staging/prod differ ([../foundations/configuration.md](../foundations/configuration.md), [new-project.md](new-project.md)).

## Compliance & SLOs

- [ ] Regulatory posture is stated (none, GDPR, HIPAA, PCI, SOC2, etc.) — it feeds data classification, retention, and audit-logging requirements ([../operations/data-handling.md](../operations/data-handling.md), [../operations/security.md](../operations/security.md)). Irreversible-grade: ask when you can.
- [ ] SLO targets and an error budget are set (availability, latency objectives) so alerts and rollout gates have a definition of "healthy" ([rollout-and-slo-readiness.md](rollout-and-slo-readiness.md)).

## Defaults When The Spec Is Silent

The fallbacks for step 3 above. "Record as" says where the assumption must be disclosed: **ADR** means a real ADR in the project's `decisions/`; **note** means a line in the delivery summary / project README assumptions list. Anything not in this table has no silent default — it comes from the spec or the requester.

| Decision | Default when the spec is silent | Record as |
|---|---|---|
| Browser client / CORS | no browser clients → no CORS middleware; a server-rendered web app is same-origin and also needs no CORS | note |
| Authentication | bearer JWT validated against the issuer's JWKS (OIDC-shaped), via the JwtBearer handler wired per [../operations/security.md](../operations/security.md); internal-only workers and CLIs are none-by-design | ADR |
| Authorization | RBAC — role claims mapped to authorization policies applied per route group, with core-level checks for resource ownership | same ADR as auth |
| Tenancy | single-tenant | ADR, flagged at the top of the delivery notes |
| Primary store | PostgreSQL via EF Core/Npgsql per [../services/database.md](../services/database.md) (Dapper only for measured hot paths, via ADR); no second datastore | note |
| PII / classification | classify fields per [../operations/data-handling.md](../operations/data-handling.md); anything person-identifying is treated as confidential; PII never in logs or metrics | note |
| Retention | retain until user-initiated deletion; deletion is a hard delete | note |
| Broker (async required, none named) | build against the repo-owned broker-agnostic seam with SQL outbox + inbox ([../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)); prefer whatever broker the platform already operates; with zero platform signal, NATS JetStream or RabbitMQ behind the seam | ADR |
| Delivery semantics | at-least-once with inbox dedupe on consumers; HTTP unsafe writes take `Idempotency-Key` per [../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md) | note |
| External calls | timeout on every call, bounded retries with jitter per [../operations/resilience.md](../operations/resilience.md); fail-closed unless the spec says degrade | note |
| Target platform | containerized and Kubernetes-shaped — chiseled non-root image, `/livez` + `/readyz` probes, graceful shutdown, [../templates/k8s-deployment.yaml](../templates/k8s-deployment.yaml); runs locally under [../templates/docker-compose.yml](../templates/docker-compose.yml) | note |
| Secrets | env vars injected by the orchestrator/platform; no secrets-manager SDK in application code; `dotnet user-secrets` for local dev only | note |
| Container registry | `ghcr.io`, per [../templates/github-workflows-release.yml](../templates/github-workflows-release.yml) | note |
| Observability backend | OpenTelemetry with OTLP export for metrics and traces (Prometheus scrape via the OTel Prometheus exporter only when the org scrapes), structured `ILogger` JSON to stdout | note |
| Config source | committed `appsettings.json` (no secrets) with env-var overrides, options validated fail-fast at startup (`ValidateOnStart`), documented in the README config table | note (handbook default) |
| Compliance posture | none assumed; the PII and audit-logging rules above still apply | note, flagged at the top of the delivery notes |
| SLO starting targets | 99.9% availability, p99 latency 300ms — written into the runbook as *starting* targets to tune with real traffic, not promises | note + runbook per [../operations/operability.md](../operations/operability.md) |

## Verification

Intake is complete, and the build may start, when:

- every applicable box is resolved — from the spec, from the requester, or from the defaults table — and skipped sections are skipped because of shape, not convenience;
- each non-default choice and each ADR-grade default (auth model, tenancy, additional datastore, broker, platform deviation) is recorded as an accepted ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md), with the baseline stack captured in `decisions/0001-*`;
- every silently defaulted answer is listed in the delivery summary, with tenancy and compliance assumptions flagged first;
- the bounded MVP scope is written down, so a one-shot build has fixed edges.

An unanswerable box is never a silent stall: ask when you can, default and disclose when you cannot.
