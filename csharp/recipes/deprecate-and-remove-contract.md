# Recipe: Deprecate And Remove A Contract

Use this when you need to retire an HTTP route or field, a proto field, an event field, a public C# API, or a DB column without breaking any live consumer.

This is a multi-release procedure, not a single PR. Removal is gated on observed zero usage, never on a calendar date alone.

## Files To Touch

- the contract source: `api/orders/v<N>/orders.proto`, the OpenAPI description, the event schema in `api/events/`, the `[Obsolete]`-annotated public member, or a new EF Core migration
- the transport or handler that serves it: `src/Orders.Api/Endpoints/...`, `src/Orders.Api/Grpc/...`, or the producer/consumer under `src/Orders.Infrastructure/Messaging/`
- telemetry: the deprecation-usage counter on the service's `Meter` (in `src/Orders.Api/Telemetry/`) plus the call site that increments it
- `CHANGELOG.md` / release notes (announce deprecation, then announce removal with the target version)
- transport, contract, and mixed-version compatibility tests

## Steps

1. Mark deprecated in the contract source, following the governing [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md) (for transport specifics: DB columns [../services/database.md](../services/database.md), proto [../services/grpc-services.md](../services/grpc-services.md), events [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)). This release stays fully backward compatible: nothing is removed yet, only annotated.
   - proto field or RPC: add `[deprecated = true]` (or `option deprecated = true;` for an RPC) and rebuild. Do NOT renumber or delete yet.
   - HTTP route or field: emit a `Sunset: <HTTP-date>` response header (RFC 8594) plus a `Deprecation` header (RFC 9745) on the deprecated route — an endpoint filter on the route group is the natural seam — and document the notice in the OpenAPI description. For a deprecated field, keep serving it unchanged.
   - event field: mark it deprecated in the `api/events/` schema and the envelope type's doc comment; keep populating the field so existing consumers do not break.
   - public C# API (library surface): add `[Obsolete("Use CreateOrderV2 instead. Removed in v3.0.0.", DiagnosticId = "ORD0001", UrlFormat = "https://example.com/deprecations/{0}")]`. The message names the replacement and the removal version; the `DiagnosticId` lets consumers suppress or escalate this one deprecation precisely. Internal callers break immediately under `TreatWarningsAsErrors` — migrate them in the same PR.
   - DB column: begin expand/contract. The deprecation release is the "expand-done, stop-writing" phase — code stops writing the column but the column still exists and old rows keep their values. Do not drop in this release.
2. Add deprecation telemetry so you can PROVE traffic reached zero before removal. Add a counter to the service's `Meter`, keep tags low-cardinality — tag by the deprecated contract element, never by request/user/tenant ID — and increment it at the exact point the deprecated field/route/column is read or served. Follow [add-metric.md](add-metric.md) for the seam mechanics.
3. Announce in `CHANGELOG.md` / release notes for the deprecation release: what is deprecated, the replacement, the deprecation-window expectation, and the earliest version it MAY be removed in. State that removal is contingent on the usage metric hitting zero.
4. Remove only in a later release, after the metric has shown zero usage for the full documented window across all live client versions.
   - proto: delete the field and add `reserved <number>;` and `reserved "<name>";` so the tag and name can never be reused. Delete a removed RPC's method.
   - HTTP: delete the route/field; return `410 Gone` for a removed route only if clients still probe it, otherwise `404`.
   - event: stop populating and remove the field from the envelope and schema.
   - public C# API: optionally escalate first with `[Obsolete(..., error: true)]` for one release (compile error, still binary-present), then delete the member in the MAJOR release. For a library, the removal also lands in the PublicAPI files — see [release-library-version.md](release-library-version.md).
   - DB: add a new contract migration whose `Up` runs `DROP COLUMN` and whose `Down` documents that the drop is forward-only (recovery is restore-from-backup, not a down-migration). Run it only after no deployed code references the column. See [add-migration.md](add-migration.md).
   - Announce the removal in `CHANGELOG.md` with the canonical release tag (e.g. `v2.0.0`), and remove the deprecation counter in the same or a following PR.

## Invariants To Preserve

- mixed-version rollout never breaks: during deprecation, old and new binaries (and old and new clients) must both work against the same contract
- removal is gated on observed zero usage of the deprecation counter, not on time alone
- proto field numbers and names are reserved on removal and never reused
- DB removal follows expand/contract: stop-writing, observe zero reads, then drop in a separate later migration — never drop a column the running code still reads
- the deprecation window is documented in release notes before removal happens
- additive replacement ships before or with the deprecation, so consumers always have a migration path

## Proof

- `pwsh ./verify.ps1` is green at every release in the sequence
- the deprecation counter reads zero over the documented window before the removal PR merges (cite the metrics query / dashboard in the removal PR)
- mixed-version compatibility test: a test that decodes a payload produced by the old contract with the new code and vice versa, e.g. `dotnet test tests/Orders.UnitTests --filter DeprecatedFieldRoundTrip`
- proto removal: the repo's proto-break check passes because the number/name is reserved, not reused
- HTTP: a `WebApplicationFactory` transport test (e.g. `--filter DeprecatedRouteHeaders`) asserting the deprecated route serves the `Deprecation` and `Sunset` headers during the window, and — after the removal release — that the route is gone (`404`, or `410` if clients still probe it)
- DB: apply the drop migration against a real database behind the `-Integration` switch, prove no live query references the dropped column, and confirm via the expand/contract proof in [../services/database.md](../services/database.md) that release N and N+1 code both run against the release-N schema
- `CHANGELOG.md` contains the deprecation entry and, in the removal release, the removal entry with the version tag
