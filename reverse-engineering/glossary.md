# Glossary

> **Lookup aid, not required reading.** Consult a single entry when a term is unclear.

Canonical vocabulary for this handbook. Each entry has one meaning and one owning doc. Terms are alphabetical.

## Anchor
A location in the client binary confirmed by a strong, re-findable clue — a unique string reference, a reflected name, a twin-build match, or a verified byte signature — used to find other code relative to it. Owned by [techniques/static-binary-analysis.md](techniques/static-binary-analysis.md).

## Build ID
The identifier a finding is pinned to: the Steam `buildid` from `appmanifest_<appid>.acf`, backed by a committed hash file of the analyzed files. Owned by [operations/patch-tracking.md](operations/patch-tracking.md).

## Evidence ID
The `evidence_id` REA returns with every result, naming a retained record with the target's hash, coverage, and limitations. Findings cite Evidence IDs. Owned by [foundations/agentic-re-workflow.md](foundations/agentic-re-workflow.md).

## Finding ledger
The engagement's `findings.md`: one row per conclusion, typed observation or inference, with Evidence IDs and confidence, plus the Unknowns table. The only record of conclusions. Owned by [foundations/agentic-re-workflow.md](foundations/agentic-re-workflow.md); template [templates/finding-ledger.md](templates/finding-ledger.md).

## Game prefix / tools prefix
The game prefix is the Proton Wine prefix at `compatdata/<appid>/pfx` that the game runs in; the tools prefix is a separate Wine prefix for Windows-only RE tools. Tools never go in the game prefix. Owned by [foundations/lab-setup-linux-proton.md](foundations/lab-setup-linux-proton.md).

## IoStore
UE5's container format for the Zen loader: `.utoc` table-of-contents files plus `.ucas` data files. Owned by [foundations/unreal-engine-internals.md](foundations/unreal-engine-internals.md).

## Kismet bytecode
The compiled form of Blueprint graphs, stored in cooked assets and run by the engine's script VM. Owned by [techniques/blueprint-analysis.md](techniques/blueprint-analysis.md).

## Act-versus-tool gap
The §1201 structure where a triennial exemption can permit a person's own act of circumvention for a covered purpose, but no exemption reaches the anti-trafficking prohibitions (17 U.S.C. § 1201(a)(2), (b)), so building, sharing, or publishing the tool or method stays unlawful. Owned by [operations/game-preservation.md](operations/game-preservation.md).

## Local-gameplay limit
The clause in the 2024 US preservation rulemaking (37 CFR 201.40) that scopes the game-preservation exemption to play "not through an online service or facility", which excludes restoring an MMO's online multiplayer. Owned by [operations/game-preservation.md](operations/game-preservation.md).

## Mapping file (`.usmap`)
A community-format file describing every reflected type's property layout for one build, needed to parse cooked packages that use unversioned properties. Build-specific and never redistributed. Owned by [techniques/asset-extraction.md](techniques/asset-extraction.md).

## ProcessEvent
`UObject::ProcessEvent`, the engine function that executes a UFunction with a parameter block; the single best runtime observation point for reflected calls. Owned by [foundations/unreal-engine-internals.md](foundations/unreal-engine-internals.md).

## Recorded-runtime / live-runtime
Evidence types: recorded-runtime comes from a saved artifact of a past session (a capture); live-runtime comes from an approved instrumentation session (a breakpoint log). Owned by [maintainer-reference.md](maintainer-reference.md).

## Reflection dump (SDK)
Headers and indexes generated from a running game's reflection data: classes, structs, enums, functions with flags, and member offsets. Owned by [techniques/reflection-dumping.md](techniques/reflection-dumping.md).

## RPC index
`protocol/rpc-index.md`: every UFunction with a network flag (Server, Client, NetMulticast) from the reflection dump; the client→server request surface in text. Owned by [techniques/reflection-dumping.md](techniques/reflection-dumping.md).

## Preservation server
An independent server, written from the documented protocol, that a client you own connects to after the official servers are discontinued. It contains no publisher server code; reconstructed server rules are inferences from client evidence. Owned by [operations/game-preservation.md](operations/game-preservation.md).

## Runtime gate
The [checklists/runtime-session-gate.md](checklists/runtime-session-gate.md) checklist, committed before any action that runs, attaches to, injects into, debugs, or intercepts the client. Owned by [foundations/scope-legal-and-ethics.md](foundations/scope-legal-and-ethics.md).

## Snapshot (REA)
A provider-neutral analysis snapshot REA imports on `open_binary` so later sessions skip auto-analysis; bound to one binary hash, provider, and analysis profile. Owned by [recipes/open-the-client-in-rea.md](recipes/open-the-client-in-rea.md).

## Twin build
A self-built shipping executable of a blank project on the same engine version, kept with its PDB, used to port engine symbols to the stripped target. Owned by [techniques/static-binary-analysis.md](techniques/static-binary-analysis.md).

## Unknown
An open question recorded with REA `record_unknown` and in the ledger, naming what evidence would settle it. Owned by [foundations/agentic-re-workflow.md](foundations/agentic-re-workflow.md).

## Unversioned properties
Cooked serialization without property names or type tags (`PKG_UnversionedProperties`); readers need a mapping file. Owned by [foundations/unreal-engine-internals.md](foundations/unreal-engine-internals.md).
