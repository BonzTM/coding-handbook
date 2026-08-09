# Checklists

Reviewer-visible gates for TypeScript project lifecycle and risk. Each checked box must point to repository, CI, deployment, or operational evidence.

| Checklist | Use it when | Governing docs |
|---|---|---|
| [spec-intake.md](spec-intake.md) | Resolve project and feature intent before implementation. | [README](../README.md) |
| [new-project.md](new-project.md) | Bootstrap a repository from the house templates. | [Project setup](../foundations/project-setup.md) |
| [pr-review.md](pr-review.md) | Review any proposed change. | [Style and review](../foundations/style-and-review.md) |
| [feature-definition-of-done.md](feature-definition-of-done.md) | Close a feature with behavior and operations proven. | [Testing](../quality/testing.md) |
| [security-review.md](security-review.md) | Review a trust, identity, data, or supply-chain boundary. | [Security](../operations/security.md) |
| [release.md](release.md) | Publish or deploy a verified artifact. | [CI and release](../operations/ci-and-release.md) |
| [rollout-and-slo-readiness.md](rollout-and-slo-readiness.md) | Approve a production rollout. | [Operability](../operations/operability.md) |
| [dependency-upgrade.md](dependency-upgrade.md) | Upgrade a direct or transitive package. | [Framework selection](../decisions/framework-selection.md) |
| [incident-response.md](incident-response.md) | Manage detection through follow-up. | [Operability](../operations/operability.md) |
| [handoff.md](handoff.md) | Transfer repository and on-call ownership. | [Onboarding and handoff](../onboarding-and-handoff.md) |

Unchecked items are open work. Record an approved, owned, expiring exception rather than silently skipping a gate.
