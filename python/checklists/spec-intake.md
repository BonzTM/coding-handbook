# Spec Intake Checklist

Pre-flight checklist run BEFORE code. The handbook supplies the HOW—layout, boundaries, errors, migrations, and `make verify`; this checklist resolves the WHAT. A box is answered only when the result is concrete enough to wire, not “TBD.”

## How To Resolve Each Box

Resolve every applicable box in this order:

1. **From the spec.** Record an answer already present.
2. **Ask.** Batch questions that materially change the build. Tenancy and compliance always deserve a question when interaction is possible.
3. **Default.** When the requester is unavailable or delegates the choice, take the documented fallback below, record it as an assumption, and proceed. Flag tenancy and compliance at the top of delivery notes.
4. **Skip.** Skip a section only because the chosen project shape makes it irrelevant.

Never invent an answer outside the spec, requester response, or defaults table. Record ADR-grade defaults in project `decisions/` under [the ADR contract](../decisions/architecture-decision-records.md); disclose every other default in the README/delivery summary.

## Section Applicability By Shape

| Section | HTTP/web service | gRPC service | Worker | CLI | Library |
|---|---|---|---|---|---|
| Shape & Scope | yes | yes | yes | yes | yes |
| Identity & Access | yes | yes | broker credential only | skip unless remote auth | skip |
| Tenancy | when external-principal data persists | same | same | skip | skip |
| Data | when data persists | same | same | files/state only | skip unless API handles data |
| Integration | external/async boundaries | same | yes | external calls | dependency contracts only |
| Runtime & Deploy | yes | yes | yes | distribution rows | distribution rows |
| Compliance & SLOs | yes | yes | yes | compliance if applicable | support/security policy |

## Shape & Scope

- [ ] Shape is HTTP API, server-rendered web app, gRPC service, worker, CLI, library, or a named combination; entry points and package owners follow [project setup](../foundations/project-setup.md).
- [ ] The bounded v1 feature set and explicit non-goals are written down.
- [ ] Each boundary is classified synchronous request/response or asynchronous queued/event-driven.
- [ ] Browser use is decided; when present, exact CORS origins/credentials or same-origin web posture is listed.
- [ ] Published contracts—OpenAPI/JSON, protobuf, event schema, CLI, or library API—are named with owners and compatibility expectations.
- [ ] Acceptance criteria are observable and testable, including negative/error behavior.

## Identity & Access

- [ ] Authentication scheme is named: OIDC/bearer JWT, mTLS, API key, platform identity, or none-by-design; issuer and validator/key source are identified.
- [ ] Authorization model is named: RBAC, ABAC, resource ownership, or combination; enforcement boundaries and deny-by-default behavior are stated.
- [ ] Principal, tenant, and service identity propagation across HTTP/messages/jobs is defined.
- [ ] Authentication failures, authorization denials, privileged writes, and regulated reads requiring [audit logging](../operations/security.md#audit-logging) are listed.

## Tenancy

- [ ] Single-tenant or multi-tenant is decided and recorded in an ADR; this is irreversible-grade.
- [ ] For multi-tenancy, choose application-scoped `tenant_id`, PostgreSQL RLS, or database/schema isolation with its tradeoff.
- [ ] Tenant resolution from the authenticated principal and propagation into every query/cache/event/idempotency key is defined.
- [ ] Cross-tenant negative-test evidence is part of acceptance criteria.

## Data

- [ ] PostgreSQL is the primary store unless an ADR names a different constraint; every additional cache/search/object/queue store is justified.
- [ ] Every boundary/persisted field has a [data classification](../operations/data-handling.md), purpose, owner, and telemetry rule.
- [ ] Retention, scheduled expiry, subject deletion, export, backup behavior, and legal/contractual basis are stated per dataset.
- [ ] Schema evolution and rollback/forward-recovery expectations are defined for the rollout model.
- [ ] Test data is synthetic; production data is forbidden in fixtures and lower environments.

## Integration

- [ ] Each dependency names its protocol, owner, environment endpoint source, availability/latency expectation, timeout, retry classification, idempotency, and fail-closed/degraded behavior.
- [ ] Outbound destinations and redirect/DNS rules are constrained for SSRF under [security](../operations/security.md).
- [ ] Event-driven work names the broker and delivery, ordering, prefetch, retry/age, settlement, DLQ, replay, and schema-compatibility contracts.
- [ ] At-least-once consumers define durable inbox/idempotency behavior; state-plus-publish dual writes define an outbox.
- [ ] Integration proof names real PostgreSQL/broker/stub-server dependencies and the Docker-enabled CI lane.

## Runtime & Deploy

- [ ] Target platform and replica/worker model are named; default is one Uvicorn worker per container with horizontal replicas.
- [ ] Secret provenance and rotation mechanism are named; default is orchestrator env/mounted files plus rolling restart.
- [ ] Container registry/package index and protected publish identity are named.
- [ ] Config sources per environment and which values may safely default are listed.
- [ ] Observability destinations are named: JSON log sink, Prometheus scrape, OpenTelemetry collector/exporter, and audit sink.
- [ ] Resource requests/limits, concurrency/queue/pool bounds, probes, shutdown grace, and migration job ownership are specified or assigned to measured rollout tuning.

## Compliance & SLOs

- [ ] Compliance posture is stated—none, GDPR, HIPAA, PCI DSS, SOC 2, or named regime—and mapped to classification, retention, encryption, audit, and disclosure controls; this is irreversible-grade.
- [ ] User journeys and initial availability/latency/freshness SLIs are named as good/valid ratios.
- [ ] Each SLO states target, rolling window, owner, error-budget action, and rollout abort posture.
- [ ] On-call rotation, escalation, runbook owner, incident communication, and vulnerability-report path are assigned.

## Defaults When The Spec Is Silent

| Decision | Default | Record as |
|---|---|---|
| browser/CORS | no browser caller; no CORS middleware | note |
| authentication | platform/OIDC bearer identity for services; none-by-design for local CLI/library | ADR |
| authorization | RBAC at adapter plus core resource-ownership checks | auth ADR |
| tenancy | single-tenant | ADR; prominently flagged |
| primary store | PostgreSQL through SQLAlchemy async + asyncpg; Alembic migrations | note |
| data | unclassified fields take most restrictive plausible tier; PII never telemetry | note |
| retention | shortest documented product/legal need; hard delete/anonymize on expiry/request | note |
| broker | broker-neutral Protocol seam; use the platform's operated broker, otherwise ask | ADR; no universal broker default |
| delivery | at-least-once with durable inbox; outbox for DB-state publication | note |
| external calls | timeout everywhere, bounded idempotent-only full-jitter retry, fail closed | note |
| target | container/Kubernetes-shaped, one process, probes, bounded drain | note |
| secrets | env/mounted files injected by platform; rolling restart rotation | note |
| registry | organization registry; template placeholder must be resolved | note |
| observability | JSON stdout, Prometheus `/metrics`, OpenTelemetry via standard env config | note |
| config | pydantic-settings from env/files with fail-fast startup | note |
| compliance | no named regime; handbook security/privacy controls still apply | note; prominently flagged |
| SLO | no numeric promise is silently invented; collect baseline and obtain owner approval | open decision |

## Verification

```bash
# No implementation command runs until intake evidence is recorded.
test -f decisions/0001-<baseline>.md
rg -n "Assumptions|Non-goals|Acceptance Criteria" README.md docs decisions
```

- [ ] Every applicable box is resolved from spec, requester, or disclosed default; shape-inapplicable sections alone are skipped.
- [ ] Every ADR-grade choice/default has an accepted decision record and every other assumption is visible in delivery notes.
- [ ] Tenancy and compliance assumptions appear first, and no numeric SLO or broker vendor was silently invented.
- [ ] MVP boundaries and acceptance criteria are fixed enough for implementation and proof.
