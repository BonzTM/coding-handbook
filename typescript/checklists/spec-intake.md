# Checklist: Spec Intake

Resolve these decisions before implementation. Link every answer to the spec, issue, ADR, or written product decision.

## Shape And Contract

- [ ] Repository shape is named: HTTP service, worker, CLI, library, React app, or an approved combination.
- [ ] Users, operators, and independently deployed consumers are identified.
- [ ] Success behavior, negative outcomes, and stable contract surfaces are written.
- [ ] Compatibility window and deprecation expectations are recorded.
- [ ] Runtime is Node 24 LTS, one npm package, ESM, unless an accepted ADR says otherwise.

## Trust And Data

- [ ] Every input boundary, maximum size/count, and Zod parsing owner is listed.
- [ ] Authentication source and server-side authorization rules are explicit.
- [ ] Tenant isolation and horizontal/vertical privilege cases are specified.
- [ ] Data classification, residency, retention, deletion, and audit requirements are stated.
- [ ] Secrets, rotation owner, and frontend-public configuration are distinguished.

## Dependencies And Failure

- [ ] PostgreSQL, external services, messages, and browser APIs are named with owners.
- [ ] Timeout, cancellation, retry, idempotency, concurrency, and overload behavior are resolved.
- [ ] Transaction, consistency, ordering, duplication, and replay semantics are explicit.
- [ ] Expected degraded behavior and rollback constraints are written.
- [ ] Any non-default framework or hard-to-reverse decision has an ADR owner.

## User And Operations

- [ ] Accessibility, keyboard, focus, loading, empty, error, and success requirements are defined.
- [ ] SLI population, SLO target/window, capacity, and alertable symptoms are named.
- [ ] Rollout stages, observation window, abort criteria, and recovery owner are stated.
- [ ] Test boundaries and required real dependencies are identified.

## Proof

- [ ] A reviewer can route every requested change through [../AGENTS.md](../AGENTS.md).
- [ ] No unresolved decision would materially change architecture, security, compatibility, or data handling.
- [ ] Open assumptions have owners and deadlines; implementation does not silently choose them.
