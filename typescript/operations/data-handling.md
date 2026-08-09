# Data Handling

Classification, minimization, retention, deletion, export, backup, and telemetry rules for managed data.

## Default Approach

Collect the minimum data required, classify it before storage, and make its full lifecycle enforceable.

### Classification And Inventory

Classify fields and datasets using the organization's scheme, at minimum public, internal, confidential, and restricted/sensitive. Identify personal data, credentials, authentication material, financial/health data, tenant data, and regulated records explicitly.

Maintain an inventory of source, purpose, owner, lawful/business basis where required, storage locations, processors, access roles, retention, backup behavior, exports, and deletion path. A new field is incomplete until this lifecycle is known.

### Minimization And Purpose

Do not collect a value because it may be useful later. Store only the precision and duration required for the stated purpose. Prefer derived or coarse values when raw sensitive input is unnecessary.

Do not repurpose data for telemetry, analytics, model training, or product behavior without the required review and user/contract basis. Synthetic fixtures replace production data in development and tests.

### Access And Tenant Isolation

Apply least privilege to application roles, operators, support tools, analytics, backups, and exports. Enforce tenant scope server-side at every read and write boundary; client-supplied tenant IDs are not authority.

Sensitive access is authenticated, authorized, auditable, and reviewed. Bulk endpoints and support impersonation receive stronger controls and explicit audit events.

### Storage And Transport

Use platform encryption in transit and at rest with owned key management. Application-level encryption requires a threat model, key rotation, query/index impact, backup/restore behavior, and failure recovery.

Never put secrets or sensitive values in cache keys, URLs, queue metadata, metric attributes, traces, logs, exception messages, source maps, or frontend storage. Redact at construction and minimize before emission.

### Retention And Deletion

Define retention by data category and purpose, not one service-wide number. Enforce expiry through bounded, resumable jobs with metrics, failure alerts, and idempotent retry.

Deletion covers primary rows, derived tables, search indexes, caches, objects, queues where feasible, analytics copies, and processors. Document legal holds and backup expiry separately; never promise immediate physical deletion from immutable backups when the system cannot deliver it.

Use deletion tombstones only when they do not preserve the sensitive content and when they are needed to prevent resurrection during replay or restore.

### Export And Portability

Exports authenticate the requester, authorize scope, snapshot consistently, constrain size, avoid cross-tenant joins, encrypt delivery where required, expire artifacts, and audit creation/download.

Use documented, machine-readable formats and safe filenames. Formula injection, path traversal, unsafe archives, and resource exhaustion are part of export threat testing.

### Backups And Restore

Backups inherit the source classification, access policy, encryption, residency, retention, and audit requirements. Define recovery point and recovery time objectives and test restore into an isolated environment.

Restore procedures reconcile deletions, tombstones, credentials, external side effects, and replay so old sensitive records or revoked access are not silently resurrected.

### Non-Production Data

Production data does not enter local development, test fixtures, snapshots, or lower environments by default. Approved copies are minimized or irreversibly transformed, access-controlled, time-limited, and audited.

## Common Mistakes And Forbidden Patterns

- Data fields with no owner, purpose, classification, retention, or deletion path.
- Production payloads copied into tests, tickets, logs, or lower environments.
- Tenant scope trusted from request input or applied only in the UI.
- Retention documented but not enforced or observed.
- Deletion limited to one primary table while caches, indexes, objects, and processors retain copies.
- Claiming immediate backup deletion the platform cannot provide.
- Export files left indefinitely accessible or vulnerable to spreadsheet injection.
- Sensitive identifiers used as metric labels or cache keys.

## Verification And Proof

- Data inventory maps every changed sensitive field to purpose, owner, stores, access, retention, export, and deletion.
- Cross-tenant and privilege negative tests cover reads, writes, bulk operations, and exports.
- Retention jobs are bounded, resumable, idempotent, observable, and tested at expiry boundaries.
- Deletion tests prove primary, derived, cached, indexed, and processor paths plus documented backup expiry.
- Export tests cover authorization, consistency, size, expiry, safe formatting, and audit.
- Restore exercise meets RPO/RTO and does not resurrect deleted data or revoked credentials unexpectedly.
- Telemetry and artifact scans find no secret or sensitive payload leakage.

Related: [security.md](security.md), [observability.md](observability.md), and [../services/caching.md](../services/caching.md).
