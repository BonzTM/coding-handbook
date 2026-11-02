# Recipes

Step-by-step procedures for the common work in a UE5 client reverse-engineering engagement. Each recipe has the fixed shape **Files To Touch / Steps / Invariants To Preserve / Proof** and links the topic doc that owns its rules. Start routing at the Change Routing table in [../AGENTS.md](../AGENTS.md).

## First Contact

- [fingerprint-a-ue5-build.md](fingerprint-a-ue5-build.md) — identify engine version, containers, protections, and SDKs for a build, statically. Governed by [../techniques/recon-and-fingerprinting.md](../techniques/recon-and-fingerprinting.md).
- [open-the-client-in-rea.md](open-the-client-in-rea.md) — build a Ghidra snapshot from a shell, open it through REA, and land the first anchors. Governed by [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md) and [../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md).

## Content

- [extract-assets-and-datatables.md](extract-assets-and-datatables.md) — keys, mappings, and scripted JSON export of config, DataTables, and string tables. Governed by [../techniques/asset-extraction.md](../techniques/asset-extraction.md).
- [dump-the-reflection-sdk.md](dump-the-reflection-sdk.md) — gated runtime reflection dump with UE4SS or Dumper-7, then indexing for the agent. Governed by [../techniques/reflection-dumping.md](../techniques/reflection-dumping.md).

## Network

- [capture-and-classify-traffic.md](capture-and-classify-traffic.md) — record one scripted session of your own client and classify its endpoints. Governed by [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md).
- [write-a-wireshark-dissector.md](write-a-wireshark-dissector.md) — Lua dissector that Wireshark, TShark, and REA all use. Governed by [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md#dissectors-rea-can-use).

## Maintenance

- [diff-a-patch.md](diff-a-patch.md) — patch-day re-fingerprint, re-export, re-locate, and ledger update. Governed by [../operations/patch-tracking.md](../operations/patch-tracking.md).

## Where To Go Next

- Gates: [../checklists/README.md](../checklists/README.md)
- Fill-in artifacts: [../templates/README.md](../templates/README.md)
