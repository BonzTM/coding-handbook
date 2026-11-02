# Recipes

Step-by-step, target-neutral procedures for common reverse-engineering work. Each recipe has the fixed shape **Files To Touch / Steps / Invariants To Preserve / Proof** and links the topic doc that owns its rules. Game-specific recipes (fingerprint, asset export, reflection dump, preservation server) live in the games section — see [../domains/games/README.md](../domains/games/README.md). Start routing at the Change Routing table in [../AGENTS.md](../AGENTS.md).

## First Contact

- [open-a-target-in-rea.md](open-a-target-in-rea.md) — build a Ghidra snapshot from a shell, open it through REA, and land the first anchors. Governed by [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md) and [../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md).

## Network

- [capture-and-classify-traffic.md](capture-and-classify-traffic.md) — record one scripted session of your own target and classify its endpoints. Governed by [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md).
- [write-a-wireshark-dissector.md](write-a-wireshark-dissector.md) — a Lua dissector that Wireshark, TShark, and REA all use. Governed by [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md#dissectors-rea-can-use).

## Maintenance

- [diff-a-patch.md](diff-a-patch.md) — patch-day re-fingerprint, re-export, re-locate, and ledger update. Governed by [../operations/patch-tracking.md](../operations/patch-tracking.md).

## Game Recipes

Indexed in the games section: [../domains/games/README.md](../domains/games/README.md#recipes) — fingerprint a UE5 build, extract assets and DataTables, dump the reflection SDK, reconstruct a preservation server.

## Where To Go Next

- Gates: [../checklists/README.md](../checklists/README.md)
- Copy-in starting points: [../templates/README.md](../templates/README.md)
