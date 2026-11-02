# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the reverse-engineering handbook itself, not to an engagement workspace.

How to change this handbook without breaking the contract it enforces.

## Default Approach

This handbook is a control plane, not a tutorial. A change is correct only when the fast path ([AGENTS.md](AGENTS.md)), the routing, the slow path ([maintainer-reference.md](maintainer-reference.md)), the recipes, and the templates still agree after it lands.

### Know The Two-Speed Model

- **Fast path — [AGENTS.md](AGENTS.md).** Invariants, the Change Routing table, the verification bar. Terse and authoritative. Refine it; never weaken it.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Lifecycle model, evidence taxonomy, rationale, troubleshooting.

Target-neutral topic depth lives in `foundations/`, `techniques/`, and `operations/`; procedures in `recipes/`; gates in `checklists/`; fill-in artifacts in `templates/`; tool picks and contested stances in `decisions/`. Domain- and platform-specific depth lives in `domains/<domain>/` and `platforms/<platform>/`, each with its own `README.md` index (currently `domains/games/` and `platforms/linux-proton/`). A new rule goes in the core when it holds for any target, and in a section when it is specific to that domain or platform. The core never links down into a section for a *rule*; it names the section and lets the section own the specifics.

### Use The House Shapes

- **Topic docs**: `# Title` → one-line purpose → `## Default Approach` with `###` subsections → `## Common Mistakes And Forbidden Patterns` → `## Verification And Proof` → `## Where To Go Next`.
- **Recipes**: Files To Touch / Steps / Invariants To Preserve / Proof, with concrete commands or MCP calls.
- **Checklists**: `# <Name> Checklist`, one scope sentence, `##` groups of `- [ ]` questions, ending with `## Proof`.
- **Templates**: skeletons with explicit `<placeholder>`s.

Voice: short active sentences, one instruction per sentence in procedures, warnings before instructions.

### Every Claim Needs A Source

Tool behavior changes fast in this domain, and folklore is common. Each factual claim about a tool, engine internal, law, or platform cites a primary source inline: the tool's README or docs, Epic's documentation or engine source path, the statute text, or the vendor announcement. Prefer a short direct quote for load-bearing claims. A claim without a source is removed, not softened.

### Version Pins Live In One Place

Tool versions and supported-engine ranges change monthly. Pins live only in [decisions/tooling.md](decisions/tooling.md) with the date they were checked. Prose names a tool, not its version, unless a rule depends on a version.

### Scope Changes Require A Decision Record

The boundaries in [foundations/scope-legal-and-ethics.md](foundations/scope-legal-and-ethics.md) and the Repo-Wide Invariants in [AGENTS.md](AGENTS.md) do not move by opinion. Record the change in [decisions/README.md](decisions/README.md) with alternatives and consequences first. Content that teaches anti-cheat, DRM, or integrity-check circumvention, cheat development, or attacks on live services is never in scope and is rejected in review.

### Keep The Sync Surfaces In Sync

When a doc is added, renamed, or removed, update in the same change: [README.md](README.md) reading paths and map, the [AGENTS.md](AGENTS.md) Change Routing table, the owning directory index, and [glossary.md](glossary.md) for new terms.

## Common Mistakes And Forbidden Patterns

- Adding a tool recommendation without its source, or pinning its version in prose.
- Copying a technique from a cheat forum without checking it against a primary source and the scope boundaries.
- Adding a runtime technique without routing it through [checklists/runtime-session-gate.md](checklists/runtime-session-gate.md).
- Adding a doc without a routing row, an index entry, and a README link.
- Turning a topic doc into a step-by-step tutorial; procedures belong in `recipes/`.

## Verification And Proof

- All relative links resolve and every cited `### anchor` exists.
- Every added doc is reachable from [README.md](README.md), routed in [AGENTS.md](AGENTS.md), and indexed by its directory README.
- No tool version appears outside [decisions/tooling.md](decisions/tooling.md).
- Every new factual claim has an inline source.

## Where To Go Next

- [maintainer-reference.md](maintainer-reference.md) — the lifecycle model and rationale behind the rules.
- [decisions/README.md](decisions/README.md) — how scope and tooling decisions are recorded.
