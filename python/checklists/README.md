# Checklists

Executable gates for moments where omission is expensive: intake, project creation, review, release, rollout, incident response, dependency maintenance, and ownership transfer. Each checklist is a grouped set of traceable `- [ ]` evidence plus a closing Verification section; run it top to bottom and do not skip applicable items.

For change obligations, use the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## Lifecycle

- [spec-intake.md](spec-intake.md) - resolves the WHAT decisions before code: shape, identity, tenancy, data, integrations, runtime, compliance, and SLOs.
- [new-project.md](new-project.md) - builds a fresh `src/`-layout repo from templates to a `make verify`-green baseline.
- [pr-review.md](pr-review.md) - reviews behavior, boundaries, typing, async ownership, and operational safety before merge.
- [release.md](release.md) - cuts a versioned package or container release with verified artifacts.
- [handoff.md](handoff.md) - transfers access, operations, decisions, and proof to a new owner.

## Operations

- [incident-response.md](incident-response.md) - works an active incident from acknowledgement through postmortem.
- [rollout-and-slo-readiness.md](rollout-and-slo-readiness.md) - proves an artifact is safe to place in front of traffic.
- [dependency-upgrade.md](dependency-upgrade.md) - upgrades Python, packages, and tools with an understood lock diff.

## Quality And Security

- [feature-definition-of-done.md](feature-definition-of-done.md) - confirms behavior is complete, tested, observable, compatible, and documented.
- [security-review.md](security-review.md) - gates a security-sensitive boundary or supply-chain change.

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Change routing: [../AGENTS.md](../AGENTS.md) (`## Change Routing`)
- Implementation procedures: [../recipes/README.md](../recipes/README.md)
