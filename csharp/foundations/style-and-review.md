# Style and Review

Idioms and review heuristics that keep C# code obvious, boring, and resilient.

## Default Approach

The `.editorconfig` is law, and `dotnet format` enforces it. Prefer clarity over cleverness; a reviewer should never have to argue taste that a config file already decided.

### Formatting And Configuration

- Copy the `.editorconfig` from the template ([../templates/README.md](../templates/README.md)); it is the single source for style rules, naming rules, and analyzer severities. Style debates end by editing it, not by per-file exceptions.
- `dotnet format --verify-no-changes` is the format gate inside `pwsh ./verify.ps1`; unformatted code does not merge. `EnforceCodeStyleInBuild=true` makes style diagnostics build diagnostics, and `TreatWarningsAsErrors=true` makes them failures — the analyzer setup lives in [../quality/linting.md](../quality/linting.md).
- File-scoped namespaces, always (`namespace Orders.Core;`) — one less indent level on every file, enforced by `.editorconfig`.
- One top-level type per file, file named after the type (`OrderService.cs`). Small private nested types and a type's tightly coupled companions (e.g. an enum used only by that type) may share the file; a second public type may not.
- `usings` are sorted with `System` first (the formatter does this); `ImplicitUsings` stays enabled so files carry only the non-obvious imports.

### Naming

- Types, methods, properties, events, constants: `PascalCase`. Locals and parameters: `camelCase`. Private fields: `_camelCase`.
- Interfaces carry the `I` prefix (`IOrderRepository`); type parameters carry `T` (`TKey`, `TResult`).
- Async methods returning `Task`/`ValueTask`/`IAsyncEnumerable` carry the `Async` suffix (`CreateOrderAsync`), including private ones. The suffix is the call-site signal that an `await` is owed — see [cancellation-and-async.md](cancellation-and-async.md).
- Acronyms follow .NET casing: two-letter acronyms stay upper (`IO`), longer ones are cased as words (`HttpClient`, `JsonSerializer`), and identifiers use `Id`, `Db`, `Ok` — `OrderId`, not `OrderID`.
- No Hungarian notation, no type-encoding prefixes (`strName`, `iCount`), no abbreviations that save three characters at the cost of a grep (`cfg`, `mgr`, `svc`).
- Avoid stutter across the namespace qualifier: `Orders.Core.Order`, not `Orders.Core.OrdersOrder`; a class named after its namespace fights the reader.
- Name length scales with scope: `i` is fine in a three-line loop; anything class-level or long-lived gets a descriptive name.
- The `.editorconfig` naming rules enforce the mechanical part (prefixes, casing) as build diagnostics; the judgment part is review.

### Documentation Standards

Docs are part of the contract. Treat them with the same rigor as the code they describe.

- Every public type and member carries an XML doc comment (`///`) that states the contract — preconditions, exceptions thrown, null behavior, thread safety, ownership of returned values — not a restatement of the signature. If `<summary>` only repeats the name, delete it or replace it with the contract.
- Libraries set `GenerateDocumentationFile=true`, which both ships the IntelliSense XML and turns missing public docs into `CS1591` diagnostics — under warnings-as-errors, an undocumented public member fails the build. Services document their public seams (Core interfaces, DTOs, shared constructs); endpoint lambdas and internals document themselves through naming and tests.
- Use the structural tags where they carry contract: `<param>` for non-obvious parameters, `<returns>` when the shape needs explaining, `<exception cref="...">` for every deliberately thrown exception type, `<remarks>` for usage constraints. `<inheritdoc/>` on implementations keeps the contract stated once, on the interface.
- Prefer `<see cref="..."/>` links over prose type names; the compiler verifies them, so docs cannot silently dangle after a rename.
- Deprecate in-band with `[Obsolete("...; use X.")]` so every call site gets an analyzer-visible warning — prose or changelog-only deprecation reaches nobody. The full deprecation and removal sequence lives in [contracts-and-compatibility.md](contracts-and-compatibility.md).
- Doc-as-code: architecture notes, design rationale, and diagrams live in the repo (ADRs per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)), versioned and reviewed in the same PR as the change they describe. A wiki or slide deck that drifts from `main` is worse than no doc.

### API Shape

- Accept the narrowest abstraction the caller benefits from (`IReadOnlyList<T>`, an `IOrderRepository` port); return concrete or well-known types you own the implementation of.
- Use constructors (or factories) when invariants must hold or dependencies must be injected — see [data-modeling.md](data-modeling.md); dependencies arrive by constructor injection, never a service locator.
- Keep boolean parameters rare; `Publish(order, true, false)` is unreadable at the call site. Prefer a two-value enum or a second method name.
- Default to the least visibility: `internal` unless a caller outside the assembly needs it, `sealed` unless designed for inheritance. `InternalsVisibleTo` is for the matching test project only, per [solution-and-project-design.md](solution-and-project-design.md).

### Expression Bodies And var

- Expression-bodied members (`=>`) only when the body is a single expression that fits on one line and reads as a definition (`public int Count => _lines.Count;`). The moment it needs a second clause, a ternary-inside-a-ternary, or a line wrap, use a block body. Expression bodies are for *what it is*, not for compressing *what it does*.
- `var` when the type is apparent from the right-hand side (`var order = new Order(...)`, `var lines = order.Lines`); the explicit type when it is not (`decimal total = Subtotal(order);` — a bare method call tells the reader nothing). The `.editorconfig` encodes this stance; do not fight it file by file.

### Pattern Matching And Control Flow

- Keep the happy path at the lowest indent with guard clauses and early returns; `ArgumentNullException.ThrowIfNull(x)` and `ArgumentOutOfRangeException.ThrowIfNegative(n)` are the standard one-line guards.
- Prefer pattern matching over type-check-and-cast pairs: `if (payment is CardPayment card)` rather than an `is` check followed by a cast, and property patterns (`order is { Status: OrderStatus.Pending, Lines.Count: > 0 }`) over chained accessor comparisons when they read better — *when they read better*, not as a cleverness contest.
- Prefer a `switch` expression over an `if`/`else` chain when mapping one value to another over a closed set. Make the default arm a loud failure (`_ => throw new UnreachableException($"Unhandled status {status}")`), never a silent fallback value that hides a new enum member.
- Nesting more than two or three levels deep is a signal to extract a method or invert the condition.

### Immutability And Defensive Copies

Copy and mutation semantics are part of the contract, not an afterthought.

- Prefer immutable value objects: `sealed record` / `readonly record struct`, validated once at construction, changed with `with` rather than mutated in place. The type-level rules live in [data-modeling.md](data-modeling.md).
- Never expose or store a live mutable collection across a public boundary: expose `IReadOnlyList<T>`, copy caller-supplied collections on the way in ([data-modeling.md](data-modeling.md) has the canonical example).
- Mutable structs are forbidden; a struct that needs mutation is a class. Mark structs `readonly` so the compiler proves it.
- No mutable static state for config, caches, or services. `static readonly` is for genuinely immutable data (lookup tables, singletons without state). Everything else is constructed and wired through DI in `Program.cs` — see [configuration.md](configuration.md) and [solution-and-project-design.md](solution-and-project-design.md).
- Shared mutable state that must exist is owned by one type, guarded internally, and never exposed as a field; the concurrency rules live in [cancellation-and-async.md](cancellation-and-async.md).

### LINQ Readability Limits

LINQ is for making a transformation *read as* a transformation. It stops earning its place the moment a reviewer has to simulate it.

- Keep chains short — roughly three operators. Beyond that, name intermediate results with local variables or extract a well-named method.
- No side effects inside LINQ operators. `Select`/`Where` lambdas that mutate state, log, or call I/O are loops in disguise — write the loop.
- Materialize deliberately and once: end a query with `ToList()`/`ToArray()` when it will be enumerated more than once or leaves the method. Returning a lazy `IEnumerable<T>` built over a `DbContext` or an open resource is a deferred-execution bug waiting for its moment.
- Method syntax is the default; query syntax only where joins make it genuinely clearer.
- In measured hot paths, a plain loop that avoids allocation is fine — clarity of intent first, then measured performance; never dogma in either direction.
- LINQ against EF Core's `IQueryable` is a different contract (it compiles to SQL); those rules live in [../services/database.md](../services/database.md).

## Common Mistakes And Forbidden Patterns

- Style nits argued in review that `.editorconfig` + `dotnet format` should settle mechanically — fix the config, not the PR thread.
- Block-scoped namespaces, multiple public types per file, or files named unlike their type.
- Swallowed exceptions (`catch { }` or catch-log-continue without handling) — failure handling rules live in [errors-and-logging.md](errors-and-logging.md).
- Async methods without the `Async` suffix, or `async void` outside event handlers ([cancellation-and-async.md](cancellation-and-async.md)).
- Public members with no XML doc contract, or `<summary>` comments that restate the name (`/// <summary>Gets the name.</summary>` on `Name`).
- Mutable static/global state, service-locator lookups (`IServiceProvider` passed around as a dependency), or configuration read from ambient statics.
- Mutable structs, or exposing `List<T>`/live internal collections across a public boundary.
- A `switch` on a closed set with a silent default arm that turns a new enum member into a wrong answer instead of a loud failure.
- Ten-operator LINQ chains, LINQ with side effects, or repeated enumeration of a lazy query.
- Speculative abstraction: interfaces with one implementation and no seam purpose, generic type parameters before a second concrete use, "manager"/"helper" grab-bag classes.
- Regions (`#region`) used to fold a class that should have been split.
- Marking something obsolete in prose or a changelog without the in-band `[Obsolete]` attribute, so call sites get no warning.

## Review Questions

| Question | What it catches |
|---|---|
| Is the project boundary still clean — did wire, persistence, or framework types leak into `Orders.Core`? | dependency-direction erosion |
| Does the API expose only what callers need (visibility, abstraction level, no live internals)? | accidental surface, mutation leaks |
| Are error paths, logging, and telemetry consistent with the rest of the repo? | one-off failure handling |
| Would a new contributor find this behavior by reading the call site and the type names? | misplaced logic, misleading names |
| Did the change add complexity a smaller refactor or simpler type could avoid? | speculative abstraction |
| If a new enum member or derived type appears next month, does this code fail loudly or drift silently? | silent default arms, open switches |
| Can the tests for this change fail for the right reason? | tests pinned to implementation detail |

## Verification And Proof

```powershell
dotnet format --verify-no-changes   # the format/style gate on its own
pwsh ./verify.ps1                   # restore (locked), format-check, build (warnings-as-errors), test, audit
```

The build is the style gate: with `EnforceCodeStyleInBuild=true` and warnings-as-errors, `.editorconfig` violations, naming-rule breaches, missing-doc `CS1591` (where enabled), and analyzer findings all fail `pwsh ./verify.ps1` — style is not a separate, skippable step.

For public APIs, also read the XML docs as if you were a first-time caller: every public type should state a purpose, every member a contract — preconditions, exceptions, ownership. If the contract is unclear in IntelliSense, it will be unclear in code review too. Confirm `<see cref>` links compile and documented exceptions match what the code throws; a doc that lies is worse than no doc.
