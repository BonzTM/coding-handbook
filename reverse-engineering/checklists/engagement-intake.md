# Engagement Intake Checklist

Gate before any analysis starts on a new target or engagement. Run it once per engagement, then re-run the Target group after every major build. The rules live in [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md) and [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md). The Target and Tooling boxes below show the game/Proton engagement; a non-game target fills the equivalent fingerprint and tooling for its own format and platform.

## Purpose And Authorization

- [ ] Is the purpose one of the allowed purposes in [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md#allowed-purposes), written as a concrete goal in `scope.md`?
- [ ] Are the ToS/EULA clauses on reverse engineering, modding, and third-party software quoted verbatim in `scope.md` with a retrieval date?
- [ ] Is a bug bounty or VDP checked for and recorded (URL or "none")?
- [ ] Is every item in "Never In Scope" left unchecked and unmodified?

## Workspace

- [ ] Does a dedicated git repository exist with the layout from [../templates/re-workspace-layout.md](../templates/re-workspace-layout.md)?
- [ ] Does `.gitignore` exclude keys, SDK dumps, snapshots, exports, captures, and binaries?
- [ ] Does `AGENTS.md` exist from [../templates/re-workspace-agents.md](../templates/re-workspace-agents.md), with the target fields filled?
- [ ] Does `.claude/settings.json` (or the equivalent for the agent client) deny or ask for every runtime tool?

## Target

- [ ] Is `builds/<build-id>/fingerprint.md` filled per [../techniques/recon-and-fingerprinting.md](../techniques/recon-and-fingerprinting.md) — format, architecture, version, hashes, protections, dependencies (for a game, per [../domains/games/recipes/fingerprint-a-ue5-build.md](../domains/games/recipes/fingerprint-a-ue5-build.md): engine version, container list, anti-cheat, packer)?
- [ ] Is `builds/<build-id>/hashes.sha256` committed?
- [ ] Is the architecture classified as a first hypothesis, marked as inference (for a game client/server, per [../domains/games/mmo-client-server-model.md](../domains/games/mmo-client-server-model.md))?

## Tooling

- [ ] Does `rea doctor --provider ghidra --json` (or `binary_session`) show the Ghidra check healthy?
- [ ] Are `tshark` and `capinfos` on PATH, and `REA_TSHARK_COMMAND` set if field-level capture inspection is planned?
- [ ] Is the lab ready for the target's platform (for a Windows target under Proton, a separate Wine tools prefix per [../platforms/linux-proton/lab-setup-linux-proton.md](../platforms/linux-proton/lab-setup-linux-proton.md))?

## Proof

- [ ] `git log` shows `scope.md`, `AGENTS.md`, and the build fingerprint committed before the first Evidence bundle under `evidence/`.
- [ ] A fresh agent session, given only the workspace path, states the target, the purpose, and the forbidden actions correctly.
