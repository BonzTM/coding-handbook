# Unreal Engine Internals For Analysts

The parts of Unreal Engine 5 that make a shipped game analyzable: the reflection system, the global object and name tables, `ProcessEvent`, cooked content and its containers, unversioned properties, and Blueprint bytecode. This is an analyst's map, not an engine tutorial; the engine source and Epic's API reference are the authorities.

## Default Approach

Unreal is unusually analyzable because the engine needs runtime type information for serialization, networking, Blueprints, and the editor. A shipping build strips symbols, but it keeps much of that type information. Most UE reverse engineering is the craft of recovering the engine's own metadata and letting it name everything else.

### Get The Matching Engine Source

Epic publishes the engine source to licensees on GitHub. Access requires that you have "associated your GitHub account with your Epic Games account" and accepted an invitation to the `@EpicGames` organization (`dev.epicgames.com/documentation/en-us/unreal-engine/downloading-source-code-in-unreal-engine`). Check out the tag matching the target's engine minor version.

The source is the best decompiler aid available:

- It defines every structure layout the binary uses, version-exact.
- A self-built shipping executable of an empty project, with its PDB, gives a symbol-rich twin of the engine code in the target. Port names from twin to target with Ghidra Version Tracking, which "enables the reverse engineer to port comments and labels into the new context" (Ghidra help, `VersionTrackingPlugin/Version_Tracking_Intro.html`). Procedure: [../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md#port-symbols-from-a-self-built-twin).
- Do not copy engine source into findings or publish it; the Unreal Engine EULA governs it.

### The Reflection System

Unreal Header Tool (UHT) reads `UCLASS`, `USTRUCT`, `UENUM`, `UPROPERTY`, and `UFUNCTION` markup and generates registration code compiled into the binary. At startup that code builds runtime objects describing every reflected type:

| Runtime type | Describes | Analyst value |
|---|---|---|
| `UClass` | a reflected class, its super class, its properties and functions | full class hierarchy with names |
| `UScriptStruct` | a reflected struct | field layouts for message and save structs |
| `UEnum` | a reflected enum with names and values | decodes opcodes, states, item types |
| `UFunction` | a reflected function, its parameters, and flags (including network flags) | the RPC surface; native thunk pointers |
| `FProperty` (`FField` subclasses) | one property: name, type, offset, flags | exact member offsets in live objects |

Because a `UFunction` carries network flags, every Server, Client, and NetMulticast RPC is enumerable by name and parameter list from reflection alone.

### Global Tables: Objects And Names

- **`GUObjectArray`** — the global `FUObjectArray` (`CoreUObject/Public/UObject/UObjectArray.h`) holding every live `UObject`. Walking it enumerates every class, function, and instance. Modern engines store it as a chunked array; Dumper-7 asks porters to add custom layouts to `FChunkedFixedUObjectArrayLayouts` "for UE4.21 and higher" (`github.com/Encryqed/Dumper-7#readme`).
- **Names** — every identifier is an `FName`: an index into a global name table plus a number suffix. Converting an `FName` to text requires the engine's name-pool logic. UE4SS locates it through the `FName::ToString` and `FName` constructor functions (`docs.ue4ss.com/dev/guides/fixing-compatibility-problems.html`).

Dumpers need exactly these anchors. UE4SS's compatibility guide lists `GUObjectArray`, `FName_ToString`, `FName_Constructor`, `StaticConstructObject`, and `GMalloc` as the signatures a game may need custom patterns for, and says the `GUObjectArray` pattern "must return the exact address of the global variable named 'GUObjectArray'".

### `ProcessEvent`: The Universal Call Gate

`UObject::ProcessEvent(UFunction* Function, void* Parms)` is "called by VM to execute a UFunction with a filled in UStruct of parameters" (Epic API reference, `UObject/Object.h`). Blueprint calls, many native-to-script calls, and RPC dispatch paths pass through it. It is a virtual function, so it sits at a fixed slot in every `UObject` vtable.

For the analyst it is the single best observation point: a log of `ProcessEvent` calls is a log of reflected function calls with names. Observing it is runtime instrumentation and goes through [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md).

### Cooked Content And Containers

Shipped games contain cooked assets: platform-specific, editor-data-stripped packages.

- **Pak files (`.pak`)** — the classic container. Optional AES-256 encryption (Epic: `FAES::FAESKey`, "a 256 bit AES key") and RSA signing. Project Settings names the sections "Encryption" and "Signing"; signing exists "to prevent tampering of the data", and signatures ship as `.sig` files described by `FPakSignatureFile` (Epic API reference; `project-section-of-the-unreal-engine-project-settings`).
- **IoStore (`.utoc` + `.ucas`)** — UE5's container format for the Zen loader. Epic: "Zen Loader is the new runtime loader for Unreal Engine 5" and "the .utoc file describes the container file, including chunk size and offset" (`dev.epicgames.com/documentation/en-us/unreal-engine/zen-loader-in-unreal-engine`). A small `.pak` often ships beside IoStore containers for loose and legacy content.
- **Encryption key** — one key or several (per-chunk keys). Without the key, encrypted containers expose only their directory, or nothing. Recovering a key from your own copy of the client is covered in [../techniques/asset-extraction.md](../techniques/asset-extraction.md#encryption-keys).
- **Config** — `DefaultEngine.ini`, `DefaultGame.ini`, and `DefaultInput.ini` are cooked into containers and reveal NetDriver classes, online subsystems, server URLs, and feature flags.

### Unversioned Properties And Mappings

Cooked UE5 packages usually serialize properties without names or type tags. Epic's package flag `PKG_UnversionedProperties` means the package "uses unversioned property serialization instead of versioned tagged property serialization" (Epic API reference, `EPackageFlags`). The reader must already know each class's property layout. Tools get it from a mapping file (`.usmap`) generated from the running game's reflection data. CUE4Parse fails without one: "Package has unversioned properties but mapping file is missing, can't serialize" (`CUE4Parse/UE4/Assets/AbstractUePackage.cs`). `.usmap` is a community format; UAssetAPI calls it "the unofficial .usmap format".

The consequence: full asset parsing of a modern UE5 game usually needs one runtime reflection dump, or a mapping file someone else produced from the same build. See [../techniques/reflection-dumping.md](../techniques/reflection-dumping.md).

### Blueprints And Kismet Bytecode

Blueprint graphs cook to Kismet bytecode stored in the asset's `ScriptBytecode`, executed by the engine's script VM. Game logic written in Blueprints is therefore in the containers, not the executable. UAssetAPI supports "reading and writing raw Kismet (blueprint) bytecode" (`github.com/atenfyr/UAssetAPI`). Analysis: [../techniques/blueprint-analysis.md](../techniques/blueprint-analysis.md).

### Shipping Builds Strip, But Not Everything

- No PDB, no symbols, inlined and optimized code, no `UE_LOG` below the configured shipping verbosity.
- Reflected names, enum strings, config keys, asset paths, and many log format strings remain.
- Networking Insights tracing (`-trace=net -NetTrace=<VERBOSITY>`) is a development tool; do not expect it in a shipping client.

## Common Mistakes And Forbidden Patterns

- Starting decompilation before obtaining the matching engine source tag.
- Hand-reversing engine structures (`FUObjectArray`, `FName`, `UClass`) that the source defines exactly.
- Assuming the engine minor version from the release date instead of the fingerprint.
- Treating an IoStore game as pak-only, or the reverse; most UE5 games ship both.
- Parsing unversioned assets with a mapping file from a different build.
- Copying engine source or decompiled engine code into public findings.

## Verification And Proof

- The engine source checkout tag matches the fingerprint's engine minor version, recorded in `builds/<build-id>/fingerprint.md`.
- Every structure layout used in analysis cites a source path in the matching engine tag.
- Asset parsing failures are traced to a missing key, missing mapping, or wrong engine version, recorded as an unknown, not ignored.

## Where To Go Next

- [../techniques/asset-extraction.md](../techniques/asset-extraction.md) — containers, keys, and mappings in practice.
- [../techniques/reflection-dumping.md](../techniques/reflection-dumping.md) — producing the SDK and `.usmap`.
- [../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md) — using reflection and the engine source in Ghidra through REA.
