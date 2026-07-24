# Recipe: Add A Unit Test

Use this when a behavior change needs a headless-runnable test — pure logic, a scene's contract, or a signal emission. Runner selection (GUT for GDScript-only repos, gdUnit4 when C# tests or the scene runner matter) is owned by [../quality/testing.md](../quality/testing.md); this recipe wires a test into whichever runner the repo already uses. Do not mix runners in one repo.

## Files To Touch

- a new test script under the repo's `test/` tree, mirroring the source path with a `test_` prefix (e.g. `test/unit/characters/goblin/test_goblin.gd` for `characters/goblin/goblin.gd` — a convention this handbook adopts; the runner only requires the prefix)
- the script under test, only if decision logic must first be extracted out of `_process`/`_input` orchestration to be callable headless
- the CI workflow, only if this is the repo's first test — see [set-up-ci-export.md](set-up-ci-export.md) and [../templates/github-workflows-ci.yml](../templates/github-workflows-ci.yml)

## Steps

1. Create the test script extending the framework base class — `extends GutTest` (GUT) or `extends GdUnitTestSuite` (gdUnit4) — with `snake_case` file and function names and a `test_` prefix on both, so discovery stays convention-based.
2. Test decision logic as plain function calls first: call the method, assert the return value or resulting state. No scene tree needed for logic that took [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md)'s advice and stayed out of the node lifecycle.
3. For node behavior, instantiate headless and let the framework free it: GUT `add_child_autofree(Goblin.new())` / `autofree(...)`, gdUnit4 `auto_free(...)`. Inject fakes through the same seams the owner scene uses per [add-a-scene.md](add-a-scene.md) — method calls, handed-over references, `Callable` properties — never by stubbing autoloads.
4. When the behavior needs frames, simulated input, or signal timing, use gdUnit4's scene runner: `var runner := scene_runner("res://characters/goblin/goblin.tscn")`, drive it with `simulate_frames(n)` and its keyboard/mouse/action simulation, and await signals; the runner "is managed by the GdUnit API and is automatically freed after use" (`godot-gdunit-labs.github.io/gdUnit4/latest/advanced_testing/sceneRunner/`). In GUT repos, cover the same need with `add_child_autofree()` plus its awaiters (`gut.readthedocs.io`).
5. Assert on outcomes: returned values, node state after the call, and emitted signals per the contract in [add-a-signal-contract.md](add-a-signal-contract.md) — not on private fields or child-node internals.
6. Run headless locally with the exact command CI runs:
   - GUT: `godot --headless -d -s addons/gut/gut_cmdln.gd -gdir=res://test -ginclude_subdirs -gexit` — exits 0 when all tests pass, 1 on any failure (`gut.readthedocs.io/en/latest/Command-Line.html`)
   - gdUnit4: `GODOT_BIN=<godot binary> ./addons/gdUnit4/runtest.sh -a ./test` — exit 0 is success, 101 is warnings, anything else is failure (`godot-gdunit-labs.github.io/gdUnit4/latest/faq/ci/`)
7. Fix any orphan-node report from the runner before finishing; an orphan in a test is a leak the game will have too.

## Invariants To Preserve

- tests pass under `--headless` with no display, no editor, and no manual step; a test that needs the editor open is a broken test
- the local test command and the CI test command are the same command ([../AGENTS.md](../AGENTS.md) (## Baseline Verification))
- one behavior per test function; failure output must name what broke without reading the test body
- dependencies reach the code under test by injection, not by the test pre-loading global state
- everything a test instantiates is freed (`autofree`/`auto_free`/scene runner); the orphan count stays zero
- test file and function naming keeps the `test_` prefix so no runner configuration is needed to find them

## Proof

- the new test fails before the behavior change and passes after it
- the headless runner exits 0 locally using the exact command from step 6
- the CI test stage is green on the branch ([../operations/ci-and-release.md](../operations/ci-and-release.md))
- `gdformat --check` and `gdlint` pass on the test script per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md)
