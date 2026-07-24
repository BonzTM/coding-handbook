# Recipes

Step-by-step implementation guides for the common Godot changes this handbook governs. Each recipe is a fixed-shape contract — **Files To Touch / Steps / Invariants To Preserve / Proof** — and links the topical doc that owns the rules it applies. Use a recipe when you know what kind of change you are making and want the exact file set and proof steps without rediscovering them.

For routing a change to its recipe and related obligations, start at the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## Scene And Architecture Recipes

- [add-a-scene.md](add-a-scene.md) - add a self-contained scene with its script, `snake_case` file placement, and injected dependencies instead of outward references. Governed by [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md).
- [add-an-autoload.md](add-an-autoload.md) - register a wide-scope singleton only after the per-scene, `static` func, and shared `Resource` alternatives are ruled out. Governed by [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) and [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md).
- [add-a-signal-contract.md](add-a-signal-contract.md) - add a past-tense signal with typed arguments, wired call-down/signal-up, without turning it into a command channel. Governed by [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md).

## Input And UI Recipes

- [add-an-input-action.md](add-an-input-action.md) - add an InputMap action with device bindings and gameplay handling in `_unhandled_input`, never raw keycodes in game logic. Governed by [../foundations/input-handling.md](../foundations/input-handling.md).
- [add-a-ui-screen.md](add-a-ui-screen.md) - add a Control screen styled from the project theme, with local overrides limited to one-off layout tweaks. Governed by [../systems/ui-and-theming.md](../systems/ui-and-theming.md).

## Data And Persistence Recipes

- [add-a-save-field.md](add-a-save-field.md) - add a save-file field with versioned migration and load-time validation, without `ResourceLoader.load()` on user-writable files. Governed by [../systems/save-and-load.md](../systems/save-and-load.md) and [../foundations/resources-and-data.md](../foundations/resources-and-data.md).

## Quality And Pipeline Recipes

- [add-a-unit-test.md](add-a-unit-test.md) - add a headless-runnable unit test in the project's chosen framework, exercised through a scene-free seam where possible. Governed by [../quality/testing.md](../quality/testing.md).
- [set-up-ci-export.md](set-up-ci-export.md) - stand up a CI pipeline that lints, tests, and exports via `--headless` with pinned engine version and export templates. Governed by [../operations/ci-and-release.md](../operations/ci-and-release.md) and [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).

## Where To Go Next

- Routing a change to the right files: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- Handbook overview: [../README.md](../README.md)
- Checklists for lifecycle gates: [../checklists/README.md](../checklists/README.md)
- Copy-in starting points for config files: [../templates/README.md](../templates/README.md)
