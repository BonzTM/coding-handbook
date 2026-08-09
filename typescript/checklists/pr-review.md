# Checklist: PR Review

Each box requires a diff, test, CI result, artifact, or linked decision.

## Intent And Architecture

- [ ] PR states the user/operator outcome, scope, and excluded work.
- [ ] Change Routing files and all sync surfaces are present in the diff.
- [ ] `api -> core <- db`, feature ownership, and composition boundaries remain intact.
- [ ] New dependency, public API, persistence, background work, or network behavior is justified.
- [ ] Any invariant or hard-to-reverse change links an accepted ADR.

## Correctness And Risk

- [ ] External input is bounded and parsed from `unknown` before domain use.
- [ ] Authentication, resource authorization, tenant isolation, and redaction are correct where applicable.
- [ ] Async work has owner, timeout, cancellation, concurrency bound, and observed failure.
- [ ] SQL is parameterized, rows parsed, transactions cleaned up, and migrations rollout-safe.
- [ ] React behavior is semantic, keyboard accessible, and cleans up effects.
- [ ] Compatibility, idempotency, retry, partial failure, and rollback are addressed.

## Maintainability And Proof

- [ ] Functions have one primary responsibility and touched large functions were split or justified.
- [ ] Names, types, comments, and exports state domain meaning without unsafe assertions.
- [ ] Focused tests cover success plus material negative and cancellation paths.
- [ ] Real boundary proof exists for PostgreSQL, emitted ESM, packed library, or deployed route as applicable.
- [ ] Lockfile, generated, schema, migration, config, changelog, and runbook diffs are reviewed.
- [ ] Suppressions are exact, local, explained, and no `@ts-ignore` or broad disable was added.

## Proof

- [ ] CI clean install and `npm run verify` are green with zero warnings.
- [ ] Required integration, security, accessibility, and rollout evidence is linked.
- [ ] Reviewer can state how failure is detected, handled, observed, and recovered.
