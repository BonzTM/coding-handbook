# Testing

Owns the proof bar for Godot repos: which framework proves behavior, where tests live, how scenes and signals get tested, and how the suite runs headless in CI.

## Default Approach

Test game logic below the scene tree first, scenes second, and the assembled game last. A repo picks exactly one test framework, wires it to run headless from the command line on day one, and treats the local headless run and the CI run as the same command. The framework defaults follow the [../README.md](../README.md) Default Stack table; deviations are recorded via [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

### Test Taxonomy

| Test Type | Use for | Default location |
|---|---|---|
| unit | pure logic in `static` functions, `RefCounted` classes, and custom `Resource` validation — no tree required | `test/unit/` |
| scene | one scene instantiated in isolation: public methods, state transitions, emitted signals | `test/unit/` |
| integration | multi-scene interaction, physics-tick behavior, save/load round-trips, autoload-backed systems | `test/integration/` |

The cheapest tests are the ones that never touch the tree. Scene architecture makes this possible: logic extracted into `static` functions on `class_name` scripts or into `Resource` classes (see [../foundations/resources-and-data.md](../foundations/resources-and-data.md)) tests without instantiating anything, which is one more reason [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) pushes logic out of deep node hierarchies.

### Framework Selection

| Situation | Framework | Why |
|---|---|---|
| GDScript-only repo (the handbook default) | GUT 9.x | GDScript-native, targets Godot 4.x, CLI runner for headless CI, doubles/stubs/spies, JUnit XML export (`github.com/bitwes/Gut`) |
| C# scripts present, or IDE test-adapter integration matters | gdUnit4 (+ gdUnit4Net for C#) | tests in GDScript and C#, scene runners, VS/Rider test adapters, parameterized tests, JUnit XML reports (`github.com/godot-gdunit-labs/gdUnit4`, `github.com/godot-gdunit-labs/gdUnit4Net`) |

- One framework per repo. Two frameworks means two runners, two CI stages, and two assertion dialects for the same proof; do not mix them.
- The selection rule is mechanical: adopting C# per [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md) is the trigger that justifies gdUnit4. Picking gdUnit4 "just in case" on a GDScript-only repo is not — record the reason in an ADR if you deviate.
- Both frameworks are addons and live under `addons/` like any third-party asset (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`). They are dev-time dependencies: exclude `addons/gut/` (or `addons/gdUnit4/`) and `test/` from every export preset's resource filters so test tooling never ships (see [../operations/ci-and-release.md](../operations/ci-and-release.md)).

### Test Layout And Naming

- Tests live under `test/` at the project root, split into `test/unit/` and `test/integration/`. This matches GUT's own CLI defaults and examples (`gut.readthedocs.io/en/latest/Command-Line.html`) and gdUnit4's CI examples, which run the `./test` directory.
- Test files are `test_<subject>.gd` — `test_` is GUT's default discovery prefix (`gut.readthedocs.io/en/latest/Command-Line.html`), and `snake_case` file naming is already the repo-wide rule owned by [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md).
- Mirror the source layout: the test for `characters/player/player.gd` is `test/unit/characters/player/test_player.gd`. A reviewer finds the proof for a script without searching.
- Test code follows the same style and static-typing rules as production code. Untyped test scripts hide exactly the argument-shape bugs the tests exist to catch.
- Commit the runner configuration — GUT reads `res://.gutconfig.json` by default (`gut.readthedocs.io/en/latest/Command-Line.html`) — so the editor panel, the local CLI run, and CI all discover the same directories with the same options.
- The step-by-step version of adding a test is [../recipes/add-a-unit-test.md](../recipes/add-a-unit-test.md).

### Scene And Signal Testing

Scene tests are where the handbook's architecture rules pay rent: a scene designed with no external dependencies (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`) can be instantiated alone in a test, handed fakes for whatever its parent would inject, and asserted against its public surface.

- Instantiate the packed scene under test, add it to the tree, exercise it, and free it in teardown. Use the framework's auto-free helpers so a failing assertion cannot leak nodes into the next test — orphaned nodes are cross-test state, and both frameworks exist to make freeing automatic rather than optional.
- Assert against the scene's contract: its public methods, its exported state, and the signals it emits. Reaching into a child scene's internal node paths (`get_node("Sprite2D/AnimationPlayer")` from a test) couples the test to layout that [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) explicitly allows to change.
- Signals are the observable output of "signal up" designs, so test them as outputs: both frameworks provide signal watching and emission assertions. Assert the signal fired with the expected arguments — that is the contract [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md) creates, and its naming and payload rules are owned by [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md).
- Never wait on wall-clock time. `await get_tree().create_timer(0.5).timeout` in a test is either a hidden race or wasted suite time. Await the signal you are actually waiting for, or drive frames deterministically through the framework's simulation helpers (gdUnit4's scene runner exists for exactly this).
- Behavior that lives in `_physics_process()` runs on the fixed physics tick, not the render frame (`docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`) — test it by simulating physics frames through the framework, and put it in `test/integration/`. The movement rules themselves are owned by [../systems/physics-and-movement.md](../systems/physics-and-movement.md).
- Drive input through InputMap actions, not synthesized raw keycodes, so tests exercise the same action boundary gameplay uses — the boundary [../foundations/input-handling.md](../foundations/input-handling.md) owns.
- Logic that depends on an autoload is a test smell before it is a test problem: the autoload best-practices page warns that global state means "one object is now responsible for all objects' data" (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`). Inject the dependency instead of mutating the singleton from the test; if the autoload survived its ADR, give it a resettable interface the test can control.
- Save/load proof belongs in integration tests: round-trip real files through the serialization path, including rejection of malformed input — the validation rules are owned by [../systems/save-and-load.md](../systems/save-and-load.md).

### Headless CI Runs

Godot 4 runs fully headless with `--headless`, which is required on runners without GPU access (`docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`). The test suite runs headless on every push, before the export job.

- GUT runs through its command-line script; it returns exit code 0 when all tests pass and 1 when any fail (`gut.readthedocs.io/en/latest/Command-Line.html`), so the CI step gates on the exit code with no output parsing:

```bash
godot --headless -d -s --path "$PWD" addons/gut/gut_cmdln.gd \
  -gdir=res://test/unit,res://test/integration -ginclude_subdirs -gexit
```

- gdUnit4 ships a runner script driven by a `GODOT_BIN` environment variable — `./runtest.sh -a ./test` — and emits JUnit XML reports for CI ingestion (`github.com/godot-gdunit-labs/gdUnit4`).
- CI must run the exact Godot version the repo pins — minor versions may break compatibility "in very specific areas" (`docs.godotengine.org/en/stable/about/release_policy.html`), so a version drift between CI and the team is a silent proof gap. The pin lives with [../foundations/project-setup.md](../foundations/project-setup.md).
- `.godot/` is gitignored per [../foundations/project-setup.md](../foundations/project-setup.md), so a fresh CI checkout has no import cache. Run `godot --headless --import` first — it "starts the editor, waits for any resources to be imported, and then quits" (`docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`) — then run the tests. The committed [../templates/github-workflows-ci.yml](../templates/github-workflows-ci.yml) owns this ordering; the pipeline stages and export gates are owned by [../operations/ci-and-release.md](../operations/ci-and-release.md).
- The local proof command and the CI command are the same command. If a developer cannot reproduce a CI test failure by running the headless line above locally, fix the pipeline, not the test.

## Common Mistakes And Forbidden Patterns

- Two test frameworks in one repo, or gdUnit4 adopted on a GDScript-only project without the C#/IDE trigger and without an ADR.
- Tests that only pass inside the editor — relying on a rendered window, editor plugins, or focus state instead of running under `--headless`.
- `await get_tree().create_timer(...)` or other wall-clock waits standing in for awaiting the actual signal or simulating frames.
- Asserting a child scene's internal node paths instead of its signals and public methods.
- Instantiated nodes never freed in teardown, leaking orphans that make later tests fail for unrelated reasons.
- Testing autoload-dependent logic by mutating the live singleton and hoping test order restores it.
- Physics-tick behavior asserted from render-frame tests, then blamed on flakiness.
- Shipping `test/` or the framework addon because export presets never excluded them.
- CI running a different Godot version than the repo pins, or skipping the `--import` step and shrugging at resource-load errors.
- Committing tests that were never run headless locally — green-in-editor is not the gate.

## Verification And Proof

```bash
# import cache first on a fresh checkout, then the suite — same commands locally and in CI
godot --headless --import
godot --headless -d -s --path "$PWD" addons/gut/gut_cmdln.gd \
  -gdir=res://test/unit,res://test/integration -ginclude_subdirs -gexit
# gdUnit4 repos instead:
#   GODOT_BIN=<path-to-godot> ./runtest.sh -a ./test
```

- the headless run exits 0; the CI job gates on that exit code, and gdUnit4 repos also upload the JUnit XML report
- every new script or scene lands with its mirrored `test_` file per [../recipes/add-a-unit-test.md](../recipes/add-a-unit-test.md)
- scene tests free what they instantiate; the framework's orphan/auto-free reporting stays clean
- signal contracts added via [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md) have an emission assertion proving them
- export presets exclude `test/` and the framework addon, verified by inspecting a CI-produced export per [../operations/ci-and-release.md](../operations/ci-and-release.md)

Testing is done when the headless suite proves the behavior the change touched — logic below the tree, the scene's public contract, and the signals it emits — not when the editor panel turns green.
