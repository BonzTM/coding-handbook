# Blueprint Analysis

How to read game logic that a UE5 game implements in Blueprints: find the cooked Blueprint assets, read their Kismet bytecode, and connect Blueprint calls to native functions and RPCs.

## Default Approach

Studios often implement UI flows, quest logic, ability scripting, and interaction rules in Blueprints. That logic cooks into Kismet bytecode inside assets, not into the executable. Static asset analysis therefore covers a large part of "how does the client decide X" without touching the binary or the running game.

### Find The Blueprints

- Blueprint classes cook as assets whose generated class name ends in `_C` (for example `BP_PlayerCharacter_C`). Search the container listing for `BP_`, `WBP_` (widgets), `ABP_` (animation), and `GA_`/`GE_` prefixes where the project uses the Gameplay Ability System.
- A reflection dump lists Blueprint-generated classes alongside native ones, with their functions. Cross-reference the dump to decide which assets matter.
- String tables point at UI flows; a quest string's key often shares a prefix with the widget or quest Blueprint that shows it.

### Read The Bytecode

- **UAssetGUI / UAssetAPI** — UAssetAPI supports "reading and writing raw Kismet (blueprint) bytecode"; UAssetGUI shows `ScriptBytecode` as a Kismet tree node (`UAssetGUI/TableHandler.cs`). Use it to read one function at a time.
- **kismet-analyzer** — "Tools for analyzing and manipulating kismet bytecode in cooked Unreal Engine assets" (`github.com/trumank/kismet-analyzer`). Its only published build is a development pre-release, and its supported engine versions are not documented; verify on one known asset before trusting output.
- **FModel** — shows Blueprint exports and properties; use it to locate assets, then a bytecode tool to read them.

Kismet bytecode references other functions by import. A call to a native UFunction appears as a call to its reflected name; a call to an RPC is a call to the RPC's UFunction. That is how Blueprint logic connects to the network layer.

### What To Record

For each Blueprint function relevant to a question:

- The asset path and build ID.
- The function name and its reflected flags (exported from the asset).
- Calls to native functions and RPCs, in order, with the arguments' sources.
- Branches on replicated properties or server responses.

Write these into the finding ledger like any static observation. Bytecode export files stay in `exports/<build-id>/`, which git ignores.

### Agent Role

Bytecode reading is static and agent-safe. Give the agent exported JSON or text of one Blueprint at a time, with the RPC index, and ask for the call sequence and decision points. Do not give it the whole export tree.

## Common Mistakes And Forbidden Patterns

- Searching only the executable for logic that lives in Blueprints.
- Trusting a bytecode tool's output on an engine version it has not been verified against.
- Modifying and re-cooking Blueprints for a live-service client.
- Treating a client Blueprint's checks as the server's rules; the server may re-check differently.

## Verification And Proof

- Each Blueprint-derived finding cites the asset path, build ID, and the export run's Evidence ID.
- The bytecode tool was validated on one asset whose behavior is independently known (for example a UI flow observed in play).
- Every RPC call found in a Blueprint exists in `protocol/rpc-index.md`.

## Where To Go Next

- [asset-extraction.md](asset-extraction.md) — getting the assets out.
- [reflection-dumping.md](reflection-dumping.md) — the RPC index Blueprint calls are matched against.
- [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md) — what those RPCs look like on the wire.
