# AGENTS.md - Coding Handbook Router

This root file is routing-only. It is not a second source of truth for any language handbook.

## Fast Path

1. Identify the language or domain: Go tasks read `golang/AGENTS.md`; C#/.NET tasks read `csharp/AGENTS.md`; engine-agnostic game design tasks read `game-design/AGENTS.md`; Godot tasks read `godot/AGENTS.md`. If the task names no language or domain, ask or infer from the target repo before proceeding.
2. Use the Change Routing table in that handbook's `AGENTS.md` to route the change.
3. Read the relevant topical docs under that handbook before editing or answering.

## Current Handbooks

- `golang/` — full handbook: `golang/README.md` for human onboarding, `golang/AGENTS.md` for the fast-path contract, `golang/maintainer-reference.md` for slower-path architectural rationale.
- `csharp/` — full handbook: `csharp/README.md` for human onboarding, `csharp/AGENTS.md` for the fast-path contract, `csharp/maintainer-reference.md` for slower-path architectural rationale. Covers Windows, Linux, and macOS development; the verification gate is `pwsh ./verify.ps1` on all three.
- `game-design/` — full handbook: `game-design/README.md` for human onboarding, `game-design/AGENTS.md` for the fast-path contract, `game-design/maintainer-reference.md` for slower-path architectural rationale. Engine-agnostic game design: pillars, core loops, one-page system docs, and playtesting as the proof discipline.
- `godot/` — full handbook: `godot/README.md` for human onboarding, `godot/AGENTS.md` for the fast-path contract, `godot/maintainer-reference.md` for slower-path architectural rationale. Covers Godot 4.x with typed GDScript; the verification gate is `gdformat --check`, `gdlint`, headless tests, and a `godot --headless --export-release` export smoke.
