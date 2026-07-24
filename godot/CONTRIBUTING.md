# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the handbook itself, not to game repos. It is not part of the app-generation contract.

How to change the Godot handbook itself without breaking the contract it exists to enforce.

## Default Approach

This handbook is a control plane, not prose. A change is correct only when the fast path, the routing, the slow path, the recipes, and the templates still agree with each other after it lands. Edit with that whole-system bar in mind, not one file in isolation.

### Know The Two-Speed Model You Are Editing

The handbook is structured as a two-speed documentation system. Find where your change belongs before you touch a file:

- **Fast path — [AGENTS.md](AGENTS.md).** Repo-wide invariants, change defaults, the Change Routing table (change-type to file-set mapping), and the verification bar. Terse and authoritative. Refine it; never weaken it. Edit the routing table when you add a new kind of change or a new file that a change type should also touch; most other changes are *not* fast-path changes.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Architecture, scene-tree and directory map, lifecycle, test taxonomy, and rationale. Edit this when the *why* behind a rule changes or needs fuller background than the fast path carries.

Topic depth lives in `foundations/`, `systems/`, `quality/`, and `operations/`; copy-paste scaffolding in `templates/`; step-by-step procedures in `recipes/`; gates in `checklists/`; and the binding decisions in `decisions/`.

### Use The House Templates

Match the existing shape exactly; do not invent a new one.

- **Topic docs** (`foundations/*`, `systems/*`, `quality/*`, `operations/*`): `# Title` -> one-line purpose -> `## Default Approach` (with `###` subsections) -> `## Common Mistakes And Forbidden Patterns` -> `## Verification And Proof` -> an optional related-links tail. To deepen a topic, insert `###` subsections in the right place and extend Common Mistakes and Verification; never reorder or duplicate existing content.
- **Recipes** (`recipes/*`): Files To Touch / Steps / Invariants To Preserve / Proof, with concrete copy-pasteable commands.
- **Checklists** (`checklists/*`): `# <Name> Checklist`, a one-line scope sentence, then `##` groups of traceable `- [ ]` boxes with no prose between items. Each box is tied to evidence a reviewer can point at and states the check, not the fix — the fix lives in the owning doc, linked when the answer is non-obvious. Every checklist ends with a `## Proof` group of concrete commands or observable behaviors, and lands with its [AGENTS.md](AGENTS.md) Change Routing row in the same PR.
- **Templates** (`templates/*`): fill-in skeletons with explicit `<placeholder>`s, not finished prose; filenames encode destinations per [templates/README.md](templates/README.md).
- **Index/navigational docs** (the glossary, README routing tables) may use their own clear shape.

Voice everywhere: terse, opinionated, contract-not-tutorial. State the standardized engineering decision; this is not a Godot engine tutorial. Cross-link only files that already exist (or are landing in the same change).

### Keep The Sync Surfaces In Sync

A new file that only exists is invisible; the handbook's entry points must route to it. When you add, rename, or remove a doc, update every routing surface in the same PR:

- **[README.md](README.md)** — the Reading Paths table and/or the Handbook Map must reach the file.
- **[AGENTS.md](AGENTS.md) Change Routing** — the change types the file governs get a row (or an updated row) pointing at it; every row must point at files that exist, cite anchors (`### Section`) that exist, and list the real also-update set.
- **The directory index** — [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), [decisions/README.md](decisions/README.md), and [templates/README.md](templates/README.md) each index their own directory; a file missing from its index is a defect.

Beyond routing, three sets of files are load-bearing and drift silently. When you change one, change its partners in the same PR:

- **[AGENTS.md](AGENTS.md) <-> [maintainer-reference.md](maintainer-reference.md) <-> recipes.** An invariant added to the fast path needs its rationale in the slow path and its procedure in a recipe; a new recipe needs a routing row and, if it introduces a rule, a fast-path line. None of the three may contradict the others.
- **Topic docs and the templates that scaffold them.** If you change a default that `templates/` embodies — a `.gdlintrc` rule, a `.gitignore` entry, a CI workflow stage, a project-settings convention — update the template and its governing doc together.
- **The verify gate.** The canonical stage list lives once in [AGENTS.md](AGENTS.md) (## Baseline Verification); every other doc links to it or names a single stage, never restates the full command list. Duplicated command sequences are how gates drift.

There is no `reference/` exemplar project yet — compiling reference projects are a planned later phase. Saying they are planned is fine; a link is a broken promise.

### Engine Version Updates

This handbook is pinned to the current stable Godot 4.x line, and all `docs.godotengine.org/en/stable/` citations track the newest stable release. Godot's release policy is that minor releases (4.x) add features but "minor compatibility breakage in very specific areas *may* happen", patch releases fix bugs without breaking compatibility, and a stable branch is supported "at least until the next stable branch is released and has received its first patch update" (`https://docs.godotengine.org/en/stable/about/release_policy.html`). That policy dictates the update procedure:

- **Exact pins live only in `templates/`.** The engine version in [templates/github-workflows-ci.yml](templates/github-workflows-ci.yml) and any toolchain pin (gdtoolkit, test-framework version) are the single place a version number is written. Prose says "Godot 4.x", or names a minor only when a rule depends on a feature that landed in it. A patch version in prose is a defect.
- **On a new 4.x minor release**: bump the template pins, re-check every `en/stable` citation still says what the handbook claims (the stable channel now serves the new minor), confirm the lint and test toolchain pins support it, and sweep prose for minor-version-flagged claims. Verify any new pin against its upstream source before committing it.
- **On a patch release**: bump the CI template pin; no prose sweep is required by policy.
- **Godot 3.x is the long-term-supported previous major** per the release policy; it is out of scope here. Do not add 3.x guidance or dual-version instructions — a repo stuck on 3.x is not governed by this handbook.

### Changing An Invariant Requires An ADR

The repo-wide invariants in [AGENTS.md](AGENTS.md) and the Non-Negotiables in [README.md](README.md) are not edited by opinion. To change one — to relax a default, swap the default test framework or lint toolchain, or alter the verification bar — record an Architecture Decision Record first per [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md). The ADR captures the alternatives and consequences; only then do the fast path, slow path, recipes, and templates move together to match it. Tool picks are single-homed: the test framework is owned by [quality/testing.md](quality/testing.md), the lint toolchain by [quality/linting-and-formatting.md](quality/linting-and-formatting.md), CI actions by [operations/ci-and-release.md](operations/ci-and-release.md) — link the owner rather than restating a pick in another doc.

## Common Mistakes And Forbidden Patterns

- Editing one file in isolation and leaving AGENTS / maintainer-reference / the recipe out of sync.
- Adding a new doc without touching the routing surfaces: README reading paths or map, the AGENTS Change Routing table, and the owning directory's README index.
- Weakening a fast-path invariant in passing instead of through an ADR.
- Inventing a new document shape instead of using the house template for that doc type.
- Cross-linking a file that does not exist yet, or linking `reference/` before that phase lands.
- Writing an engine or toolchain version into prose, or restating the verify-gate stages instead of linking the Baseline Verification table in [AGENTS.md](AGENTS.md).
- Citing a `docs.godotengine.org` URL without confirming it resolves on the `en/stable` channel — pages move between releases (the autoloads best-practices page's pre-4.x URL now 404s).
- Turning a contract doc into a Godot tutorial — explaining nodes, signals, or GDScript syntax instead of stating the standardized decision and citing the primary source.
- Changing a default that `templates/` embodies without updating the template in the same PR.
- Reordering or duplicating existing sections when deepening a topic doc instead of inserting `###` subsections in place.

## Verification And Proof

- All internal cross-links resolve to files that exist after the change, and every cited `### anchor` exists in its target file (grep for the heading).
- Every added or renamed doc is reachable from [README.md](README.md), routed in [AGENTS.md](AGENTS.md), and indexed by its directory README.
- AGENTS.md, maintainer-reference.md, and the affected recipe(s) tell one consistent story; no rule contradicts another.
- No version pin appears outside `templates/`; any changed pin was verified upstream; every external citation resolves on `en/stable`.
- Any change to a repo-wide invariant references an ADR in the PR.
- The PR itself uses [templates/pull_request_template.md](templates/pull_request_template.md): small, one logical change, green CI, one review.

Related: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) for the ADR shape, [templates/README.md](templates/README.md) for the template index and filename conventions, and [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership changes.
