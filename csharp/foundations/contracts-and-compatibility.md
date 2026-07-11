# Contracts and Compatibility

Schemas, APIs, and data contracts are first-class engineering surfaces. Treat them like code, not like incidental documentation.

## Default Approach

Every boundary should have a clear owner and a clear contract.

| Boundary | Contract form | Primary owner |
|---|---|---|
| HTTP API | request and response DTOs, documented status codes, RFC 9457 ProblemDetails error model, optional OpenAPI if the repo publishes one | `Orders.Api/Endpoints/` plus the owning Core service |
| gRPC API | `.proto` files in `api/` and the stubs generated from them | `api/<service>/v1` plus the gRPC endpoint project |
| database | EF Core migrations, query shape, transaction rules | `Orders.Infrastructure/Data/` |
| event or queue payload | explicit envelope record or schema, versioning and idempotency rules | owning publisher and consumer |
| public NuGet API | the library's `public` surface, tracked in `PublicAPI.Shipped.txt`/`PublicAPI.Unshipped.txt` | the library project |

## Wire Contracts

- Minimal API endpoints define stable request and response DTOs at the transport boundary, per [serialization.md](serialization.md).
- Error responses use one consistent shape — RFC 9457 ProblemDetails — not ad hoc JSON per endpoint.
- gRPC services use versioned proto packages from day one; the `.proto` files in `api/` are the source of truth and generated stubs are never hand-edited.
- Generated code policy should be explicit and enforced consistently. See [../services/grpc-services.md](../services/grpc-services.md).

## Data Contracts

- Every schema change ships as an EF Core migration, applied by the explicit migration step — never auto-migrate on normal startup. See [../services/database.md](../services/database.md) and [../recipes/add-migration.md](../recipes/add-migration.md).
- Application code and schema changes need a mixed-version rollout story when the system deploys gradually.
- Queries, indexes, and migration order are part of the contract, not just implementation detail.
- Destructive migrations require an explicit rollback or compatibility plan.

## Event And Message Contracts

- Event payloads are contracts, not internal implementation residue.
- Give each published event one authoritative schema source and one stable event name.
- Use explicit metadata for event ID, type, source, time, and correlation context.
- Treat additive evolution as the default. If meaning changes incompatibly, publish a new contract rather than mutating the old one in place.
- Delivery semantics are part of the contract too: ordering guarantees, retry behavior, idempotency expectations, and DLQ policy should be documented where operators and consumers can find them.

Runtime guidance for publishers, consumers, retries, outbox, inbox, and DLQ behavior lives in [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

## Compatibility Rules

- Additive changes are usually safest: new optional fields, new endpoints, new enum values handled carefully.
- Renames, required-field additions, and destructive schema changes need explicit transition plans.
- Any contract consumed outside one project follows the deprecation policy below before removal.
- Internal contracts can change faster than public ones, but they still need coordinated callers and proof.

### Binary Versus Source Compatibility

A compiled .NET library has two compatibility surfaces, and a change can break one without breaking the other. Both kinds of break are breaking changes for versioning purposes.

- **Binary compatibility**: an assembly compiled against the old version keeps working against the new one without recompiling. This is what an app that transitively references your package actually depends on — NuGet unifies to one version per package in a dependency graph, so a consumer two levels up may load your new assembly against their old compile.
- **Source compatibility**: existing consumer code recompiles against the new version without edits.

| Change to a public surface | Binary | Source |
|---|---|---|
| Add a new type, method, or overload | compatible | usually compatible; a new overload can break via ambiguity |
| Add an optional parameter to an existing method | **breaking** (the old signature is gone; compiled callers bind by exact signature) | compatible |
| Change a public field to a property | **breaking** | compatible |
| Rename a parameter | compatible | **breaking** for callers using named arguments |
| Change the value of a public `const` | silently wrong — the old value is inlined into consumer assemblies at their compile time | compatible-looking, which is worse |
| Add a member to a public interface | **breaking** for implementers (unless a default implementation is provided) | **breaking** |
| Remove or rename a member, change a return type, class↔struct, sealed↔unsealed | **breaking** | **breaking** |

Consequences: prefer overloads over optional parameters on published APIs, prefer `static readonly` over `const` for anything that might ever change, and treat "it still compiles" as insufficient evidence of compatibility. When in doubt, it is breaking.

### Track The Public Surface In The Repo

Libraries adopt `Microsoft.CodeAnalysis.PublicApiAnalyzers`: every public symbol is listed in `PublicAPI.Shipped.txt` and `PublicAPI.Unshipped.txt` committed next to the project. Any change to the public surface fails the build (`RS0016`/`RS0017`) until the text files are updated — which turns every API change into a reviewable diff in the PR instead of an accident discovered by a consumer. At release, unshipped entries move to shipped as part of the release recipe ([../recipes/release-library-version.md](../recipes/release-library-version.md)). Editing the files to silence the analyzer without reviewing the surface change defeats the point.

## Versioning And Deprecation Policy

A published surface is anything a caller depends on that is not internal to one solution: a NuGet package's public API, an HTTP or gRPC API, an event schema. Every published surface is versioned, and nothing published is removed without going through a deprecation window. Internal (non-published) types are exempt and may move as fast as their callers can be updated in the same change — that is what `internal` and the project-reference boundaries are for.

### Semver Discipline

- Published surfaces follow semantic versioning. NuGet resolves and orders by SemVer; the release tag `v1.2.3` is canonical and drives the packed `<Version>` — the tag is the source of truth, not a hand-edited property.
- A MAJOR bump is triggered by any breaking change to a public or wire contract: a binary **or** source break per the table above, removing or renaming a proto field or RPC, changing the meaning of an existing field, or tightening validation that previously accepted valid input. When in doubt, it is breaking.
- MINOR covers backward-compatible additions: a new public type or method, a new optional field, a new endpoint, a new proto field with a fresh tag number.
- PATCH covers backward-compatible fixes with no contract change.
- `0.x.y` signals an unstable surface where MINOR may break. Prerelease suffixes (`1.2.0-preview.1`) signal the same for a specific line. Reaching `1.0.0` is a commitment to the policy above; do not ship a `1.0` you are not ready to keep stable.

### Package Major Versions

- .NET has no counterpart to Go's `/v2` import-path trick. A new major ships under the **same package ID**, and NuGet resolves exactly one version of a package per application graph — so a breaking major forces every consumer in the graph, including transitive ones, to move together.
- That makes breaking majors more expensive than in Go, which is one more reason additive evolution is the default. Consumers defer a major by pinning the old version in `Directory.Packages.props`; they cannot run both.
- Side-by-side majors require a new package ID (`Orders.Client` → `Orders.Client.V2`-style splits). That is a rewrite-scale tool, not the routine mechanism — use it only when a long coexistence window is genuinely required, and record the decision as an ADR.

### Deprecation Markers

- Mark a deprecated .NET symbol with `[Obsolete]`, stating the replacement and the removal target: `[Obsolete("Use CreateOrderAsync; SubmitOrder will be removed in 3.0.")]`. The compiler flags every caller (`CS0618`) — the marker is analyzer-visible by construction, nothing else is needed to surface it. Give the attribute a `DiagnosticId` (and optionally `UrlFormat` pointing at migration notes) so consumers can suppress or track it precisely rather than blanket-silencing `CS0618`.
- Keep the marker a **warning** during the window. Escalating to `[Obsolete(..., error: true)]` is itself a source-breaking change; it belongs in the major that removes the symbol, if it is used at all.
- For HTTP APIs, signal deprecation on the wire as routed in [../decisions/framework-selection.md](../decisions/framework-selection.md): a `Deprecation` header plus a `Sunset` header (RFC 8594) carrying the removal date.
- For proto surfaces, mark the field, RPC, or message with `option deprecated = true;` (`[deprecated = true]` on a field). C# codegen emits `[Obsolete]` on the generated member, so callers get the same compiler warning — but do not rely on that alone; pair it with the announcement step below.
- For events, record the deprecation against the schema's authoritative source and announce it on the same channel consumers watch for the schema.

### Deprecation Window

Removal is a sequence, not a single commit. Every removal of a published contract runs this window:

1. Mark deprecated in code with the marker above and a stated target release for removal.
2. Announce it in the CHANGELOG and release notes for the release that ships the marker. See [../operations/ci-and-release.md](../operations/ci-and-release.md).
3. Measure usage. Removal does not happen until usage is observably zero (telemetry on the deprecated path, request logs, or consumer confirmation), and not before the stated target release.
4. Remove in the stated later release, which is a MAJOR bump if the surface is at 1.0+.

The full mechanics — choosing the target release, instrumenting the deprecated path, and the removal change itself — live in [../recipes/deprecate-and-remove-contract.md](../recipes/deprecate-and-remove-contract.md). This policy defines when the window applies and what each step must produce; the recipe is the executable how.

## Common Mistakes And Forbidden Patterns

- DTOs or proto files that mirror EF entities and database tables instead of transport needs.
- Changing a public payload shape without compatibility review.
- Treating generated stubs as the contract while letting the `.proto` source drift.
- Shipping migrations that assume every process upgrades at once.
- Relying on tribal knowledge instead of one obvious source of truth for the boundary.
- Shipping a binary-breaking change as MINOR because consumer source "still compiles" — adding an optional parameter and converting a field to a property are the classic offenders.
- Changing a public `const` value and expecting consumers to see it without recompiling; the old value is inlined into their assemblies.
- Updating `PublicAPI.Shipped.txt` mechanically to make `RS0016`/`RS0017` go away instead of reviewing whether the surface change is intentional and correctly versioned.
- A bare `[Obsolete]` with no replacement and no removal target, which flags the symbol but gives callers nowhere to go.
- Jumping straight to `[Obsolete(..., error: true)]` or deleting the symbol without the window — removal on the next release because it "felt long enough" instead of waiting for the stated release and observed-zero usage.

## Verification And Proof

- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit; `[Obsolete]` usage inside the repo surfaces as build warnings, which are errors under the gate unless explicitly and narrowly suppressed at the call site that is still allowed to use the old surface
- transport tests that prove request decoding, validation, and response shape, per [serialization.md](serialization.md)
- proto lint or generation checks for gRPC surfaces
- migration apply tests and compatibility review for schema changes
- consumer-focused tests for event or external payload changes
- release notes when contracts, env vars, migrations, or compatibility expectations change
- for libraries: the `PublicAPI.*.txt` diff in the PR matches the intended surface change, and the version bump matches its nature (any binary or source break ⇒ MAJOR)
- a canonical `v1.2.3` tag whose major component matches the nature of the change
- for any removal of a published contract: the deprecation marker in a prior release, the CHANGELOG and release-note entry that announced it, and the usage measurement showing zero before removal
