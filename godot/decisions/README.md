# Decisions

The rules for making and recording hard-to-reverse choices. This handbook keeps the decision *process* here; the big default picks themselves are single-homed in the docs that own them, and this index tells you which doc that is.

This directory holds handbook-level process. Project-level ADRs — the actual records of decisions a specific game or tool made — live in **that project's own `decisions/` directory**, not here. This directory governs how those records are written and which defaults they may safely assume.

## Decision Docs

- [architecture-decision-records.md](architecture-decision-records.md) - the ADR process, the record shape to copy, and when a decision requires an ADR before it merges. Use it whenever a change is non-obvious or hard to reverse — engine version bumps, autoload additions, save-format changes, dependency adoption.

## Decisions Owned Elsewhere

Unlike some handbooks, there is no separate framework-selection doc: each default choice lives in the doc that owns the concern, next to the rules it constrains. Do not restate these picks in an ADR — link to the owner and record only the project-specific deviation.

| Decision | Owning doc |
| --- | --- |
| Scripting language (GDScript vs C#) | [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md) |
| Engine version pin and project baseline | [../foundations/project-setup.md](../foundations/project-setup.md) |
| Test framework | [../quality/testing.md](../quality/testing.md) |
| Lint and format toolchain | [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md) |
| CI export tooling and release flow | [../operations/ci-and-release.md](../operations/ci-and-release.md) |
| Save file format | [../systems/save-and-load.md](../systems/save-and-load.md) |

## Relation To Change Routing

Most changes are not decisions — route them with the Change Routing table in [../AGENTS.md](../AGENTS.md) (## Change Routing) and follow the row. Write an ADR only when no row covers the change or the change would weaken a repo-wide invariant; in that case the ADR, the owning doc, and the affected routing row move together in the same PR, per [../CONTRIBUTING.md](../CONTRIBUTING.md).

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- The ADR process and record shape: [architecture-decision-records.md](architecture-decision-records.md)
- Routing a decision-shaped change: [../AGENTS.md](../AGENTS.md) (## Change Routing)
