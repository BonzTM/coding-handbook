# Reflection Dumping

How to turn a UE5 game's reflection data into an SDK (headers with classes, structs, enums, functions, and offsets) and a `.usmap` mapping file — where runtime dumping is in scope — and what static work substitutes where it is not.

## Default Approach

A reflection dump names everything: every reflected class and member offset, every enum value, every UFunction including RPCs with their parameters. It turns a stripped executable into a labeled one. The best dumps come from the running game, because the reflection objects exist only after startup. That makes dumping a runtime action, gated by [../../checklists/runtime-session-gate.md](../../checklists/runtime-session-gate.md) and out of scope on anti-cheat-protected live clients.

### When Runtime Dumping Is In Scope

Dump at runtime only when one of these holds and `scope.md` records it:

- The client ships no anti-cheat, and the session is offline, at the main menu without logging in, or against a private server.
- The game offers an offline or single-player mode that runs without the protected launcher.
- Written authorization covers it.

Otherwise, skip to [Static Alternatives](#static-alternatives).

### Dumper Choice

| Tool | What it produces | Notes |
|---|---|---|
| RE-UE4SS | Lua and C++ mod runtime; live object viewer ("Search, view, edit & watch the properties of every loaded object"); C++ headers "with offsets"; UHT-compatible headers; `.usmap` | Targets "UE Versions: From 4.7 To 5.8"; the changelog lists "Added support for UE Version 5.6" under an unreleased version, so 5.5/5.6 support ships only in the `experimental-latest` build, not the current stable release (`github.com/UE4SS-RE/RE-UE4SS`, `assets/Changelog.md`; versions in [../../decisions/tooling.md](../../decisions/tooling.md)). |
| Dumper-7 | C++ SDK, `.usmap` (`MappingGenerator`), IDA mappings (`.idmap`), Dumpspace output | README: "Supported versions are all of UE4 and UE5." Built as a DLL and injected; default output `C:\Dumper-7`, configurable in `Dumper-7.ini` (`github.com/Encryqed/Dumper-7`). |

Default to UE4SS when you also want the live viewer and Lua for observation; default to Dumper-7 for a one-shot SDK. Check open issues for the exact engine minor before a session: Dumper-7 has an open "[UE5.6] Crash" report (issue #490).

### UE4SS Under Proton

UE4SS loads through a proxy DLL. Its changelog says "only the proxy DLL (dwmapi.dll) remains in the game's Win64 directory", with everything else in a `ue4ss/` subfolder; the install guide adds that the proxy "can have a name of any DLL that is loaded by the game engine" (`docs.ue4ss.com/installation-guide.html`). Settings live in `UE4SS-settings.ini`.

Under Proton, Wine must prefer the native proxy over its built-in DLL. Wine's override syntax abbreviates the load order as `n` (native) and `b` (builtin) (`wine` man page), and Proton appends its own overrides to a user-set `WINEDLLOVERRIDES` (Proton `proton` script, `append_to_env_str`). A UE4SS collaborator gives the launch option as `WINEDLLOVERRIDES="dwmapi.dll=n,b" %command%` (`github.com/UE4SS-RE/RE-UE4SS/issues/1189#issuecomment-3904969947`); the same thread reports breakage on newer Proton builds. The README states "Native Linux builds are not currently supported". Treat UE4SS on Proton as working-but-unsupported and verify in the lab prefix first.

Configure dump-only behavior for a reflection session: disable Lua mods you did not write, enable the dumpers you need, and remove the proxy DLL from `Win64/` after the session so the next launch is clean.

### Outputs And Where They Go

- SDK headers → `sdk/<build-id>/` (git-ignored or LFS).
- `.usmap` → `keys/<build-id>.usmap` (git-ignored, agent-denied).
- A summary of RPCs (class, function, flags, parameters) → generated into `protocol/rpc-index.md` by a script, committed. This is the client→server request surface in text form.

### Feeding The Dump To The Agent

The dump is large. Do not paste it into context.

- Index it: a script extracts classes, functions with network flags, enums, and struct layouts into a compact JSON or SQLite file under `sdk/<build-id>/index.*`. REA `inspect_sqlite_database` can then answer questions about it with paging.
- Name the binary from it: where the dump records function addresses or offsets relative to the module base, a script converts them to analysis addresses and writes `builds/<build-id>/names.json`, which the agent applies with REA `set_addresses_names`.
- Import types: Ghidra's C parser can "extract data type information from C-Header files" (Ghidra help, `CParserPlugin/CParser.htm`). Import the struct and class headers so decompiler output shows field names. Expect to trim headers that rely on C++ features the parser rejects.

### Static Alternatives

When runtime dumping is out of scope:

- **Asset-side reflection** — cooked Blueprint classes and user-defined structs carry their own property descriptions in the assets. Export them with the asset tools ([asset-extraction.md](asset-extraction.md)).
- **Registration data in the binary** — UHT-generated registration code for native classes is compiled into the executable, with class, function, and property names as strings. Find a known reflected name with REA `search_strings`, follow `xrefs` to the registration tables, and decode their layout against the matching engine source. This is slow, manual, and per-version, but fully static.
- **Engine-only symbols** — porting names from a self-built twin ([../../techniques/static-binary-analysis.md](../../techniques/static-binary-analysis.md#port-symbols-from-a-self-built-twin)) names the engine half of the binary without touching the game at runtime.

Never obtain a dump or mapping by downloading one posted by others for a live-service game; you cannot verify its build, and redistributing it is not yours to authorize.

## Common Mistakes And Forbidden Patterns

- Injecting UE4SS or Dumper-7 into an anti-cheat-protected client connected to live services.
- Leaving the proxy DLL in `Win64/` after the session.
- Logging into a live account with an injected client.
- Pasting SDK headers into agent context instead of indexing them.
- Reusing a dump across builds without re-validating offsets.
- Running mods you did not write during a reflection session.

## Verification And Proof

- The runtime gate checklist for the session is committed under `briefs/` before the dump exists.
- `sdk/<build-id>/` and `keys/<build-id>.usmap` exist and are git-ignored.
- `protocol/rpc-index.md` lists every UFunction with a network flag from the dump, with counts matching the index.
- A sample of five names from `names.json` resolves in REA to functions whose decompiled bodies are consistent with their names.

## Where To Go Next

- [recipes/dump-the-reflection-sdk.md](recipes/dump-the-reflection-sdk.md) — the gated procedure.
- [../../techniques/static-binary-analysis.md](../../techniques/static-binary-analysis.md) — using the dump in REA.
- [unreal-engine-internals.md](unreal-engine-internals.md) — what the dump describes.
