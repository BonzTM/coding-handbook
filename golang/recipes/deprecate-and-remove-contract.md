# Recipe: Deprecate And Remove A Contract

Use this when you need to retire an HTTP route or field, a proto field, an event field, or a DB column without breaking any live consumer.

This is a multi-release procedure, not a single PR. Removal is gated on observed zero usage, never on a calendar date alone.

## Files To Touch

- the contract source: `api/<svc>/v<N>/<svc>.proto`, the OpenAPI spec, the event envelope/schema, or `internal/db/migrations/NNNN_<desc>.sql`
- the transport or handler that serves the field: `internal/api/http/...`, `internal/api/grpc/...`, or the producer/consumer package
- telemetry: `internal/telemetry/telemetry.go` (add a deprecation-usage counter to the `Metrics` seam) plus the call site that increments it
- `CHANGELOG.md` / release notes (announce deprecation, then announce removal with the target version)
- transport, contract, and mixed-version compatibility tests

## Steps

1. Mark deprecated in the contract source, following the governing [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md) (for transport specifics: DB columns [../services/database.md](../services/database.md), proto [../services/grpc-services.md](../services/grpc-services.md), events [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)). This release stays fully backward compatible: nothing is removed yet, only annotated.
   - proto field or RPC: add `[deprecated = true]` (or `option deprecated = true;` for an RPC) and regenerate. Do NOT renumber or delete yet.
   - HTTP route or field: emit a `Sunset: <HTTP-date>` response header (RFC 8594) on the deprecated route, plus a `Deprecation` header (the in-progress IETF `draft-ietf-httpapi-deprecation-header`; pin the exact form in [../decisions/framework-selection.md](../decisions/framework-selection.md) since the draft's value format is not yet stable), and document the notice in the OpenAPI description. For a deprecated field, keep serving it unchanged.
   - event field: add a `// Deprecated:` note to the envelope struct and the schema; keep populating the field so existing consumers do not break.
   - Go-visible types: add a `// Deprecated: use X instead; removed in vN.` godoc comment so `staticcheck` (SA1019) flags internal callers.
   - DB column: begin expand/contract. The deprecation release is the "expand-done, stop-writing" phase — code stops writing the column but the column still exists and old rows keep their values. Do not drop in this release.
2. Add deprecation telemetry so you can PROVE traffic reached zero before removal. Add a method to the `Metrics` interface in `internal/telemetry/telemetry.go` (mirror `IncWidgetCreated`), keep labels low-cardinality — label by the deprecated contract element, never by request/user/tenant ID — and increment it at the exact point the deprecated field/route/column is read or served. Follow [../recipes/add-metric.md](../recipes/add-metric.md) for the seam mechanics.
3. Announce in `CHANGELOG.md` / release notes for the deprecation release: what is deprecated, the replacement, the deprecation-window expectation, and the earliest version it MAY be removed in. State that removal is contingent on the usage metric hitting zero.
4. Remove only in a later release, after the metric has shown zero usage for the full documented window across all live client versions.
   - proto: delete the field and add `reserved <number>;` and `reserved "<name>";` so the tag and name can never be reused. Delete a removed RPC's method.
   - HTTP: delete the route/field; return `410 Gone` for a removed route only if clients still probe it, otherwise `404`.
   - event: stop populating and remove the field from the envelope and schema.
   - DB: add a new contract migration `internal/db/migrations/NNNN_drop_<col>.sql` whose `-- +goose Up` runs `ALTER TABLE ... DROP COLUMN ...;` and whose `-- +goose Down` documents that the drop is forward-only (recovery is restore-from-backup, not a down-migration). Run it only after no deployed code references the column. See [../recipes/add-migration.md](../recipes/add-migration.md).
   - Announce the removal in `CHANGELOG.md` with the canonical release tag (e.g. `v2.0.0`), and remove the deprecation counter in the same or a following PR.

## Invariants To Preserve

- mixed-version rollout never breaks: during deprecation, old and new binaries (and old and new clients) must both work against the same contract
- removal is gated on observed zero usage of the deprecation counter, not on time alone
- proto field numbers and names are reserved on removal and never reused
- DB removal follows expand/contract: stop-writing, observe zero reads, then drop in a separate later migration — never drop a column the running code still reads
- the deprecation window is documented in release notes before removal happens
- additive replacement ships before or with the deprecation, so consumers always have a migration path

## Proof

- `make verify` is green at every release in the sequence
- the deprecation counter reads zero over the documented window before the removal PR merges (cite the metrics query / dashboard in the removal PR)
- mixed-version compatibility test: a test that decodes a payload produced by the old contract with the new code and vice versa, e.g. `go test ./internal/api/grpc/... -run TestDeprecatedFieldRoundTrip`
- proto removal: `buf breaking` (or the repo's proto-break check, run under `make verify`) passes because the number/name is reserved, not reused
- HTTP: write a transport test (e.g. `go test ./internal/api/http/... -run TestDeprecatedRouteHeaders`) that asserts the deprecated route serves the `Deprecation` and `Sunset` headers during the window, and — after the removal release — that the route is gone (`404`, or `410` if clients still probe it). The reference service does not currently demonstrate deprecation headers; model the test on its existing `httptest`-based handler tests.
- DB: apply-then-reapply the drop against a real database (`goose -dir internal/db/migrations <driver> "$TEST_DSN" up`), prove no live query references the dropped column, and confirm via the expand/contract proof in [../services/database.md](../services/database.md) that release N and N+1 code both run against the release-N schema
- `CHANGELOG.md` contains the deprecation entry and, in the removal release, the removal entry with the version tag
