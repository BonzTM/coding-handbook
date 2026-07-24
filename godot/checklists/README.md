# Checklists

Executable gates for the moments where missing a step is expensive. Each checklist is a grouped set of `- [ ]` items where every box is a question or verifiable assertion a reviewer can tie to evidence — a file, a setting, a passing command — plus a closing Proof section. Run it top to bottom and do not skip items.

For routing a change to its related obligations, see the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## Project Checklists

- [new-project.md](new-project.md) - bootstrapping a fresh Godot 4.x repo from the templates to a verify-green start: project settings, `.gitignore` covering `.godot/`, lint config, tests, and headless CI export. Governed by [../foundations/project-setup.md](../foundations/project-setup.md) and [../operations/ci-and-release.md](../operations/ci-and-release.md).
- [pr-review.md](pr-review.md) - review gate for Godot changes that affect scenes, scripts, contracts, or the exported build: scene and resource diffs, signal and scene contracts, export and settings edits, and the proof the PR must carry. Fixes route through the Change Routing table in [../AGENTS.md](../AGENTS.md).
- [release.md](release.md) - release-cut gate run on the release commit before the tag is pushed and again before any artifact reaches players: version and tag agreement, the CI gate, artifact smoke tests, and save compatibility. Governed by [../operations/ci-and-release.md](../operations/ci-and-release.md).
- [handoff.md](handoff.md) - ownership-transfer gate walked with both the outgoing and incoming owner present: access and credentials, engine and toolchain pins, signing keys and secrets, and addon and asset licenses. Governed by [../onboarding-and-handoff.md](../onboarding-and-handoff.md).

Any future checklist follows the house checklist shape owned by [../CONTRIBUTING.md](../CONTRIBUTING.md) (### Use The House Templates).

## Relation To Change Routing

Checklists and the routing table answer different questions. The Change Routing table in [../AGENTS.md](../AGENTS.md) (## Change Routing) routes a *change type* to the files it touches and the proof it needs. A checklist gates a *moment* — starting a repo, cutting a release — regardless of which change types led up to it. When a checklist box fails, route the fix through the table; do not patch it ad hoc.

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Routing a change to the right files: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- Recipes for implementation steps: [../recipes/README.md](../recipes/README.md)
- Skeleton files a new project starts from: [../templates/README.md](../templates/README.md)
