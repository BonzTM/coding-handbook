# Data Handling

How Python systems classify, minimize, redact, retain, protect, export, and delete owned data.

## Default Approach

Every field crossing a trust boundary or entering storage has an owner, classification, purpose, lawful/contractual basis where applicable, retention, deletion/export behavior, and telemetry rule. Decide these at [spec intake](../checklists/spec-intake.md), before schema or DTO code.

### Classification

| Tier | Examples | Required handling |
|---|---|---|
| public | intentionally published content and identifiers | integrity and availability controls |
| internal | non-sensitive operational state | authenticated internal access; safe only in approved sinks |
| confidential | contracts, business plans, private system data | least privilege, encryption, broad-sink redaction |
| restricted | PII, financial/health data, credentials, authentication material | strongest access, encryption, no telemetry, retention/deletion rights |

Classify the field, not merely the table or model. Record classification beside the schema/contract or in a versioned data inventory containing field, purpose, owner, storage, flows, retention, and deletion/export obligations. Unclassified data defaults to the most restrictive plausible tier until reviewed.

### Minimization And Purpose

Collect and persist only fields the owned feature needs. Do not retain payloads “for debugging” or future analysis. Prefer stable opaque identifiers, aggregates, or anonymized results over raw personal data. A new purpose, recipient, or derived dataset requires classification and retention review; existing consent or authorization is not assumed to cover it.

Test fixtures, telemetry, caches, queues, search indexes, backups, and analytics are part of the data flow. They do not become exempt because they are secondary stores.

### Redaction At The Logging Boundary

PII, restricted values, credentials, tokens, request/response bodies, and sensitive query parameters never enter logs, metric labels, span attributes, exception strings, or audit payloads. Operational identifiers are allowed only when classified and necessary; request/trace IDs are not user identity.

Use one allowlist-based record-filter/formatter helper at the logging sink boundary. It serializes only the stable schema, replaces known sensitive fields with a constant marker, and safely renders unknown objects without calling a secret-bearing `repr`. Pydantic `SecretStr` reduces accidental display but is not the redaction control. Unit-test the helper with nested mappings, exceptions, Pydantic models, dataclasses, and malicious objects.

Errors carry stable codes and safe context, not payload data. Traces follow the same allowlist; metric labels contain only bounded dimensions per [observability](observability.md).

### Retention And Deletion

Every persisted dataset and telemetry sink has a documented duration and disposition. Enforce expiry with a bounded, observable, idempotent scheduled operation or storage lifecycle rule. Soft deletion does not satisfy erasure by itself. Define how primary databases, caches, indexes, object stores, broker/DLQ data, derived tables, replicas, archives, and backup expiry converge.

Deletion accepts a typed subject/resource identity, authorizes the request, records a safe audit event, deletes or irreversibly anonymizes within the promised window, and produces a reconcilable result. Failed partial deletion is retryable and visible. Backups follow documented expiry and restore-time re-deletion policy rather than unsafe in-place editing.

### Export And Portability

Exports authenticate and authorize the requester, include only the subject's permitted data, use a versioned documented format, and run within bounded size/time. Large exports execute as owned jobs with expiring opaque download links, encryption, rate limits, cancellation, and audit events. Never email raw restricted data or expose a predictable object path.

Contract tests prove completeness and prevent accidental inclusion of internal fields. Export and delete inventories remain synchronized.

### Encryption And Key Ownership

Use TLS for every hop carrying confidential/restricted data, including service-to-service and database connections, with termination boundaries documented. Use platform/storage encryption at rest by default. Add application/field-level envelope encryption only when the threat model requires separation from storage operators; the KMS/library choice, key hierarchy, rotation, recovery, and search limitations require an ADR.

Keys and credentials follow [security](security.md): runtime injection, least privilege, rotation, no source/image/telemetry exposure. Encryption does not relax authorization, minimization, or deletion.

### Test Data

Production data never appears in fixtures, snapshots, local databases, demos, screenshots, or lower environments. Build synthetic deterministic factories that preserve schema edge cases without copying identity. If realistic statistical shape is required, generate or irreversibly anonymize through an approved documented pipeline; tokenization alone is not anonymization when a lookup remains.

Test artifacts and failure output follow the same classification/retention rules. CI logs must not dump database rows, HTTP bodies, settings, or environment variables.

## Common Mistakes And Forbidden Patterns

- Classification at table/model level only, leaving mixed-sensitivity fields ambiguous.
- Unclassified data treated as public or collected for unspecified future use.
- PII, tokens, payloads, email addresses, user IDs, or raw paths in logs, metrics, traces, errors, or audit records.
- Call-site-only redaction, denylist-only filters, or relying on `SecretStr.__repr__` as the sink control.
- Retention documented but not executed; soft deletion presented as complete erasure.
- Delete/export path covering only the primary table while caches, indexes, queues, derived data, or backups are ignored.
- Export link without expiry/authorization or export schema exposing internal fields.
- Encryption claimed without identifying every transport hop, storage layer, and key owner.
- Production rows copied into fixtures or lower environments, even after superficial field replacement.

## Verification And Proof

```bash
uv run pytest -k "redact or retention or deletion or export or classification"
make verify
```

Review the versioned data inventory against DTOs, tables, events, caches, telemetry, and backups. Seed canary PII/secrets and scan representative logs, traces, metrics, errors, audit records, CI output, and image artifacts for absence. Run retention, subject deletion, and export end to end across every listed store, including partial failure/retry and backup policy. Prove TLS/storage encryption configuration and confirm fixtures contain only synthetic approved data.

Related: [serialization](../foundations/serialization.md), [security](security.md), [observability](observability.md), and [incident response](../checklists/incident-response.md).
