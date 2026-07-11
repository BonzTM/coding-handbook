# AGENTS.md - Coding Handbook Router

This root file is routing-only. It is not a second source of truth for any language handbook.

## Fast Path

1. Identify the language: Go tasks read `golang/AGENTS.md`; C#/.NET tasks read `csharp/AGENTS.md`. If the task names no language, ask or infer from the target repo before proceeding.
2. Use the Change Routing table in that handbook's `AGENTS.md` to route the change.
3. Read the relevant topical docs under that handbook before editing or answering.

## Current Handbooks

- `golang/` — full handbook: `golang/README.md` for human onboarding, `golang/AGENTS.md` for the fast-path contract, `golang/maintainer-reference.md` for slower-path architectural rationale.
- `csharp/` — full handbook: `csharp/README.md` for human onboarding, `csharp/AGENTS.md` for the fast-path contract, `csharp/maintainer-reference.md` for slower-path architectural rationale. Covers Windows, Linux, and macOS development; the verification gate is `pwsh ./verify.ps1` on all three.
