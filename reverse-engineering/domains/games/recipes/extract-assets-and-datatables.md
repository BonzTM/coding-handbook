# Recipe: Extract Assets And DataTables

Use this to export a build's config, DataTables, and string tables to diffable JSON. Governed by [../asset-extraction.md](../asset-extraction.md).

> **Warning:** Decrypting containers is circumvention of a technical protection measure. Confirm `scope.md` records the legal basis before step 3.

## Files To Touch

- `keys/<build-id>.txt`, `keys/<build-id>.usmap` (git-ignored)
- `exports/<build-id>/` (git-ignored)
- `scripts/export_assets.*` (new or updated; committed)
- `evidence/<build-id>-assets-<yyyy-mm-dd>.json` (new)
- `findings.md`

## Steps

1. Read the engine version and container list from `builds/<build-id>/fingerprint.md`.
2. In the tools prefix, open `<Project>/Content/Paks` in FModel with the matching engine version and no key. If payloads read, skip to step 5.
3. Run AESDumpster against the `-Shipping.exe` on disk. Collect key candidates.
4. Test each candidate in FModel or with `repak` against a `.pak` index. Save the working key(s) to `keys/<build-id>.txt`. Never paste a key into chat, a finding, or an agent prompt.
5. If assets list but properties do not parse, a `.usmap` is needed. Produce it only per [dump-the-reflection-sdk.md](dump-the-reflection-sdk.md) and its gate. Without one, export what parses (string tables, config, names) and record the gap as an unknown.
6. Export, in order: cooked `Config/` ini files, the DataTables and CurveTables the question needs, `.locres` string tables. Write JSON to `exports/<build-id>/<asset path>.json`.
7. Move the working export into `scripts/export_assets.*`: inputs are build ID, game path, key file, mapping file, and an asset-path list; output is normalized JSON (sorted keys). Use CUE4Parse from a small .NET console app, or retoc `to-legacy` plus repak for container work.
8. Run the script through REA so the run is Evidence:

   ```json
   {"tool": "capture_process_scenario",
    "executable": "/abs/path/scripts/export_assets.sh",
    "arguments": ["<build-id>"],
    "working_directory": "/abs/path/<workspace>",
    "timeout_ms": 1800000, "idle_timeout_ms": 600000}
   ```

9. Add findings for the tables used (asset path, row counts, schema notes) citing the run's Evidence ID. Export the bundle.

## Invariants To Preserve

- FModel and UAssetGUI run in the tools prefix only.
- Keys and mappings stay in `keys/`; exports stay in `exports/`; both are git-ignored.
- No container in the game directory is modified or added.

## Proof

- `git check-ignore keys exports` prints both paths.
- The export script re-run produces byte-identical JSON for an unchanged build.
- `findings.md` rows cite the `capture_process_scenario` Evidence ID for the export run.
