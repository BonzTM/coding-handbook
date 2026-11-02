# Recipe: Diff A Patch

Use this on patch day: a new client build replaced the one your findings describe. Governed by [../operations/patch-tracking.md](../operations/patch-tracking.md).

## Files To Touch

- `builds/<new-build-id>/fingerprint.md`, `builds/<new-build-id>/hashes.sha256` (new)
- `snapshots/<new-build-id>/`, `exports/<new-build-id>/`, `sdk/<new-build-id>/` (new, git-ignored)
- `evidence/<new-build-id>-diff-<yyyy-mm-dd>.json` (new)
- `findings.md` (verification notes, stale marks, superseded rows)

## Steps

1. Before Steam updates, archive the analyzed files of the old build (executable and containers) to a dated, git-ignored directory. If the update already ran, use the copy from the last archive.
2. Read the new build ID: `grep '"buildid"' ~/.steam/steam/steamapps/appmanifest_<appid>.acf`.
3. Run [fingerprint-a-ue5-build.md](fingerprint-a-ue5-build.md) for the new build. Stop and re-run [../checklists/engagement-intake.md](../checklists/engagement-intake.md) Target group if the engine minor version, anti-cheat, or packer changed.
4. Re-run the extraction pipeline script for the new build ID. Expect the AES key and `.usmap` to need a refresh; a decrypt failure on known containers means the key changed.
5. Diff data: `git diff --no-index --stat exports/<old>/ exports/<new>/`, then the reflection headers: `diff -r --brief sdk/<old>/ sdk/<new>/`. List new and removed `Server*`/`Client*`/`Multicast*` UFunctions.
6. Open the new executable in REA with a new snapshot per [open-the-client-in-rea.md](open-the-client-in-rea.md).
7. For each ledger finding with a function address, re-locate the function by its anchor (string xref, reflected name, `search_bytes` signature), run `analyze_function`, and compare with the old build's Evidence using `compare_functions`.
8. Capture the scripted player action on the new build (human-run, per [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md)) and compare summaries with `inspect_pcap_capture`.
9. Update `findings.md`: verified notes with new Evidence IDs, stale marks for the rest, Superseded rows for anything proved wrong.
10. Export the Evidence bundle to `evidence/<new-build-id>-diff-<date>.json`.

## Invariants To Preserve

- No finding is marked verified on the new build without a new-build Evidence ID.
- Old-build Evidence bundles are kept, not overwritten.
- Static diffs are recorded as candidates until runtime or capture evidence confirms behavior.

## Proof

- `findings.md` has zero rows that cite only old-build evidence without a stale mark.
- The diff bundle contains `compare_functions` Evidence for every function-level finding that was re-checked.
- `git status` shows only workspace text files staged; no binaries, exports, or captures.
