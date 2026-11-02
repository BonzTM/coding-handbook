# Static Binary Analysis

How to analyze a UE5 shipping executable with an agent through REA and Ghidra: separate engine code from game code, port engine symbols from a self-built twin, find reflection and networking anchors, locate message writers and readers, and keep every result as Evidence.

## Default Approach

A UE5 shipping executable is mostly engine. The game-specific code is a minority, and the engine half can be named almost completely from the matching engine source. Name the engine first, then the game code becomes the unnamed remainder, and reflection names most of that. Do all of it statically through REA; every result returns as Evidence the ledger can cite.

### Session Discipline

- One REA session per build, opened from the snapshot ([../recipes/open-the-client-in-rea.md](../recipes/open-the-client-in-rea.md)).
- At session start: apply `builds/<build-id>/names.json` with `set_addresses_names`, then read `findings.md` and `list_unknowns`.
- One hypothesis per batch of calls. "The function that references `ServerMove` builds the movement RPC" is a hypothesis; "explore networking" is not.
- At session end: `export_evidence_bundle`, update `names.json` and `findings.md`, `close_binary`.

### Port Symbols From A Self-Built Twin

The highest-leverage static technique for UE5:

1. Check out the engine source tag matching the fingerprint ([../foundations/unreal-engine-internals.md](../foundations/unreal-engine-internals.md#get-the-matching-engine-source)).
2. Build a blank project as Win64 Shipping with the same engine version and, where known, the same compiler toolset. Keep the PDB.
3. Analyze the twin in Ghidra with its PDB.
4. Correlate twin and target with Ghidra Version Tracking, which ports "comments and labels into the new context". BinDiff can also "port symbols and comments between disassemblies" and "ships with support for IDA Pro, Binary Ninja and Ghidra" (`github.com/google/bindiff`).
5. Accept only high-confidence matches; export accepted names to `names.json` with their match score, so the agent applies them with REA and the ledger can cite the match method.

Expect strong matches on engine core, containers, networking, and serialization; weak matches where the studio modified the engine or the compiler inlined differently. Unmatched engine functions are still findable by their source strings and call relationships.

### Reflection And Core Anchors

Locate these first; everything else is found relative to them:

| Anchor | How to find it statically |
|---|---|
| `GUObjectArray` | Engine-source call sites of `FUObjectArray` methods matched from the twin; UE4SS's compatibility guide documents the access patterns it scans for. |
| `FName` to-string and constructor | Twin match; heavy reference counts from string-handling code. |
| `UObject::ProcessEvent` | Twin match; its vtable slot is consistent across all `UObject` vtables. |
| `StaticFindObject`, `StaticLoadObject` | Twin match; error-format strings from the engine source. |
| UHT registration tables | `search_strings` for a reflected class or function name, then `xrefs` into `.rdata` tables. |

Record each as a finding with its address, method (twin match, string xref, signature), and Evidence ID. Generate a byte signature for each with surrounding bytes and verify uniqueness with REA `search_bytes`; signatures make the next patch faster ([../operations/patch-tracking.md](../operations/patch-tracking.md)).

### Find The Networking Layer

- **NetDriver path** — `search_strings` for `IpNetDriver`, `GameNetDriver`, `NetDriverDefinitions`, and the NetDriver class named in cooked `DefaultEngine.ini`. Twin-ported names cover `UNetDriver`, `UNetConnection`, `UActorChannel`, and the bunch read/write functions.
- **Iris** — `search_strings` for `Iris` and `UseIrisReplication`. Iris is opt-in per project and labeled "Experimental" on Epic's Iris introduction page. Its presence changes which serialization functions matter.
- **Packet handlers** — strings for `PacketHandler`, `AESGCMHandlerComponent`, `EncryptionComponent`, and `EncryptionToken`. Their presence means UDP payloads may be encrypted on the wire.
- **RPC entry points** — for each Server RPC in `protocol/rpc-index.md`, the native UFunction thunk calls into the game's implementation and validation functions. Name them `<Class>_<Function>_exec`, `_Implementation`, `_Validate` as you confirm them.
- **Custom transports** — `search_strings` for library markers: `google.protobuf`, `.proto` file names, `flatbuffers`, `msgpack`, WebSocket upgrade headers, `grpc`. Then follow socket API imports (`ws2_32` `send`, `recv`, `WSASend`, `WSARecv`) up the call graph with `procedure_callers` until reaching the game's message layer.

### Writers And Readers

For each message in the ledger, find both directions:

- The **writer** (client→server) serializes fields in order. `analyze_function` on it gives the field order, widths, and constants. Cross-check against captured bytes with the protocol spec template.
- The **reader** (server→client) shows what the client trusts and how it validates. Note every field the reader ignores; the client cannot tell you what it means.
- Use REA `trace_native_values` for bounded def-use questions ("where does this field's value come from") and `build_call_path` to connect a UI action to a writer. Both report incomplete frontiers; an incomplete path is an unknown, not a negative.

### Working With The Agent

- Brief subagents per subsystem (inventory, movement, chat, combat) with the template in [../templates/agent-brief.md](../templates/agent-brief.md). Each returns findings rows and unknowns.
- Require `analyze_function` Evidence for every function-level claim; pseudocode alone is not enough for the ledger because the dossier carries callers, callees, and references.
- When pseudocode is unclear, request `procedure_assembly` for the same function rather than guessing types.
- Prefer `annotate_native_function` comments for intermediate notes that should not become names yet.

## Common Mistakes And Forbidden Patterns

- Reversing engine functions by hand that a twin build would name.
- Accepting low-confidence version-tracking matches into `names.json`.
- Claiming a writer's field order without a captured message that agrees.
- Treating an incomplete `build_call_path` frontier as proof that no path exists.
- Analyzing anti-cheat, integrity-check, or license-check code paths. Mark them as out of scope in the ledger and move on.
- Unpacking or devirtualizing protected code.

## Verification And Proof

- `names.json` entries each record their source (twin match with score, string xref, reflection dump, manual) and apply cleanly with `set_addresses_names`.
- Every core anchor in the table above is a ledger row with a signature verified unique by `search_bytes`.
- Every message spec cites both a writer or reader Evidence ID and a capture Evidence ID, or lists the missing half as an unknown.

## Where To Go Next

- [network-protocol-analysis.md](network-protocol-analysis.md) — confirming static findings against traffic.
- [blueprint-analysis.md](blueprint-analysis.md) — logic that lives in assets, not the executable.
- [../recipes/open-the-client-in-rea.md](../recipes/open-the-client-in-rea.md) — session setup.
