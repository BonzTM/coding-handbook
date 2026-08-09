# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the handbook itself, not to application repos. It is not part of the app-generation contract.

How to change the Python handbook itself without breaking the contract it exists to enforce.

## Default Approach

This handbook is a control plane, not prose. A change is correct only when the fast path, routing, slow path, recipes, templates, and reference exemplar still agree after it lands. Edit with that whole-system bar in mind.

### Know The Two-Speed Model You Are Editing

- **Fast path — [AGENTS.md](AGENTS.md).** Repo-wide invariants, defaults, Change Routing, and verification bar. Terse and authoritative. Edit routing when a new change type or sync surface appears; most changes are not fast-path changes.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Architecture, package map, lifecycle, test taxonomy, troubleshooting, and rationale. Edit this when the why behind a rule changes or needs fuller background.

Topic depth lives in `foundations/`, `services/`, `quality/`, and `operations/`; scaffolding in `templates/`; procedures in `recipes/`; evidence gates in `checklists/`; binding choices in `decisions/`; and the verified end-to-end composition in `reference/exampleservice/`.

### Use The House Templates

Match the existing shape exactly; do not invent a new one.

- **Topic docs** (`foundations/*`, `services/*`, `quality/*`, `operations/*`): `# Title` -> one-line purpose -> `## Default Approach` with `###` subsections -> `## Common Mistakes And Forbidden Patterns` -> `## Verification And Proof` -> optional related-links tail. Deepen by inserting subsections; never reorder or duplicate the contract sections.
- **Recipes** (`recipes/*`): Files To Touch / Steps / Invariants To Preserve / Proof, with concrete copy-pasteable commands.
- **Checklists** (`checklists/*`): traceable `- [ ]` boxes tied to evidence, plus a closing Verification command block.
- **Templates** (`templates/*`): fill-in skeletons with explicit `<placeholder>`s. Filenames encode destinations per [templates/README.md](templates/README.md); Python sources use `.py.txt`, and dotfile templates omit the leading dot.
- **Index and navigational docs** may use their own clear shape.

Voice everywhere: terse, opinionated, contract-not-tutorial. State the standardized engineering decision. Cross-link only contract-listed files landing by the completed module.

### Keep The Routing Surfaces In Sync

When adding, renaming, or removing a doc, update every routing surface in the same change:

- [README.md](README.md) — Reading Paths and/or Handbook Map reaches it.
- [AGENTS.md](AGENTS.md) Change Routing — governed change types point to it, with real sync surfaces and proof.
- The directory index — `recipes/README.md`, `checklists/README.md`, `decisions/README.md`, and `templates/README.md` index their directory.

Three sets drift silently and move together:

- **AGENTS <-> maintainer-reference <-> recipes.** An invariant needs rationale and a procedure; a new recipe needs routing and, when it introduces a rule, a fast-path statement.
- **Topic docs <-> templates <-> reference exemplar.** A changed default must update the scaffolding and the executable proof that embodies it.
- **Verify-gate wording.** Every doc says `make verify` runs lock-check, frozen sync, format-check, lint, imports, types, test, audit. Do not paraphrase it into drift.

### Follow The Two-Phase Delivery Model

- **Phase 1:** docs and fill-in templates establish the contract. Cross-links may point to files committed in the same completed module plan, but no prose may claim the exemplar is green before Phase 2 proves it.
- **Phase 2:** `reference/exampleservice/` composes the contract into a complete FastAPI sidecar and must pass `make verify`. Fix any doc/template contradiction discovered while making the exemplar green.

### Templates Carry The Pins; Prose Does Not

Exact interpreter patches, package versions, action revisions, and image tags live only in `templates/` and `reference/exampleservice/`. Prose states the compatibility floor or policy and links to upstream docs; it does not pin transient versions. Verify every changed pin upstream.

### Link And Citation Hygiene

- Use relative links for handbook files, resolved from the current file.
- Link only files in the completed module contract; during staged work, distinguish planned artifacts from verified artifacts.
- Cite a section anchor only after confirming that heading exists.
- Use current official sources for Python, package, framework, tool, and security behavior. Never invent a PEP number, option, rule family, CLI flag, config key, or URL.
- A vendor or framework pick belongs in [decisions/framework-selection.md](decisions/framework-selection.md), not copied into unrelated topic docs.

### Changing An Invariant Requires An ADR

Repo-wide invariants in [AGENTS.md](AGENTS.md), Non-Negotiables in [README.md](README.md), stack defaults, boundaries, and the verification gate do not change by opinion. Record an ADR first per [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md), using [templates/adr-template.md](templates/adr-template.md); then update fast path, slow path, topics, recipes, templates, and exemplar together.

## Common Mistakes And Forbidden Patterns

- Editing one file in isolation and leaving routing, rationale, recipe, template, or exemplar contradictory.
- Adding a doc without reaching it from README, AGENTS Change Routing, and its directory index.
- Weakening an invariant without an ADR.
- Inventing a document shape instead of using the house template.
- Claiming a Phase 2 exemplar is verified before it exists and passes `make verify`.
- Linking a file outside the module contract or citing a heading that does not exist.
- Fabricating or relying on remembered flags, rule codes, config keys, PEPs, versions, or URLs.
- Writing transient pins in prose, or restating the verify stages inconsistently.
- Turning a contract into a Python tutorial.
- Changing a default embodied by templates or reference code without updating both.

## Verification And Proof

- Every internal link resolves against the completed module file contract; every cited anchor exists.
- Every added or renamed doc is reachable from [README.md](README.md), routed in [AGENTS.md](AGENTS.md), and indexed by its directory README.
- AGENTS, maintainer-reference, topic docs, recipes, templates, and exemplar tell one story.
- Official citations load and support the stated rule; all command flags, config keys, and Ruff families were checked upstream.
- No transient pin appears in prose; changed pins were verified upstream.
- Phase 1 templates are structurally complete; Phase 2 runs `make verify` in [reference/exampleservice/](reference/exampleservice/).
- Changes to invariants reference an ADR.
- The change follows [foundations/git-workflow.md](foundations/git-workflow.md): one logical change, green proof, reviewable diff.

Related: [foundations/git-workflow.md](foundations/git-workflow.md), [templates/project-contributing.md](templates/project-contributing.md), and [onboarding-and-handoff.md](onboarding-and-handoff.md).
