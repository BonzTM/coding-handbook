# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the handbook itself, not to application repos. It is not part of the app-generation contract.

How to change the C# handbook itself without breaking the contract it exists to enforce.

## Default Approach

This handbook is a control plane, not prose. A change is correct only when the fast path, the routing, the slow path, the recipes, and the templates still agree with each other after it lands. Edit with that whole-system bar in mind, not one file in isolation.

### Know The Two-Speed Model You Are Editing

The handbook is structured as a two-speed documentation system. Find where your change belongs before you touch a file:

- **Fast path — [AGENTS.md](AGENTS.md).** Repo-wide invariants, change defaults, the Change Routing table (change-type to file-set mapping), and the verification bar. Terse and authoritative. Refine it; never weaken it. Edit the routing table when you add a new kind of change or a new file that a change type should also touch; most other changes are *not* fast-path changes.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Architecture, project map, lifecycle, test taxonomy, and rationale. Edit this when the *why* behind a rule changes or needs fuller background than the fast path carries.

Topic depth lives in `foundations/`, `services/`, `quality/`, and `operations/`; copy-paste scaffolding in `templates/`; step-by-step procedures in `recipes/`; gates in `checklists/`; and the binding decisions in `decisions/`.

### Use The House Templates

Match the existing shape exactly; do not invent a new one.

- **Topic docs** (`foundations/*`, `services/*`, `quality/*`, `operations/*`): `# Title` -> one-line purpose -> `## Default Approach` (with `###` subsections) -> `## Common Mistakes And Forbidden Patterns` -> `## Verification And Proof` -> an optional related-links tail. To deepen a topic, insert `###` subsections in the right place and extend Common Mistakes and Verification; never reorder or duplicate existing content.
- **Recipes** (`recipes/*`): Files To Touch / Steps / Invariants To Preserve / Proof, with concrete copy-pasteable commands.
- **Checklists** (`checklists/*`): traceable `- [ ]` boxes, each tied to evidence a reviewer can point at, plus a closing Verification command block.
- **Templates** (`templates/*`): fill-in skeletons with explicit `<placeholder>`s, not finished prose; filenames encode destinations per [templates/README.md](templates/README.md).
- **Index/navigational docs** (the glossary, README routing tables) may use their own clear shape.

Voice everywhere: terse, opinionated, contract-not-tutorial. State the standardized engineering decision; this is not a C# language tutorial. Cross-link only files that already exist (or are landing in the same change).

### Keep The Routing Surfaces In Sync

A new file that only exists is invisible; the handbook's entry points must route to it. When you add, rename, or remove a doc, update every routing surface in the same PR:

- **[README.md](README.md)** — the Reading Paths table and/or the Handbook Map must reach the file.
- **[AGENTS.md](AGENTS.md) Change Routing** — the change types the file governs get a row (or an updated row) pointing at it; every row must point at files that exist, cite anchors (`### Section`) that exist, and list the real also-update set.
- **The directory index** — `recipes/README.md`, `checklists/README.md`, `decisions/README.md`, and `templates/README.md` each index their own directory; a file missing from its index is a defect.

Beyond routing, three sets of files are load-bearing and drift silently. When you change one, change its partners in the same PR:

- **[AGENTS.md](AGENTS.md) <-> [maintainer-reference.md](maintainer-reference.md) <-> recipes.** An invariant added to the fast path needs its rationale in the slow path and its procedure in a recipe; a new recipe needs a routing row and, if it introduces a rule, a fast-path line. None of the three may contradict the others.
- **Topic docs and the templates that scaffold them.** If you change a default that `templates/` embodies — a `Directory.Build.props` flag, a `verify.ps1` stage, the Dockerfile shape — update the template and its governing doc together.
- **The verify-gate wording.** Every doc that references the gate says "run `pwsh ./verify.ps1`" with the stages "restore (locked), format-check, build (warnings-as-errors), test, audit". Do not paraphrase it into drift.

### Templates Carry The Pins; Prose Does Not

Exact versions — SDK line, package versions, action majors, base image tags — live ONLY in `templates/` files. Prose docs say "copy from the template" and never state a version. When a template pin changes, nothing in prose should need editing; if it does, the prose was wrong. Verify any new pin against the upstream source before committing it.

### Link Hygiene

- Relative markdown links only, from the file's own directory, to files that exist in this tree.
- Never link `reference/` — compiling reference modules are phase 2 and the directory does not exist. Saying they are planned is fine; a link is a broken promise.
- Cite section anchors in the form `(### Section Name)` next to the file link, and only for headings that actually exist — grep before citing.
- A vendor or library pick belongs in [decisions/framework-selection.md](decisions/framework-selection.md); do not hard-code one into a topic doc.

### Changing An Invariant Requires An ADR

The repo-wide invariants in [AGENTS.md](AGENTS.md) and the Non-Negotiables in [README.md](README.md) are not edited by opinion. To change one — to relax a default, swap a platform-first choice, or alter the verification bar — record an Architecture Decision Record first per [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md), using [templates/adr-template.md](templates/adr-template.md). The ADR captures the alternatives and consequences; only then do the fast path, slow path, recipes, and templates move together to match it. Library and framework picks are deferred to [decisions/framework-selection.md](decisions/framework-selection.md) — route the choice there rather than hard-coding a vendor into a topic doc.

## Common Mistakes And Forbidden Patterns

- Editing one file in isolation and leaving AGENTS / maintainer-reference / the recipe out of sync.
- Adding a new doc without touching the routing surfaces: README reading paths or map, the AGENTS Change Routing table, and the owning directory's README index.
- Weakening a fast-path invariant in passing instead of through an ADR.
- Inventing a new document shape instead of using the house template for that doc type.
- Cross-linking a file that does not exist yet, linking `reference/` before phase 2 lands, or citing a `###` anchor the target file does not contain.
- Writing a version pin into prose, or restating the verify-gate stages in new words instead of the canonical wording.
- Turning a contract doc into a C# tutorial — explaining the language instead of stating the standardized decision.
- Changing a default that `templates/` embodies without updating the template in the same PR.
- Reordering or duplicating existing sections when deepening a topic doc instead of inserting `###` subsections in place.

## Verification And Proof

- All internal cross-links resolve to files that exist after the change, and every cited `### anchor` exists in its target file (grep for the heading).
- Every added or renamed doc is reachable from [README.md](README.md), routed in [AGENTS.md](AGENTS.md), and indexed by its directory README.
- AGENTS.md, maintainer-reference.md, and the affected recipe(s) tell one consistent story; no rule contradicts another.
- No version pin appears outside `templates/`; any changed pin was verified upstream.
- Any change to a repo-wide invariant references an ADR in the PR.
- The PR itself follows [foundations/git-workflow.md](foundations/git-workflow.md): small, Conventional Commits subject, one logical change, green CI, one review.

Related: [foundations/git-workflow.md](foundations/git-workflow.md) for branch/commit/PR mechanics, [templates/project-contributing.md](templates/project-contributing.md) for the downstream-repo CONTRIBUTING template, and [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership changes.
