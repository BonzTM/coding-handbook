# PR Review Checklist

Review checklist for Go changes that affect behavior, boundaries, or operational safety.

## Boundaries And Placement

- [ ] Does the code live in the right package, or is it leaking across transport, core, and storage boundaries?
- [ ] Is `main` still thin after the change?
- [ ] Did the change avoid turning `pkg/`, `util`, or `helpers` into a dumping ground?

## Correctness

- [ ] Does every I/O path take `context.Context` explicitly?
- [ ] Are errors wrapped and still matchable with `errors.Is` or `errors.As` where needed?
- [ ] Are goroutines supervised, cancellable, and covered by race-safe tests?
- [ ] Did the change avoid global mutable state and hidden configuration lookups?

## Observability And Operations

- [ ] Does new runtime behavior add the right logs, metrics, or readiness behavior?
- [ ] Are secret values kept out of logs and examples?
- [ ] If a new dependency was added, is the rationale explicit and consistent with [../decisions/framework-selection.md](../decisions/framework-selection.md)?
- [ ] If events or messages changed, is the payload contract still compatible and is idempotency or replay behavior still correct?

## Proof

- [ ] Targeted tests prove the actual behavior change.
- [ ] `make verify` is green: tidy, fmt-check, lint, vet, test, race, vuln, and build all pass.
- [ ] `golangci-lint` passes via `make lint` (or as part of `make verify`) per [../quality/linting.md](../quality/linting.md), with no new `//nolint` that lacks a justification.
- [ ] Coverage did not regress, and mandatory paths are exercised — domain core logic, error-to-status mapping, and request/message decode paths per [../quality/testing.md](../quality/testing.md).
- [ ] `go test -race ./...` or an appropriately scoped race-safe equivalent passed.
- [ ] For DB or external boundaries, at least one real integration path was exercised.
- [ ] For eventing changes, duplicate-delivery, retry, and terminal-failure behavior were actually exercised.
- [ ] If the change ships a feature, it meets every gate in [feature-definition-of-done.md](feature-definition-of-done.md).
