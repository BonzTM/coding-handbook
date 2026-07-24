# GDScript Vs C#

Owns the language decision for Godot repositories: GDScript is the default, and this doc defines when C# earns its place, what it costs at the export and interop boundaries, and the rules that govern a mixed-language codebase. All version-specific guidance targets the current Godot 4.x stable line.

## Default Approach

GDScript is the default scripting language for every repository under this handbook. It runs on every export platform Godot supports, has the tightest editor integration, and iterates with no compile step. Static typing is mandatory — the typed-opcode performance rationale and the `UNTYPED_DECLARATION` enforcement mechanism are owned by [gdscript-style-and-typing.md](gdscript-style-and-typing.md).

C# is an escalation, not a peer default. Adopting it changes the required editor build (the standard build has no C# support; C# requires the separate .NET edition of the editor — `docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html`), the CI toolchain, and the exportable platform set. That blast radius makes it an ADR-level decision recorded per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

### Approval Questions

Before introducing C#, answer all of these in writing in the ADR:

1. What concrete workload does typed GDScript fail to serve? Name the profiled system and the numbers (see [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md)) — "C# is faster" without a profile is not an answer.
2. Is web export on the shipped or plausible platform list? If yes, C# is disqualified for that project (see Export Platform Constraints below).
3. Which specific .NET libraries or existing team C# competence justify the second toolchain?
4. Where is the language boundary, which subsystem does C# own, and who maintains the interop surface?

## Decision Drivers

| Driver | Favors GDScript | Favors C# |
|---|---|---|
| Iteration speed | no build step; hot-reload script editing in the editor | build step on every change |
| Export platforms | all supported platforms, including web | no web export; Android and iOS experimental |
| Compute-heavy scripting | typed GDScript uses optimized opcodes but stays interpreted | generally faster for sustained compute-heavy code (`docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html`) |
| Ecosystem reuse | Godot asset library, GDScript-native tooling | the .NET package ecosystem and static analyzers |
| Editor integration | first-class: docs, autocomplete, debugger in-editor | external IDE (Rider/VS/VS Code) is the working environment |
| Toolchain surface | one engine binary | .NET edition editor plus .NET 8+ SDK everywhere, including CI (`godotengine.org/article/godotsharp-packages-net8/`) |

For workloads that outgrow scripting entirely (native SIMD, existing C/C++ libraries), the escalation path is GDExtension, not C# — that is a separate ADR with its own build-pipeline cost, out of scope here.

## Export Platform Constraints

- The official docs state plainly: "projects written in C# cannot be exported to the web platform" (`docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html`). This is the single biggest disqualifier — a jam build, an itch.io demo, or a playable-in-browser marketing page all die with C# in the project.
- C# on Android and iOS is experimental per the same page. Shipping mobile on C# means accepting experimental-tier platform support in the ADR, explicitly.
- Godot's C# packages require .NET 8 as the minimum since Godot 4.4 (`godotengine.org/article/godotsharp-packages-net8/`); the `Godot.NET.Sdk` NuGet package is versioned in lockstep with the engine, so engine upgrades and SDK upgrades move together.
- CI must run the .NET edition of the editor for import, test, and export steps; export presets and headless export mechanics are owned by [../operations/ci-and-release.md](../operations/ci-and-release.md) and [../recipes/set-up-ci-export.md](../recipes/set-up-ci-export.md).

## Interop Boundaries

Cross-language calls work, but the two directions are not symmetric (`docs.godotengine.org/en/stable/tutorials/scripting/cross_language_scripting.html`):

- **GDScript calling C#**: arguments and returns marshal automatically. This is the cheap, safe direction.
- **C# calling GDScript**: goes through `GodotObject.Call()`, `Get()`, and `Set()` — stringly-typed, unchecked at compile time, and a `Call` with missing required arguments "will fail silently and won't error out" per the docs. Every such call site is a latent silent failure.
- **Signals** cross the boundary cleanly: C# exposes them via `[Signal]` delegates, and `Callable.From()` bridges C# methods into Godot's signal system. Prefer signals over `Call()` for C#-to-GDScript communication — the wiring rules live in [signals-and-decoupling.md](signals-and-decoupling.md).
- **Inheritance does not cross the boundary in either direction**: "A GDScript file may not inherit from a C# script. Likewise, a C# script may not inherit from a GDScript file." Design the boundary with composition and signals; a class hierarchy that spans languages is impossible, not merely discouraged.
- Cross-language calls are dynamically dispatched. Keep per-frame interop chatter out of hot paths; batch data across the boundary instead of making many small calls.

## Mixed-Language Rules

The default is one language per repository. A mixed codebase is allowed only under an ADR, and it follows these rules:

- **Partition by subsystem, not by preference.** C# owns a whole named system (simulation core, procedural generation, a heavy data pipeline); GDScript owns scene glue, UI, and gameplay scripting. No file-by-file language choice.
- **The boundary is a named, documented API surface**: a small set of C# classes and `[Signal]` declarations the ADR lists explicitly. GDScript talks to C# through that surface only; C# reaches back into GDScript via signals, never via scattered `Call()` strings.
- **C# never depends on GDScript types.** The dependency arrow points one way: GDScript (orchestration) calls into C# (engine-room). This keeps the C# subsystem testable headless and removable if the ADR is later reversed.
- **Naming follows the platform split**: C# script files are PascalCase matching the class name; everything else is `snake_case` per the project-organization rules owned by [project-setup.md](project-setup.md) (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`).
- **Tests cover both sides of the boundary.** gdUnit4 tests GDScript and C#; gdUnit4Net provides the C# test-adapter integration. Framework selection and test placement are owned by [../quality/testing.md](../quality/testing.md); lint and format tooling per language by [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).

## Common Mistakes And Forbidden Patterns

- No adopting C# by Unity reflex. The Approval Questions demand a profiled workload or a concrete library need, in writing.
- No C# anywhere in a project whose platform list includes web — the export is impossible, and discovering that at ship time is the expensive way.
- No stringly-typed `Call()`/`Get()`/`Set()` outside the documented boundary surface. A misspelled method name or missing argument fails silently; every ad hoc call site is an undebuggable landmine.
- No cross-language calls inside `_process`/`_physics_process` hot loops without a profiler-backed justification; batch instead.
- No attempting inheritance across the boundary, and no boundary design that assumes it — composition and signals only.
- No untyped GDScript justified as "C# would have been faster anyway." Enable typed enforcement per [gdscript-style-and-typing.md](gdscript-style-and-typing.md) before any language escalation argument.
- No mixed-language repo without an ADR naming the C# subsystem, its API surface, and its owner.

## Verification And Proof

- The ADR exists and answers all four Approval Questions; the boundary API surface is listed in it.
- Every shipped platform exports headless in CI using the .NET edition build — a green `--headless --export-release` run per preset is the proof that the platform matrix in the ADR is real (see [../operations/ci-and-release.md](../operations/ci-and-release.md)).
- Boundary tests exercise every documented cross-language signal and call contract, including a negative test proving a bad `Call()` is caught by the project's own checks rather than failing silently (see [../quality/testing.md](../quality/testing.md)).
- `grep` for `GodotObject.Call`/`.Call(` in C# sources returns hits only inside the boundary module named by the ADR.
- Lint and format gates pass for both languages per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).

## Related

- [gdscript-style-and-typing.md](gdscript-style-and-typing.md) — typed GDScript rules and enforcement
- [signals-and-decoupling.md](signals-and-decoupling.md) — signal contracts, including cross-language signals
- [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) — the ADR mechanism this doc's exceptions route through
- [../operations/ci-and-release.md](../operations/ci-and-release.md) — headless export and the .NET CI toolchain
