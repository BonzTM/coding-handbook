# Architecture Decision Records

How handbook-built Godot repos capture the non-obvious decisions a future owner must understand to safely change the project.

## Default Approach

An Architecture Decision Record (ADR) is a short, immutable document that states one decision, the context that forced it, and the consequences accepted. It answers "why is it this way?" so a new owner does not have to reverse-engineer intent from scenes, scripts, commit history, or absent people. ADRs are not design proposals, status reports, or living docs: once accepted, an ADR is frozen and only superseded, never edited.

ADRs live in the game or tool repo, not in this handbook. This doc owns the process and the record shape; [README.md](README.md) lists which handbook doc owns each default pick, so an ADR records only the project-specific deviation and links to the owner for the rest.

## When An ADR Is Required

Write an ADR before merging any decision that is non-obvious or hard to reverse. At minimum:

- An **engine version jump** — any minor upgrade (one 4.x minor to the next). The official release policy states minor versions may ship "minor compatibility breakage in very specific areas" (`docs.godotengine.org/en/stable/about/release_policy.html`), so record the version chosen, the compatibility areas checked, and confirmation that export templates and CI moved in lockstep per [../operations/ci-and-release.md](../operations/ci-and-release.md).
- **Addon adoption** — anything landing in `addons/`. Record the maintenance status observed, the 4.x line it targets, what it replaces, and the removal plan if it goes unmaintained.
- A **mixed-language boundary** — introducing C# alongside GDScript, or moving a system across that line. The lock-in is real: C# requires the .NET editor build, "projects written in C# cannot be exported to the web platform", and scripts cannot inherit across the language boundary (`docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html`). The tradeoff rules live in [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md); the ADR records which side of them this project landed on and why.
- A **multiplayer transport and authority model** — which `MultiplayerPeer` backs the project and where authority lives, per the defaults in [../systems/multiplayer.md](../systems/multiplayer.md).
- A **save-format change** — any deviation from the [../systems/save-and-load.md](../systems/save-and-load.md) default, and especially any move toward loading resource files from user-writable paths, which that doc forbids for security reasons.
- An **autoload addition** beyond the baseline set — global state is the centralized-failure-point risk [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) polices; the ADR records why the system's scope earns it and follows [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md).
- Any other **deviation from a handbook default** — layout, typing enforcement, test framework, lint toolchain, CI export path — per the ownership table in [README.md](README.md).

If the decision is obvious, trivially reversible, and local to one scene or script, skip the ADR — the code and its tests are the record. When in doubt, the test is: would a competent new owner be surprised by this, and is undoing it expensive? If yes, write the ADR.

## ADR Lifecycle

Each ADR carries exactly one status:

- **Proposed** — under review, not yet binding. Used only while the PR is open.
- **Accepted** — merged and binding. This is the steady state.
- **Superseded** — replaced by a later ADR. Set the status to `Superseded by 0007` and leave the body untouched.
- **Deprecated** — no longer applies and nothing replaced it (the capability was removed).

A decision changes by **adding a new ADR**, never by editing the old one. The superseding ADR links back in its `Supersedes` field (`Supersedes: 0002`), and the old ADR's status is flipped to `Superseded by NNNN` with a link forward. The chain of links is the audit trail: a reader can walk from current state back through every prior decision — every engine bump, every addon that came and went — and see exactly what changed and why.

## Storage And Indexing

ADRs live in the project repo under `decisions/`:

```text
decisions/0001-project-baseline.md
decisions/0002-adopt-gut-for-unit-tests.md
decisions/0003-csharp-for-procedural-generation.md
```

- One file per decision: `decisions/NNNN-kebab-title.md`.
- `NNNN` is a zero-padded, monotonically increasing integer. Never reuse or renumber.
- The title line and the filename agree: `# 0003. csharp for procedural generation`.
- Every record carries the same five sections, in order: **Status**, **Context**, **Decision**, **Consequences**, **Alternatives Considered**. There is no separate template file — this shape is the template.

The first record is the project baseline: the pinned engine version and renderer, the language mix, the initial addon set, and the test framework, each stated as "handbook default" or as a deviation with its reason. Recording the baseline makes every later deviation legible. [../checklists/new-project.md](../checklists/new-project.md) confirms it exists before first merge.

Link the live `decisions/` directory from the project README so it is visible at handoff. Code shows *what* the project does; ADRs are the only durable record of *why*, and a handoff is complete when `decisions/` answers — without a meeting — why each irreversible choice was made, what was rejected, and what would trigger re-evaluating it. See [../onboarding-and-handoff.md](../onboarding-and-handoff.md).

## Common Mistakes And Forbidden Patterns

- Editing an accepted ADR to reflect a new decision instead of writing a superseding record — this destroys the audit trail.
- Renumbering, reusing, or deleting ADR files; gaps and historical entries are expected and correct.
- Bumping the engine version inside an unrelated PR with no record — the compatibility check and template/CI sync vanish with the context.
- Vendoring an addon with no ADR, leaving the next owner unable to tell a deliberate dependency from a leftover experiment.
- Writing ADRs as design proposals or aspirational plans rather than decisions actually made.
- Recording only the chosen option with no Alternatives Considered — the rejected options are half the value.
- Restating a handbook default in an ADR instead of linking the owning doc and recording only the deviation.
- Deferring the record until "later"; an ADR written after the context is forgotten is fiction.
- Burying ADRs outside `decisions/` (wiki, tickets, chat) where they rot and are invisible at handoff.

## Verification And Proof

```bash
ls decisions/
```

A repo's ADR practice is in good shape when:

- every required decision above has a corresponding `decisions/NNNN-*.md` with all five sections;
- `decisions/0001-*` records the project baseline, including the pinned engine version;
- numbering is contiguous and unique, and no accepted ADR has been edited post-merge;
- every `Superseded` ADR links forward and its replacement links back;
- the project README links the `decisions/` directory.

ADRs are done when a new owner can read `decisions/` and answer every "why is it this way?" without asking a person.

## Where To Go Next

- [README.md](README.md) — which handbook doc owns each default pick an ADR may deviate from.
- [../AGENTS.md](../AGENTS.md) (## Change Routing) — route ordinary changes there; ADRs are only for decisions no row covers.
- [../onboarding-and-handoff.md](../onboarding-and-handoff.md) — where the ADR set is consumed.
