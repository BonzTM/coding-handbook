# Checklists

Executable gates for the moments where missing a step is expensive: starting a repo, reviewing a PR, cutting a release, handing off ownership, responding to an incident, rolling out, upgrading a dependency, and closing out a feature or security review. Each checklist is a grouped set of `- [ ]` items plus a closing Verification (or Proof) section; run it top to bottom and do not skip items.

For routing a change to its related obligations, see the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## Lifecycle

- [spec-intake.md](spec-intake.md) - pre-flight checklist resolving the WHAT decisions (shape, auth, tenancy, data, integration, runtime, SLOs) before any code is written — from the spec, by asking, or from its documented defaults — so any spec yields a one-shot.
- [new-project.md](new-project.md) - bootstrapping a fresh repo from the templates to a `pwsh ./verify.ps1`-green start.
- [pr-review.md](pr-review.md) - reviewing or self-reviewing a change before merge.
- [release.md](release.md) - cutting a release with a canonical tag, changelog, and verified artifacts.
- [handoff.md](handoff.md) - transferring project ownership so the new owner can operate it unaided.

## Operations

- [incident-response.md](incident-response.md) - working an active incident from detection through resolution and follow-up.
- [rollout-and-slo-readiness.md](rollout-and-slo-readiness.md) - confirming a service is safe to roll out and has SLOs, alerts, and a runbook.
- [dependency-upgrade.md](dependency-upgrade.md) - upgrading dependencies safely with an understood diff and clean proof.

## Quality And Security

- [feature-definition-of-done.md](feature-definition-of-done.md) - confirming a feature is genuinely complete, tested, observable, and documented.
- [security-review.md](security-review.md) - reviewing a security-sensitive boundary before it ships.

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Routing a change to the right files: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- Recipes for implementation steps: [../recipes/README.md](../recipes/README.md)
