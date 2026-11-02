# Patch Tracking

How to keep findings true across client patches: one fingerprint per build, a scripted re-extraction pipeline, REA-backed function and capture comparison, and a ledger that marks findings stale instead of silently trusting them.

## Default Approach

Online games patch often, and every patch can move offsets, rename reflected types, re-key containers, and change message layouts. A finding is pinned to the build it was observed on. On a new build, every finding is "unverified on this build" until a re-check says otherwise.

### One Build, One Identity

- The build ID is whatever the game exposes most stably: the Steam build ID from `steamapps/appmanifest_<appid>.acf` (`"buildid"`), the exe version resource, or the engine build string. Record all three in `builds/<build-id>/fingerprint.md`.
- `builds/<build-id>/hashes.sha256` lists every file analyzed. A finding cites the build ID, and the hash file proves the bytes.
- Preserve the old build's files outside the game directory before Steam updates them if you need to diff binaries. Steam replaces files in place. Copy the executable and container files you analyzed into a dated, git-ignored archive.

### Re-Run The Pipeline, Do Not Re-Do The Work

Each step from first contact becomes a script under `scripts/` so the next patch is a pipeline run:

1. Fingerprint — [../recipes/fingerprint-a-ue5-build.md](../recipes/fingerprint-a-ue5-build.md).
2. Key and mapping refresh — new AES key check, new `.usmap` — [../techniques/asset-extraction.md](../techniques/asset-extraction.md).
3. Asset export of tracked DataTables and string tables to JSON — [../recipes/extract-assets-and-datatables.md](../recipes/extract-assets-and-datatables.md).
4. Reflection dump — [../recipes/dump-the-reflection-sdk.md](../recipes/dump-the-reflection-sdk.md), only where the runtime gate allows it.
5. REA snapshot of the new executable — [../recipes/open-the-client-in-rea.md](../recipes/open-the-client-in-rea.md).

Run the asset-tool steps through REA `capture_process_scenario` (`rea capture-process`) so the command line, exit code and output are kept as Evidence alongside the results.

### Diff Data Before Code

Exported JSON diffs answer most "what changed" questions cheaply. `git diff --stat` over `exports/<build-id>/` DataTables (kept in a separate, private data repo if they are large) shows new items, rebalanced stats, and new systems before any binary is opened. A reflection-dump diff (`diff -r sdk/<old>/ sdk/<new>/` on the generated headers) shows added or removed UClasses, UFunctions, and RPCs; a new `Server*` UFunction is a new client-to-server message.

### Diff Code With REA

- `compare_functions` compares explicit function Evidence from two builds. Use it for the functions already in the ledger: serializers, RPC handlers, key derivation.
- `find_changed_behavior` combines existing comparisons. REA's evidence guide is explicit that "static differences remain candidates, not causal proof".
- Re-locate a moved function by its strongest anchor, in order: a unique string reference, a reflected UFunction name, a byte signature (`search_bytes`), then call-graph position.
- Record the re-located address and the method in the ledger row's Evidence column. A signature that broke is itself a finding about the patch.

### Diff Traffic

Capture the same scripted player action on both builds (log in, move, open inventory, log out) and compare with REA `inspect_pcap_capture` summaries and `inspect_packet_capture` field extracts. A changed endpoint, SNI, port, or packet-size distribution on an unchanged action points at a protocol change before any decompilation.

### Ledger Discipline

On a new build:

1. Add a `Verified on <build-id>` note to each finding re-checked, with the new Evidence ID.
2. Mark findings not yet re-checked as `open` with "stale since <build-id>".
3. Findings proved wrong move to Superseded; never edit the old claim in place.

## Common Mistakes And Forbidden Patterns

- Trusting an offset, signature, or message layout from the previous build without a re-check.
- Letting Steam overwrite the only copy of the old build's executable before the diff.
- Re-doing manual analysis that a script could re-run.
- Treating a static diff as proof of a behavior change.
- Committing exported game data to a public repository.

## Verification And Proof

- Each `builds/<build-id>/` directory has a fingerprint and hash file.
- `scripts/` contains a single entry point that runs the pipeline for a build ID and exits non-zero on any failed step.
- Every finding in `findings.md` either names the current build in its verification note or is marked stale.

## Where To Go Next

- [../recipes/diff-a-patch.md](../recipes/diff-a-patch.md) — the step-by-step patch-day procedure.
- [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md) — the ledger and Evidence rules this doc applies.
