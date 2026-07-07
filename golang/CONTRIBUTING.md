# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the handbook itself, not to application repos. It is not part of the app-generation contract.

How to change the Go handbook itself without breaking the contract it exists to enforce.

## Default Approach

This handbook is a control plane, not prose. A change is correct only when the fast path, the routing, the slow path, the recipes, and the worked example still agree with each other after it lands. Edit with that whole-system bar in mind, not one file in isolation.

### Know The Two-Speed Model You Are Editing

The handbook is structured as a two-speed documentation system. Find where your change belongs before you touch a file:

- **Fast path — [AGENTS.md](AGENTS.md).** Repo-wide invariants, change defaults, the Change Routing table (change-type to file-set mapping), and the verification bar. Terse and authoritative. Refine it; never weaken it. Edit the routing table when you add a new kind of change or a new file that a change type should also touch; most other changes are *not* fast-path changes.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Architecture, package map, lifecycle, test taxonomy, and rationale. Edit this when the *why* behind a rule changes or needs fuller background than the fast path carries.

Topic depth lives in `foundations/`, `services/`, `quality/`, and `operations/`; copy-paste scaffolding in `templates/`; step-by-step procedures in `recipes/`; gates in `checklists/`; and the binding decisions in `decisions/`.

### Use The House Templates

Match the existing shape exactly; do not invent a new one.

- **Topic docs** (`foundations/*`, `services/*`, `quality/*`, `operations/*`): `# Title` -> one-line purpose -> `## Default Approach` (with `###` subsections) -> `## Common Mistakes And Forbidden Patterns` -> `## Verification And Proof` -> an optional related-links tail. To deepen a topic, insert `###` subsections in the right place and extend Common Mistakes and Verification; never reorder or duplicate existing content.
- **Recipes** (`recipes/*`): Files To Touch / Steps / Invariants / Proof, with concrete copy-pasteable commands.
- **Checklists** (`checklists/*`): traceable `- [ ]` boxes, each tied to evidence a reviewer can point at.
- **Templates** (`templates/*`): fill-in skeletons with explicit `<PLACEHOLDER>`s, not finished prose.
- **Index/navigational docs** (the glossary, README routing tables) may use their own clear shape.

Voice everywhere: terse, opinionated, contract-not-tutorial. State the standardized engineering decision; this is not a Go language tutorial. Cross-link only files that already exist (or are landing in the same change).

### Keep The Sync Surfaces In Sync

Three sets of files are load-bearing and drift silently. When you change one, change its partners in the same PR:

- **[AGENTS.md](AGENTS.md) <-> [maintainer-reference.md](maintainer-reference.md) <-> recipes.** An invariant added to the fast path needs its rationale in the slow path and its procedure in a recipe; a new recipe needs a routing row and, if it introduces a rule, a fast-path line. None of the three may contradict the others.
- **[AGENTS.md](AGENTS.md) Change Routing rows** must point at files that exist and list the real also-update set for that change type.
- **Templates and the worked example** must reflect the rule they scaffold. If you change a default that `templates/` or `reference/exampleservice/` embodies, update the scaffold too.

### Keep The Reference Service Green

`reference/exampleservice/` is the compiling proof that the handbook's rules are mutually consistent. If your change touches it — or changes a rule it embodies — run its gate and leave it green:

```bash
make verify   # from golang/reference/exampleservice/
```

A handbook change that breaks the reference service is wrong by construction: the example is the contract executed. `make verify` is the gate (tidy/fmt-check/lint/vet/test/race/vuln/build); coverage is a separate `make cover` and is not part of the gate.

### Changing An Invariant Requires An ADR

The repo-wide invariants in [AGENTS.md](AGENTS.md) and the Non-Negotiables in [README.md](README.md) are not edited by opinion. To change one — to relax a default, swap a stdlib-first choice, or alter the verification bar — record an Architecture Decision Record first per [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md), using [templates/adr-template.md](templates/adr-template.md). The ADR captures the alternatives and consequences; only then do the fast path, slow path, recipes, templates, and reference service move together to match it. Library and framework picks are deferred to [decisions/framework-selection.md](decisions/framework-selection.md) — route the choice there rather than hard-coding a vendor into a topic doc.

## Common Mistakes And Forbidden Patterns

- Editing one file in isolation and leaving AGENTS / maintainer-reference / the recipe out of sync.
- Weakening a fast-path invariant in passing instead of through an ADR.
- Inventing a new document shape instead of using the house template for that doc type.
- Cross-linking a file that does not exist yet, or a vendor library that belongs in [decisions/framework-selection.md](decisions/framework-selection.md).
- Turning a contract doc into a Go tutorial — explaining the language instead of stating the standardized decision.
- Changing a default that `templates/` or `reference/exampleservice/` embodies without updating the scaffold and re-running its `make verify`.
- Reordering or duplicating existing sections when deepening a topic doc instead of inserting `###` subsections in place.

## Verification And Proof

- All internal cross-links resolve to files that exist after the change.
- AGENTS.md, maintainer-reference.md, and the affected recipe(s) tell one consistent story; no rule contradicts another.
- If `reference/exampleservice/` (or a rule it embodies) was touched, `make verify` is green there.
- Any change to a repo-wide invariant references an ADR in the PR.
- The PR itself follows [foundations/git-workflow.md](foundations/git-workflow.md): small, Conventional Commits subject, one logical change, green CI, one review.

Related: [foundations/git-workflow.md](foundations/git-workflow.md) for branch/commit/PR mechanics, [templates/project-contributing.md](templates/project-contributing.md) for the downstream-repo CONTRIBUTING template, and [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership changes.
