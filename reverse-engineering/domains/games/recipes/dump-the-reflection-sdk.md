# Recipe: Dump The Reflection SDK

Use this to produce an SDK and `.usmap` for a build where runtime dumping is in scope. Governed by [../reflection-dumping.md](../reflection-dumping.md).

> **Warning:** This injects code into the game. Walk [../../../checklists/runtime-session-gate.md](../../../checklists/runtime-session-gate.md) first. If the build ships anti-cheat and the session would connect to live services, stop here and use the static alternatives in [../reflection-dumping.md](../reflection-dumping.md#static-alternatives).

## Files To Touch

- `briefs/<yyyy-mm-dd>-reflection-dump.md` (gate answers and action plan; committed before the session)
- `sdk/<build-id>/`, `keys/<build-id>.usmap` (git-ignored)
- `sdk/<build-id>/index.sqlite` (generated)
- `protocol/rpc-index.md` (generated; committed)
- `builds/<build-id>/names.json` (updated)
- `findings.md`

## Steps

1. Commit the brief with the gate checklist answers and this exact plan.
2. Take a btrfs snapshot or copy of `compatdata/<appid>/` ([../../../platforms/linux-proton/lab-setup-linux-proton.md](../../../platforms/linux-proton/lab-setup-linux-proton.md#snapshot-before-anything-runs)).
3. Choose the dumper per [../reflection-dumping.md](../reflection-dumping.md#dumper-choice). For UE 5.5+ with UE4SS, use the `experimental-latest` build.
4. UE4SS route (human-run):
   1. Copy the proxy `dwmapi.dll` and the `ue4ss/` folder into `<Project>/Binaries/Win64/`.
   2. In `UE4SS-settings.ini`, enable the C++ header dump and `.usmap` generation; disable mods you did not write.
   3. Set Steam launch options: `WINEDLLOVERRIDES="dwmapi.dll=n,b" %command%`.
   4. Launch. Stay offline or at the main menu without logging in. Trigger the dumps from the UE4SS console or keybinds per its docs.
   5. Quit the game.
5. Dumper-7 route (human-run): build the DLL in x64-Release, inject it into the running game in the lab prefix with an injector run through `protontricks-launch`, wait for output under `C:\Dumper-7` in the prefix, quit.
6. Clean up: remove the proxy DLL and `ue4ss/` from `Win64/`, clear the launch options, and verify `sha256sum -c builds/<build-id>/hashes.sha256` passes for `Win64/`.
7. Move outputs: headers to `sdk/<build-id>/`, `.usmap` to `keys/<build-id>.usmap`.
8. Run the indexing script: parse headers into `sdk/<build-id>/index.sqlite` (classes, functions with flags, properties with offsets, enums). Generate `protocol/rpc-index.md` from functions with network flags.
9. From the agent, query the index with REA `inspect_sqlite_database` (schema first, then selected tables and columns) and convert any recorded function offsets to `names.json` entries.
10. Add findings: dump tool and version, class and RPC counts, Evidence IDs for the index inspection.

## Invariants To Preserve

- No login, no live-server connection with an injected client.
- `Win64/` matches `hashes.sha256` after cleanup.
- Dumps and mappings never leave `sdk/` and `keys/`.

## Proof

- The brief's commit predates the files in `sdk/<build-id>/`.
- `sha256sum -c` for `Win64/` files passes after cleanup.
- `protocol/rpc-index.md` row count equals the index's network-flagged function count.
