# Onboarding And Handoff

> **Team-process document.** This governs project ownership transfer between people. It is not part of the app-generation contract; agents building or changing code do not read it.

This guide is for taking over a Godot project that was built from this handbook. It defines what a new owner reads on day one, the questions they must be able to answer before they truly own the repo, and what the outgoing owner is responsible for.

This is not the handbook's own "Start Here" in [README.md](README.md). That section is about working *inside the handbook*. This guide is about owning a *project that was built with it*. If you are an agent or human contributing a single change, you want [AGENTS.md](AGENTS.md) (including its Change Routing table), not this file.

## Who This Is For

- A new primary owner inheriting a game, tool, or shared Godot library project.
- The outgoing owner running the transfer.
- A reviewer confirming a handoff is actually complete before sign-off.

If the repo follows the handbook, every artifact this guide references already exists in the project. If one is missing, that is a handoff defect, not an optional extra; surface it before accepting ownership.

## Day-One Reading Path, In Order

Read these in the project repo, not in the handbook. Do not skip ahead; each step assumes the previous one.

| Step | Read | What you must come away knowing |
|---|---|---|
| 1 | Project `README.md` | What the project is, the exact pinned Godot 4.x version (standard or .NET build), and how to open and run it locally. |
| 2 | Project `AGENTS.md` | The repo's invariants, the task loop, and the exact baseline proof commands contributors must pass. |
| 3 | `decisions/` ADRs | Why the load-bearing choices were made (scene architecture, autoload set, save format, GDScript vs C#), what was rejected, and which decisions are still open. Read newest and any `Proposed`/`Accepted`-but-unimplemented records first. |
| 4 | Project `AGENTS.md` Change Routing table (or its equivalent) | How a given change routes to files, sync surfaces, and proof steps. |
| 5 | Run the verify gate locally | Confirm format, lint, headless tests, and a debug export all pass from a clean checkout before you change anything. |
| 6 | Open the project in the pinned editor version | Confirm the import cache regenerates cleanly (`.godot/` is gitignored and rebuilt on first open; see `docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`), the main scene runs, and no dependencies are missing. |

Step 5 is the gate between reading and owning. If the verify gate — the same commands CI runs, defined in [operations/ci-and-release.md](operations/ci-and-release.md) and [templates/github-workflows-ci.yml](templates/github-workflows-ci.yml) — does not pass from a clean clone on your machine, you do not yet have a working environment and the handoff is not done. See [quality/linting-and-formatting.md](quality/linting-and-formatting.md) and [quality/testing.md](quality/testing.md) for what each stage enforces.

## Questions A New Owner Must Be Able To Answer

You do not own the repo until you can answer every one of these unaided. Treat any "I'd have to ask the previous owner" as an open handoff item.

### Build, Test, Export

- Which exact Godot version is pinned, where is it recorded, and where do I download it? (README plus [foundations/project-setup.md](foundations/project-setup.md).)
- What is the full proof gate, and can I run it headless? (See [operations/ci-and-release.md](operations/ci-and-release.md) and [quality/testing.md](quality/testing.md).)
- Which export presets exist, which platforms do they target, and are the matching export templates installed for the pinned version? (`export_presets.cfg`; `docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`.)
- Can I produce a release export for every shipping platform without the previous owner present?

### Engine Version And Dependencies

- What is the engine upgrade policy, and who decides when to take a minor version? Minor 4.x releases may break compatibility "in very specific areas" (`docs.godotengine.org/en/stable/about/release_policy.html`), so upgrades are deliberate, ADR-recorded events, not routine bumps.
- What lives in `addons/`, where did each addon come from, and what license covers it and every binary asset? (See [foundations/project-setup.md](foundations/project-setup.md).)
- Which binary asset types go through Git LFS, and is LFS configured on my clone? ([templates/gitignore.txt](templates/gitignore.txt) and the project `.gitattributes`.)

### Architecture And Runtime State

- Which autoloads exist, what does each one own, and why does each earn global scope? (`project.godot` plus [recipes/add-an-autoload.md](recipes/add-an-autoload.md).)
- What is the main scene, and how does the tree get from boot to gameplay? (See [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md).)
- What is the save format, where do files live, and what is the compatibility story for saves from shipped builds? User-writable saves must never be loaded via `ResourceLoader`; see [systems/save-and-load.md](systems/save-and-load.md) for the rule and its rationale.
- Which input actions exist in the InputMap, and which are shipped as rebindable? (See [foundations/input-handling.md](foundations/input-handling.md).)
- Where does the GDScript/C# boundary sit, if the project uses both, and what is the stance for new code? (See [foundations/gdscript-vs-csharp.md](foundations/gdscript-vs-csharp.md).)

### Release And Distribution

- Where do builds ship (stores, web hosts, internal channels), what triggers a release, and how is one tagged?
- Where do signing and store credentials live (keystores, store keys, upload tokens), who grants access, and how does each rotate?
- How do I ship a hotfix build for a defect found in production, and how long does that path take end to end?

### Decisions And Direction

- Why are the load-bearing architectural choices the way they are? (`decisions/` ADRs; see [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).)
- Which decisions are still open or proposed, and what is blocking them?
- Which localization locales ship, and where do translation sources live? (See [systems/localization.md](systems/localization.md).)

If a question has no documented answer, the fix is to write it down in the project, not to keep it in your head. Undocumented knowledge is the failure this guide exists to prevent.

## Outgoing Owner Responsibilities

The outgoing owner runs the handoff and is responsible for it being complete. Walk the items below with the incoming owner; do not delegate them to the newcomer to discover.

- Update `CODEOWNERS` (or the platform equivalent) so reviews and notifications route to the new owner.
- Document where every credential lives — signing keystores, store upload keys, CI secrets — then grant the new owner access and revoke your own where it is no longer warranted. Treat signing-key custody as the highest-stakes item in the transfer: confirm the new owner holds the keys and their recovery path before you lose the ability to re-issue them.
- Transfer ownership of store dashboards, web hosting, and any crash-reporting or analytics accounts the project uses.
- Confirm the new owner can run the verify gate and produce a release export for every shipping platform independently, on their own machine, from a clean clone.
- Surface every open or proposed ADR and every undocumented decision; convert tribal knowledge into an ADR before you leave. (See [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).)
- Confirm the pinned engine version, export template versions, and addon inventory recorded in the repo match what CI and your machine actually use.
- Hand off the engine-upgrade watch: who tracks new 4.x releases and security advisories, and who decides when the project moves.
- Verify the asset and addon license inventory is complete enough that the new owner can answer "are we allowed to ship this?" for every file in the repo.

The transfer is complete only when the new owner can run the verify gate and a full release export independently and can answer the day-one questions without you.

## Where To Go Next

- The repo contract and routing hub: [AGENTS.md](AGENTS.md)
- Why decisions are recorded and how: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Release, export, and CI mechanics: [operations/ci-and-release.md](operations/ci-and-release.md)
- Engine pinning and repo hygiene: [foundations/project-setup.md](foundations/project-setup.md)
- Save-file compatibility and safety: [systems/save-and-load.md](systems/save-and-load.md)
- The deep-dive companion for maintainers: [maintainer-reference.md](maintainer-reference.md)
