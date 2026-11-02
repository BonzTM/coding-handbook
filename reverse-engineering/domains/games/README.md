# Games Domain

The game-specific layer of the reverse-engineering handbook. The general method — the agentic workflow, scope and legal frame, static and dynamic analysis, protocol work — lives at the handbook root ([../../README.md](../../README.md)); this section adds what is specific to shipped game clients, with Unreal Engine 5 online games as the worked target.

Read the general foundations first, then the docs here for the engine, the client/server model, and the game-specific techniques.

## Topic Docs

- [unreal-engine-internals.md](unreal-engine-internals.md) — the reflection system, global object/name tables, `ProcessEvent`, cooked content and containers, unversioned properties, Blueprint bytecode; the analyst's map of a UE5 binary.
- [mmo-client-server-model.md](mmo-client-server-model.md) — classifying a target's architecture (Unreal replication vs. custom backend) and what the client can and cannot reveal about the server.
- [anti-cheat-and-scope.md](anti-cheat-and-scope.md) — the game-specific scope rules that layer on the general [scope doc](../../foundations/scope-legal-and-ethics.md): anti-cheat and game terms.

## Techniques (Game-Specific)

- [asset-extraction.md](asset-extraction.md) — containers, encryption keys, mapping files, and exporting DataTables/config to diffable JSON. Applies the general [recon](../../techniques/recon-and-fingerprinting.md) method to UE5 content.
- [reflection-dumping.md](reflection-dumping.md) — producing an SDK and `.usmap` from the running game, where the runtime gate allows it, and the static alternatives.
- [blueprint-analysis.md](blueprint-analysis.md) — reading game logic stored as Kismet bytecode in cooked assets.

## Operations

- [game-preservation.md](game-preservation.md) — preserving a discontinued-server title: server re-implementation, archival extraction, the anti-cheat gates, and the act-versus-tool gap.

## Recipes

- [recipes/fingerprint-a-ue5-build.md](recipes/fingerprint-a-ue5-build.md) — identify engine version, containers, protections, and SDKs for a build.
- [recipes/extract-assets-and-datatables.md](recipes/extract-assets-and-datatables.md) — keys, mappings, and scripted JSON export.
- [recipes/dump-the-reflection-sdk.md](recipes/dump-the-reflection-sdk.md) — the gated runtime reflection dump.
- [recipes/reconstruct-a-preservation-server.md](recipes/reconstruct-a-preservation-server.md) — stand up an independent server from the documented protocol.

## Where To Go Next

- [../../AGENTS.md](../../AGENTS.md) — the change-routing table, including the games rows.
- [../../techniques/static-binary-analysis.md](../../techniques/static-binary-analysis.md) and [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md) — the general methods these docs specialize.
