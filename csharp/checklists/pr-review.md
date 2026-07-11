# PR Review Checklist

Review checklist for .NET changes that affect behavior, boundaries, or operational safety.

## Boundaries And Placement

- [ ] Does the code live in the right project — endpoints in `<App>.Api`, domain rules in `<App>.Core`, EF Core and external clients in `<App>.Infrastructure` — or is it leaking across transport, core, and storage boundaries per [../foundations/solution-and-project-design.md](../foundations/solution-and-project-design.md)?
- [ ] Is `Program.cs` still a thin composition root after the change?
- [ ] Did the change avoid turning a `Common`, `Utils`, or `Helpers` namespace into a dumping ground per [../foundations/shared-constructs.md](../foundations/shared-constructs.md)?

## Correctness

- [ ] Does every async I/O path take a `CancellationToken` and flow it to the leaf call per [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md)?
- [ ] Are failures still explicit and matchable — exceptions preserve the inner exception, boundary code maps them to `ProblemDetails` per [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md), and no `catch` swallows what it cannot handle?
- [ ] Is background and concurrent work supervised — `BackgroundService` honors `stoppingToken`, no fire-and-forget tasks, no `async void` outside event handlers, no sync-over-async (`.Result`/`.Wait()`) — and covered by deterministic tests?
- [ ] Did the change avoid global mutable state and hidden configuration or clock lookups (options injected, `TimeProvider` injected, never `DateTime.Now` per [../foundations/time.md](../foundations/time.md))?

## Observability And Operations

- [ ] Does new runtime behavior add the right logs, metrics, or readiness behavior per [../operations/observability.md](../operations/observability.md)?
- [ ] Are secret values kept out of logs, exception messages, and examples?
- [ ] If a new dependency was added, is the rationale explicit and consistent with [../decisions/framework-selection.md](../decisions/framework-selection.md), is the version pinned in `Directory.Packages.props`, and was the `packages.lock.json` diff reviewed?
- [ ] If events or messages changed, is the payload contract still compatible and is idempotency or replay behavior still correct?

## Proof

- [ ] Targeted tests prove the actual behavior change.
- [ ] `pwsh ./verify.ps1` is green: restore (locked), format-check, build (warnings-as-errors), test, audit all pass.
- [ ] The lint gate holds per [../quality/linting.md](../quality/linting.md): `dotnet format --verify-no-changes` is clean and the build passes with `-warnaserror` at `AnalysisLevel=latest-all` — with no new `#pragma warning disable`, `[SuppressMessage]`, or `.editorconfig` severity downgrade that lacks a justification.
- [ ] Coverage did not regress, and mandatory paths are exercised — domain core logic, error-to-ProblemDetails mapping, and request decode/validation paths per [../quality/testing.md](../quality/testing.md).
- [ ] Concurrency-sensitive changes are exercised with deterministic tests: tasks awaited explicitly, time driven by `FakeTimeProvider`, no `Task.Delay` to "let it settle".
- [ ] For DB or external boundaries, at least one real integration path was exercised (Testcontainers, `WebApplicationFactory`) per [../quality/testing.md](../quality/testing.md).
- [ ] For eventing changes, duplicate-delivery, retry, and terminal-failure behavior were actually exercised.
- [ ] If the change ships a feature, it meets every gate in [feature-definition-of-done.md](feature-definition-of-done.md).
