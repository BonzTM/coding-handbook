# Spec Intake Checklist

Pre-flight checklist run BEFORE an agent writes any code. The handbook supplies the HOW — layout, error model, logging, migrations, the `make verify` gate — but it deliberately defers a set of WHAT decisions to you and an ADR. This checklist makes those deferrals explicit so a complete answer set yields a one-shot build instead of a stall mid-implementation. Each item names the handbook doc or ADR that consumes the answer; an unanswered box is an open question the build cannot absorb later for free.

Run it top to bottom. A box is "answered" only when the answer is concrete enough to wire (a named scheme, a named store, a number), not "TBD" or "probably". Every answer that departs from a handbook default becomes an ADR before code starts.

## Shape & Scope

- [ ] Shape is one of service, worker, CLI, library, or a named combination — this fixes the `cmd/`+`internal/` layout and entrypoint count ([new-project.md](new-project.md)).
- [ ] The MVP's bounded feature set is written down: what ships in v1 and, explicitly, what does not. Scope creep mid-build is the most common one-shot killer.
- [ ] Each boundary is classified sync (request/response) or async (queued/event-driven); async boundaries pull in the Integration section below.
- [ ] Boundaries needing an explicit contract (`api/`, schema source, transport doc) are identified ([new-project.md](new-project.md)).

## Identity & Access

- [ ] Authentication scheme is named: OIDC, bearer JWT, mTLS, API key, or none-by-design (e.g. an internal-only worker). This is an ADR-worthy auth model ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)) and shapes the request boundary ([../operations/security.md](../operations/security.md)).
- [ ] Authorization model is named: RBAC, ABAC, resource ownership, or a combination — including where the decision is enforced (middleware vs handler vs query).
- [ ] Token/credential issuer and validator are identified: who mints credentials, who validates them, and where the keys/JWKS come from. Records to the auth ADR.
- [ ] Sensitive auth events that must be auditable are listed for the audit-logging path ([../operations/security.md](../operations/security.md)).

## Tenancy

- [ ] Single-tenant or multi-tenant is decided — this is irreversible-grade and an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- [ ] If multi-tenant, the isolation model is chosen: Postgres row-level security (RLS) vs application-scoped `tenant_id` filtering vs database/schema-per-tenant — with the tradeoff recorded.
- [ ] Tenant resolution is defined: how the tenant is derived from the authenticated principal (claim, header, subdomain) and threaded through to every query.

## Data

- [ ] Primary store is Postgres unless an ADR says otherwise; any additional datastore (cache, search, object store, queue) is named and justified in an ADR ([../decisions/framework-selection.md](../decisions/framework-selection.md), [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- [ ] Every entity/field is classified for sensitivity and PII per [../operations/data-handling.md](../operations/data-handling.md), so encryption, redaction, and logging rules are known before the schema is written.
- [ ] Retention and deletion expectations are stated per data class (how long, hard vs soft delete, deletion-on-request) ([../operations/data-handling.md](../operations/data-handling.md)).

## Integration

- [ ] If event-driven, the message broker is named and its delivery semantics are pinned: at-least-once vs exactly-once-effective, ordering guarantees, retry limits, and DLQ behavior. ADR-worthy ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md), [../decisions/framework-selection.md](../decisions/framework-selection.md)).
- [ ] At-least-once consumers have an idempotency strategy chosen — the inbox/dedupe model for broker-delivered duplicates ([../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)); HTTP write retries use the keyed-write recipe instead ([../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md)).
- [ ] External dependencies are listed with their SLAs, timeout/retry posture, and failure mode (degrade vs fail-closed) ([../operations/resilience.md](../operations/resilience.md)).

## Runtime & Deploy

- [ ] Target platform is named (Kubernetes, a specific PaaS, bare VM, serverless) — it drives health/readiness and shutdown wiring ([../operations/deployment.md](../operations/deployment.md)).
- [ ] Secrets manager / source of injected secrets is named (e.g. cloud secrets manager, mounted files, env from orchestrator) — never committed.
- [ ] Observability backend is named: Prometheus scrape vs OTel collector for metrics/traces, and the log sink ([../operations/observability.md](../operations/observability.md)).
- [ ] The multi-environment config source is decided (env vars, mounted config, parameter store) and how dev/staging/prod differ ([new-project.md](new-project.md)).

## Compliance & SLOs

- [ ] Regulatory posture is stated (none, GDPR, HIPAA, PCI, SOC2, etc.) — it feeds data classification, retention, and audit-logging requirements ([../operations/data-handling.md](../operations/data-handling.md), [../operations/security.md](../operations/security.md)).
- [ ] SLO targets and an error budget are set (availability, latency objectives) so alerts and rollout gates have a definition of "healthy" ([rollout-and-slo-readiness.md](rollout-and-slo-readiness.md)).

## Verification

Intake is complete, and the build may start, when:

- every box above is answered with a concrete, wire-ready choice — no "TBD";
- each non-default choice (auth model, tenancy model, additional datastore, broker, transport, any handbook deviation) is recorded as an accepted ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md), with the baseline stack captured in `decisions/0001-*`;
- the bounded MVP scope is written down and agreed, so a one-shot build has fixed edges.

If any box is unanswerable, that is the next conversation to have — not a default to silently assume.
