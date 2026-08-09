# Checklist: Dependency Upgrade

Use with [the bump recipe](../recipes/bump-dependency.md).

## Scope And Evidence

- [ ] Current/target versions, direct/transitive status, release type, and reason are recorded.
- [ ] Official release notes, migration guide, advisories, Node/ESM/browser support, and peer ranges are reviewed.
- [ ] Major or architecture-changing upgrades answer framework approval questions and link an ADR where required.
- [ ] Maintenance, license, install scripts, native binaries, permissions, and transitive impact are acceptable.
- [ ] Security urgency and exploit reachability are stated without hiding the normal proof gate.

## Change

- [ ] Exact direct version is in `package.json`; npm produced the lockfile change.
- [ ] `git diff -- package.json package-lock.json` contains no unexplained package movement.
- [ ] No unexpected Node, npm, module mode, package manager, or build-system change rides along.
- [ ] Caller and adapter changes preserve boundary validation, cancellation, and error behavior.
- [ ] Deprecated API uses and temporary compatibility shims have removal owners.

## Proof

- [ ] `npm ci` succeeds from clean state without lockfile rewrite.
- [ ] `npm ls <package>` resolves the intended version and valid peer graph.
- [ ] Targeted success, negative, runtime ESM, and type compatibility tests pass.
- [ ] `npm audit --audit-level=high` passes or an approved time-bounded exception is linked.
- [ ] `npm run verify` is green on Node 24.
- [ ] Required image, packed-consumer, browser, PostgreSQL, or external-boundary smoke test passes.

## Merge And Follow-Up

- [ ] Rollout and rollback plan match the dependency's behavioral risk.
- [ ] Monitoring owner and post-release observation window are named.
- [ ] Bot PR receives the same review and evidence as a human upgrade.
