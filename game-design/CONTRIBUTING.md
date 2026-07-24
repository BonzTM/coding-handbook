# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the handbook itself, not to game projects. It is not part of the design-work contract.

How to change the game design handbook itself without breaking the contract it exists to enforce.

## Default Approach

This handbook is a control plane, not prose. A change is correct only when the fast path, the routing, the slow path, the recipes, and the templates still agree with each other after it lands. Edit with that whole-system bar in mind, not one file in isolation.

### Know The Two-Speed Model You Are Editing

The handbook is structured as a two-speed documentation system. Find where your change belongs before you touch a file:

- **Fast path — [AGENTS.md](AGENTS.md).** Repo-wide invariants, change defaults, the Change Routing table (change-type to file-set mapping), and the proof bar. Terse and authoritative. Refine it; never weaken it. Edit the routing table when you add a new kind of change or a new file a change type should also touch; most other changes are *not* fast-path changes.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Architecture, doc map, lifecycle, and rationale. Edit this when the *why* behind a rule changes or needs fuller background than the fast path carries.

Topic depth lives in `foundations/`, `disciplines/`, `quality/`, and `operations/`; copy-paste scaffolding in `templates/`; step-by-step procedures in `recipes/`; gates in `checklists/`; and the binding decisions in `decisions/`.

### Use The House Templates

Match the existing shape exactly; do not invent a new one.

- **Topic docs** (`foundations/*`, `disciplines/*`, `quality/*`, `operations/*`): `# Title` -> one-line purpose -> `## Default Approach` (with `###` subsections) -> `## Common Mistakes And Forbidden Patterns` -> `## Verification And Proof` -> an optional related-links tail. To deepen a topic, insert `###` subsections in the right place and extend Common Mistakes and Verification; never reorder or duplicate existing content.
- **Recipes** (`recipes/*`): Files To Touch / Steps / Invariants / Proof, with concrete fill-in artifacts named at each step.
- **Checklists** (`checklists/*`): traceable `- [ ]` boxes, each tied to evidence a reviewer can point at — a diagram, a playtest report, a spreadsheet tab, a DDR.
- **Templates** (`templates/*`): fill-in skeletons with explicit `<PLACEHOLDER>`s, not finished prose.
- **Index/navigational docs** (the glossary, README routing tables) may use their own clear shape.

Voice everywhere: terse, opinionated, contract-not-tutorial. State the standardized design decision; this is not a game design textbook. Where a rule leans on canon — Fullerton, Schell, Swink, Schreiber, Hodent, a named GDC talk — cite the source in backticks; do not re-teach it. Engine-version-specific guidance is pinned to the engine's current stable line and links the engine's official docs. Cross-link only files that already exist (or are landing in the same change).

### Keep The Sync Surfaces In Sync

Three sets of files are load-bearing and drift silently. When you change one, change its partners in the same PR:

- **[AGENTS.md](AGENTS.md) <-> [maintainer-reference.md](maintainer-reference.md) <-> recipes.** An invariant added to the fast path needs its rationale in the slow path and its procedure in a recipe; a new recipe needs a routing row and, if it introduces a rule, a fast-path line. None of the three may contradict the others.
- **[AGENTS.md](AGENTS.md) Change Routing rows** must point at files that exist and list the real also-update set for that change type.
- **Templates and their governing docs** must scaffold the same rule. [templates/one-page-gdd.md](templates/one-page-gdd.md) follows [foundations/design-documentation.md](foundations/design-documentation.md) and [recipes/write-a-one-page-gdd.md](recipes/write-a-one-page-gdd.md); [templates/playtest-script.md](templates/playtest-script.md) and [templates/playtest-report.md](templates/playtest-report.md) follow [quality/playtesting.md](quality/playtesting.md) and [recipes/run-a-playtest.md](recipes/run-a-playtest.md); [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md) follows [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md) and [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md); [templates/ddr-template.md](templates/ddr-template.md) follows [decisions/design-decision-records.md](decisions/design-decision-records.md). If you change a default a template embodies, update the template in the same change.

Each glossary term names its owning doc; renaming or re-homing a rule updates [glossary.md](glossary.md) in the same PR.

### Reference Exemplars Are A Later Phase

This handbook has no `reference/` exemplar projects yet; they are a planned later phase. Until they land, the proof surface is link integrity and cross-doc agreement, and templates are the closest thing to executable contract — treat a template that contradicts its governing doc as a broken build. When exemplar projects land, the golang rule applies unchanged: a handbook change that breaks an exemplar is wrong by construction.

### Changing An Invariant Requires A DDR

The repo-wide invariants in [AGENTS.md](AGENTS.md) and the Non-Negotiables in [README.md](README.md) are not edited by opinion. To change one — to relax a default, swap a canon framing, or alter the proof bar — record a Design Decision Record first per [decisions/design-decision-records.md](decisions/design-decision-records.md), using [templates/ddr-template.md](templates/ddr-template.md). The DDR captures the alternatives and consequences; only then do the fast path, slow path, recipes, and templates move together to match it. Framework and model picks — MDA versus alternatives, motivation models, economy notations — are deferred to [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md); route the choice there rather than hard-coding a framework into a topic doc.

## Common Mistakes And Forbidden Patterns

- Editing one file in isolation and leaving AGENTS / maintainer-reference / the recipe out of sync.
- Weakening a fast-path invariant in passing instead of through a DDR.
- Inventing a new document shape instead of using the house template for that doc type.
- Cross-linking a file that does not exist yet, or a framework/model choice that belongs in [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md).
- Turning a contract doc into a design textbook — summarizing Schell or Fullerton at length instead of stating the standardized decision and citing the source.
- Changing a default that a `templates/` file embodies without updating the template in the same change.
- Adding a claim without a source, or citing a source for something it does not say.
- Reordering or duplicating existing sections when deepening a topic doc instead of inserting `###` subsections in place.

## Verification And Proof

- All internal cross-links resolve to files that exist after the change.
- AGENTS.md, maintainer-reference.md, and the affected recipe(s) tell one consistent story; no rule contradicts another.
- Every touched template still matches its governing doc and recipe.
- Any change to a repo-wide invariant references a DDR in the PR.
- Every new non-obvious claim carries a named source; engine-specific guidance links the engine's official docs for its current stable line.
- The PR is one logical change with a Conventional Commits subject, green CI, and one review.

Related: [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership changes, [maintainer-reference.md](maintainer-reference.md) for the doc map this file assumes, and [decisions/design-decision-records.md](decisions/design-decision-records.md) for the DDR mechanics.
