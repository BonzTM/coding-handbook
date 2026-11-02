# Asset Extraction

How to read a UE5 game's cooked content: containers, encryption keys from your own client, mapping files for unversioned properties, and exporting DataTables, string tables, and config to diffable text.

## Default Approach

Asset extraction is static, agent-safe once keys and mappings are in hand, and often the highest-value work in an MMO engagement. DataTables hold item, skill, quest, and NPC data; string tables hold every UI and quest text; cooked config holds network and backend settings. Export them to JSON, commit nothing from the game, and diff them per build.

### Tool Roles

| Tool | Role | Source |
|---|---|---|
| FModel | GUI browser and exporter for paks and IoStore, built on CUE4Parse | `github.com/4sval/FModel` |
| CUE4Parse | .NET library behind FModel; scriptable for pipelines; defines `GAME_UE5_6` | `github.com/FabianFG/CUE4Parse`, `CUE4Parse/UE4/Versions/EGame.cs` |
| retoc | "CLI tool for packing/unpacking Unreal Engine IoStore containers (.utoc/.ucas)"; converts IoStore to legacy pak (`to-legacy`) and back (`to-zen`); "versions 5.3+ are well supported" | `github.com/trumank/retoc` |
| repak | "Library and CLI tool for working with Unreal Engine .pak files"; reads encrypted paks with a "256 bit AES encryption key" | `github.com/trumank/repak` |
| UAssetGUI / UAssetAPI | Per-asset inspection and editing; UAssetGUI release notes say "Added support for Unreal Engine 5.6 and 5.7" | `github.com/atenfyr/UAssetGUI` |
| AESDumpster | "helps locate and extract AES keys from Unreal Engine games"; lists "5.0 to 5.6 (Supported)"; runs on Windows and Linux | `github.com/GHFear/AESDumpster` |

Exact versions and checked dates live in [../decisions/tooling.md](../decisions/tooling.md). FModel and UAssetGUI are Windows desktop apps; run them in the tools prefix, never the game prefix ([../foundations/lab-setup-linux-proton.md](../foundations/lab-setup-linux-proton.md#two-prefixes-never-one)). CUE4Parse-based scripts, retoc, repak, and AESDumpster have CLI forms suitable for an agent pipeline.

### Encryption Keys

> **Warning:** Container encryption is a technical protection measure. Decrypting it is circumvention, and in the US it needs a legal basis such as the § 1201(f) interoperability purpose ([../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md#the-legal-frame-united-states-and-european-union)). Record that basis in `scope.md` before recovering a key.

- Test first: open the containers in FModel with no key. If payloads read, there is no encryption.
- An encrypted game embeds its key in the client so the client can read its own content. AESDumpster scans the executable statically for key candidates. Running it against your own installed copy reads a file you own and executes nothing from the game.
- Validate each candidate by decrypting a known container's index. A key that opens the index is the key for that container; multi-key games need several.
- Store keys under `keys/<build-id>.txt`, which git ignores and the agent's permissions deny reading ([../templates/re-workspace-agents.md](../templates/re-workspace-agents.md)). Pipelines read the file at run time. Keys are never posted, shared, or written into findings.
- Keys change between builds without notice. A decrypt failure on patch day is the first symptom.

If the key cannot be recovered statically, stop. Extracting it from a running anti-cheat-protected process is out of scope.

### Mapping Files For Unversioned Properties

Cooked UE5 packages usually need a `.usmap` ([../foundations/unreal-engine-internals.md](../foundations/unreal-engine-internals.md#unversioned-properties-and-mappings)). Without it, FModel lists assets but cannot show their properties.

- Generate it from the running game with UE4SS (`DumpUSMAP`) or Dumper-7 (its `MappingGenerator` writes `.usmap`). Both are runtime injection and go through [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md); they are out of scope on anti-cheat builds connected to live services.
- A mapping file is build-specific. Store it under `keys/<build-id>.usmap` with the key, and never reuse one across builds.
- Without a mapping, raw structure work is still possible: asset names, imports and exports, and string tables often parse; property values do not.

### What To Export For An MMO

Prioritize, in order:

1. **Config** — `DefaultEngine.ini`, `DefaultGame.ini`: NetDriver class, online subsystem, backend URLs, feature flags.
2. **DataTables and CurveTables** — items, abilities, stats, loot presentation, NPCs, quests. Export to JSON.
3. **String tables and localization** — `.locres` files; every player-facing string, keyed. Useful for naming opcodes and UI flows.
4. **Blueprint classes** — for logic analysis ([blueprint-analysis.md](blueprint-analysis.md)).
5. **Maps** — world layout, spawn points, and actor placements, as needed.

Meshes, textures, and audio are rarely needed for protocol or system work. They are also the content most clearly not yours to redistribute.

### Scripted Exports

Build the export as a script once, so patch day is a pipeline run ([../operations/patch-tracking.md](../operations/patch-tracking.md)):

- Use the CUE4Parse library from a small .NET console project, or FModel's batch export, to write selected asset paths to `exports/<build-id>/` as JSON.
- Run the script through REA `capture_process_scenario` (`rea capture-process`) so the command, exit status, and output are Evidence.
- Normalize exported JSON (sorted keys, stable formatting) so diffs show content changes only.

## Common Mistakes And Forbidden Patterns

- Installing FModel or UAssetGUI into the game's Proton prefix.
- Posting, committing, or pasting AES keys or `.usmap` files anywhere outside `keys/`.
- Reusing a mapping file or key from another build and trusting partial output.
- Modifying containers or adding `_P` patch paks to a live-service client.
- Exporting the whole game "to have it" instead of the tables a question needs.
- Extracting keys or mappings from an anti-cheat-protected live client.

## Verification And Proof

- `keys/` and `exports/` are ignored: `git check-ignore keys exports`.
- The export script exits non-zero on any missing key, mapping, or asset path.
- Each exported table used in a finding is named with its asset path and build ID, and the export run's Evidence ID is cited.

## Where To Go Next

- [../recipes/extract-assets-and-datatables.md](../recipes/extract-assets-and-datatables.md) — the step-by-step procedure.
- [reflection-dumping.md](reflection-dumping.md) — producing a mapping file where in scope.
- [blueprint-analysis.md](blueprint-analysis.md) — reading logic in cooked Blueprints.
