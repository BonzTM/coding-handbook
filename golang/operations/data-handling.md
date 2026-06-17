# Data Handling

How a Go service classifies, minimizes, redacts, protects, and accounts for the data it touches — so privacy and compliance posture is a property of the code, not a slide deck.

## Default Approach

Every field that crosses a boundary or lands in storage has a known classification, a known retention, and a known fate in logs and traces. The contract is decided once, at intake, and enforced in code — not rediscovered during an incident or an audit. Classification drives every other rule on this page: how long you keep a field, whether it may appear in a log line, and how it is encrypted.

This doc governs the data itself. For where secret *material* lives and rotates see [security.md](security.md) `### Secrets`; for the telemetry sinks that must never carry it see [observability.md](observability.md); for the wire boundary where redaction is applied see [../foundations/serialization.md](../foundations/serialization.md). New work declares its data classifications during [spec intake](../checklists/spec-intake.md) before any code is written.

### Classification

Assign every field a tier. Do this at the spec stage for every field that crosses a trust boundary (request, response, event, log) or is persisted.

| Tier | Examples | Handling |
|---|---|---|
| public | published product names, public IDs | no restriction |
| internal | non-sensitive operational data, internal counters | not for external exposure; safe in logs |
| confidential | business-sensitive data, contractual terms | access-controlled; redact from broad sinks |
| restricted-PII | names, emails, government IDs, location, financial, health, credentials | encrypt at rest, redact everywhere, retention-bound, subject-rights-eligible |

- Classify the field, not just the table. A `users` row mixes tiers: `id` (public), `created_at` (internal), `email` (restricted-PII). The unit of classification is the column / wire field.
- Default unclassified data to the **most restrictive plausible tier**, not the least. An unlabeled field is treated as confidential until someone proves it is internal or public — the opposite of the common failure where everything defaults to public.
- Record the tier where the type is defined (a struct-tag, a comment, or the data inventory below) so a reviewer can see it without guessing. A `restricted-PII` field carries that label at its definition site.

### Minimization & Retention

Collect only what a feature needs, and decide when each tier expires *when you decide to collect it*.

- Do not collect a field "in case it is useful later." Every restricted-PII field you hold is liability and audit scope. If a feature does not need it, the field does not exist.
- Set a default retention per tier and make expiry a real, scheduled operation — a deletion job, a partition drop, a TTL — not a wishful comment. Deletion is a feature with code and tests, not an afterthought.

| Tier | Default retention |
|---|---|
| public | indefinite |
| internal | bounded to operational need (e.g. 90 days for ops logs) |
| confidential | bounded to business/contractual need; documented per dataset |
| restricted-PII | shortest that satisfies the lawful basis; delete or anonymize on expiry and on subject request |

- Prefer anonymization or aggregation over raw retention. A counter that survives the underlying PII being deleted keeps the analytic value without the liability.
- Tie retention to the deletion path in [### Lawful Basis & Subject Rights](#lawful-basis--subject-rights): the same code that expires data on schedule expires it on request.

### Redaction

PII and secrets never appear in logs, metric labels, traces, or error messages. Redact at the boundary, before anything reaches a sink.

- A `restricted-PII` or secret value is **never** a log field value, a span attribute, an error string, or a metric label. Metric labels also have a separate hard rule — they stay low-cardinality (no user IDs, emails, request IDs); see [observability.md](observability.md) `### Metrics`.
- Redact where the data enters the sink, not optimistically at every call site. A logging helper that knows a field is sensitive (a tagged struct field, a `LogValue()` that returns a masked form, a wrapper type) makes leakage the exception path rather than the default. Tag sensitive fields as such at their definition so the redaction is structural, not a reviewer's memory.
- Errors carry context, not payloads. `fmt.Errorf("decode user %d: %w", id, err)` is fine; `fmt.Errorf("decode user %s: %w", email, err)` leaks PII into every log and trace that error touches. Keep secrets and PII out of error strings — they propagate to sinks you did not anticipate (see [security.md](security.md) `### Secrets` and the audit trail in [security.md](security.md) `### Audit Logging`).
- The serialization boundary is where redaction is enforced for the wire: a DTO that omits or masks a restricted field is the structural fix, not a downstream filter. See [../foundations/serialization.md](../foundations/serialization.md).

### At-Rest & In-Transit

Encrypt restricted data at rest and use TLS for every hop; treat plaintext as the exception that needs justification.

- TLS everywhere in transit — external callers, internal service-to-service, and the database connection. Plaintext transport for restricted or confidential data is a defect, not a tuning knob.
- Encrypt `restricted-PII` at rest. The default is transparent storage-layer encryption (encrypted volume / managed-DB encryption); add application-layer (column / envelope) encryption when the threat model requires that the storage operator cannot read the field.
- Key management — KMS, envelope keys, rotation — is a library/platform decision, not an ad-hoc import: it departs from the stdlib default, so it earns an ADR per [../decisions/framework-selection.md](../decisions/framework-selection.md). Source the key material itself per [security.md](security.md) `### Secrets` (injected at runtime, rotatable, never in source).

### Lawful Basis & Subject Rights

Hold a minimal, defensible GDPR/CCPA posture, backed by an audit trail an auditor can read.

- **Data inventory.** Maintain a living inventory of every restricted-PII and confidential dataset: what field, which tier, where it is stored, its retention, and its lawful basis for processing. This is the artifact an audit asks for first; the per-field classification above is what populates it.
- **Lawful basis.** Each restricted-PII field has a stated reason it is processed (consent, contract, legitimate interest, legal obligation). A field with no basis is a field you delete.
- **Subject rights — export and delete on request.** Build the export ("what do you hold about me") and delete/anonymize paths as real features with tests, reusing the retention deletion path above. A delete that misses the search index, the cache, the backups policy, or the event log is not a delete.
- **Audit trail.** Access to and changes affecting restricted data emit a structured, tamper-evident audit event — who, what, when, on which subject — routed through [security.md](security.md) `### Audit Logging`. This is what makes the posture SOC2-friendly: the control is demonstrable, not asserted.

## Common Mistakes And Forbidden Patterns

- PII or secrets in log fields, metric labels, span attributes, or error strings — the single most common privacy leak, and the hardest to claw back once shipped to a sink.
- No retention, so every tier accumulates forever; the dataset grows unbounded and so does breach blast radius and audit scope.
- Classifying nothing, so everything is implicitly treated as public and the most-restrictive-default rule is inverted.
- A restricted-PII field collected "just in case," with no feature needing it and no lawful basis stated.
- A delete or export path that covers the primary table but silently misses caches, search indexes, replicas, or event logs.
- Application code reading raw secret/PII into an error or panic message; route material per [security.md](security.md) `### Secrets`.
- High-cardinality identifiers used as metric labels — both a cardinality bug and, for user IDs/emails, a PII leak (see [observability.md](observability.md) `### Metrics`).

## Verification And Proof

- A data inventory exists and is current: every restricted-PII and confidential dataset lists field, tier, storage location, retention, and lawful basis.
- Logs, traces, and metric labels are scanned (a grep over a representative capture, plus a redaction unit test on sensitive types) and show no restricted-PII or secret values.
- Retention is enforced by a scheduled, tested deletion/expiry path — not a comment — and the subject-delete path is proven to reach every store (primary, cache, index, replica, event log).
- Restricted data is encrypted at rest (storage- or application-layer per threat model) and every hop carrying it uses TLS.
- A subject export and a subject delete each run end to end in a test and produce the expected result.
- The audit trail records access to restricted data with who/what/when/subject (see [security.md](security.md) `### Audit Logging`).
