# Recipe: Open The Client In REA

Use this for the first native-analysis session on a new client build: open the UE5 executable with REA's Ghidra provider, keep a snapshot so later sessions skip auto-analysis, and land the first anchored findings. Governed by [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md) and [../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md).

## Files To Touch

- `builds/<build-id>/fingerprint.md` (read; must exist from [fingerprint-a-ue5-build.md](fingerprint-a-ue5-build.md))
- `snapshots/<build-id>/` (new, git-ignored)
- `builds/<build-id>/names.json` (new, committed): confirmed address names
- `evidence/<build-id>-anchors-<yyyy-mm-dd>.json` (new)
- `findings.md` (new rows)

## Steps

1. Confirm the provider. Run `rea doctor --provider ghidra --json`, or call `binary_session` with `{}` from the agent. The `ghidra` entry under `analysis_provider_candidates` must report `available`. If it does not, fix the provider per [../foundations/lab-setup-linux-proton.md](../foundations/lab-setup-linux-proton.md#rea-and-ghidra-provider); do not re-run agent setup.
2. Identify the real game module. UE5 games ship a small launcher `.exe` at the install root and the real binary under `<Project>/Binaries/Win64/<Project>-Win64-Shipping.exe`. Open the `-Shipping.exe`, not the launcher. Use the path recorded in the fingerprint.
3. Run the first auto-analysis from a shell, not through the agent, so no MCP client tool-call timeout applies. Set the heap and deadline per [../foundations/lab-setup-linux-proton.md](../foundations/lab-setup-linux-proton.md#hardware-and-time-budget), then build the snapshot with the CLI. The CLI `--snapshot` flag is documented as "Load and update a local analysis snapshot" (`rea decompile --help`):

   ```bash
   export GHIDRA_INSTALL_DIR=~/.local/opt/<ghidra-release>
   export GHIDRA_HEADLESS_MAXMEM=24G REA_GHIDRA_STARTUP_TIMEOUT_MS=21600000
   mkdir -p snapshots/<build-id>
   rea decompile "/abs/path/<Project>-Win64-Shipping.exe" entry \
     --provider ghidra --snapshot "$PWD/snapshots/<build-id>/client.snapshot" --token-limit 2000
   ```

   Run it in a background terminal; it can take hours. `entry` is a cheap first question; any procedure name or address works.
4. From the agent, open with an explicit provider and the existing snapshot. `open_binary` imports a snapshot that "must match the binary identity, concrete provider, and canonical analysis profile exactly" (REA `open_binary` schema):

   ```json
   {"tool": "open_binary",
    "path": "/abs/path/<Project>-Win64-Shipping.exe",
    "provider_id": "ghidra",
    "snapshot_path": "/abs/path/<workspace>/snapshots/<build-id>/client.snapshot"}
   ```
5. Land anchors with literal searches, one at a time:
   - `search_strings` for the engine branch string recorded in the fingerprint (`++UE5+Release-5.` on stock branches; studio branches use their own, for example `++Project+<Name>-Release`).
   - `search_strings` for `/Script/Engine` and `/Script/CoreUObject`, the reflected package names; follow with `xrefs` to reach registration code.
   - `search_strings` for `NetDriver`, `IpNetDriver`, `GameNetDriver` to locate networking setup.
6. For each anchor, call `analyze_function` on the referencing function. Record a ledger row per confirmed anchor: string, address, function, Evidence ID.
7. Record each confirmed name in `builds/<build-id>/names.json` (`{"0x<address>": "<Name>"}`), then apply the file with `set_addresses_names`. Names are session-scoped metadata edits. REA's `close_binary` contract says "metadata edits prevent immutable snapshot saves until the session is recreated", so the names file, not the snapshot, is the durable copy. Re-apply it at the start of every session.
8. Call `export_evidence_bundle` to `evidence/<build-id>-anchors-<date>.json`, then `close_binary` without a snapshot path.

## Invariants To Preserve

- The opened file's SHA-256 in the Evidence matches `builds/<build-id>/hashes.sha256`.
- The snapshot is never shared across builds; a new build gets a new snapshot directory.
- Every ledger row cites an Evidence ID from the exported bundle.
- No runtime tool was called; this recipe is static only.

## Proof

- `rea inspect-evidence-bundle` on the exported file (with `detail: "summary"`) lists the anchor operations and the target SHA-256.
- Re-opening with the same `snapshot_path` returns results without a new auto-analysis run.
- Re-applying `names.json` with `set_addresses_names` returns success for every address.
- `findings.md` has at least the engine-version anchor row and one networking anchor row, both typed `observation`.
