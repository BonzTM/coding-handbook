# Decisions

Index of binding decisions for the reverse-engineering handbook, and how to record a new one.

## Current Decisions

- [tooling.md](tooling.md) — default tool per job, verified versions, and supported-engine ranges. The only file that pins versions.
- Scope boundaries — owned by [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md). The "Never In Scope" list is a fixed decision of this handbook: anti-cheat, DRM, and integrity-check circumvention, cheat and bot development, and live-service tampering are excluded and do not change by decision record.

## Recording A Decision

A decision record is a short dated section appended to the owning file (or a new file in this directory for a cross-cutting choice), with:

1. **Context** — what changed: a tool broke on an engine version, a better tool appeared, a legal or platform fact changed.
2. **Decision** — the new default or rule, in one sentence.
3. **Alternatives** — what else was considered and why it lost.
4. **Consequences** — which docs, recipes, templates, and checklists change with it.
5. **Sources** — primary sources for every factual claim.

Land the record and the doc changes it requires in the same change, per [../CONTRIBUTING.md](../CONTRIBUTING.md).

## Where To Go Next

- [../AGENTS.md](../AGENTS.md) — the Change Routing row for decisions.
- [../maintainer-reference.md](../maintainer-reference.md) — rationale for the current defaults.
