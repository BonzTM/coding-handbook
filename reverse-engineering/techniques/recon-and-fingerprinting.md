# Recon And Fingerprinting

How to identify exactly what a UE5 client build is, how its content is packaged, which protections and SDKs it ships, and which techniques that rules in or out — before opening a disassembler.

## Default Approach

Fingerprinting is cheap, static, and decides the rest of the engagement. It answers four questions, each recorded as a ledger finding: which engine version, which container format, which protections, which online backends. Run it first on every new game and again on every patch ([../recipes/fingerprint-a-ue5-build.md](../recipes/fingerprint-a-ue5-build.md)).

### Install Layout

A Steam UE5 game typically looks like this:

```text
<Game>/
  <Launcher>.exe                    # small bootstrap; often starts the real binary
  Engine/                           # engine content and third-party DLLs
  <Project>/
    Binaries/Win64/<Project>-Win64-Shipping.exe   # the game module
    Content/Paks/*.pak *.utoc *.ucas *.sig
  EasyAntiCheat/ | BattlEye/        # present only if that anti-cheat ships
```

The `-Shipping.exe` is the analysis target. The launcher is not.

### Engine Version

- The engine build string is the fastest signal. On stock Epic branches it reads `++UE5+Release-5.<minor>`; studios that fork the engine use their own branch names. Search both UTF-16 and ASCII strings.
- A missing or custom branch string means the version is narrowed by other evidence: PE version resources (REA `inspect_pe_resources`), container format and header versions, and which engine version the asset tools parse cleanly. Record a narrowed version as an inference.
- Pin the version before choosing tool settings. UE4SS, Dumper-7, CUE4Parse, retoc, and AESDumpster all have per-version behavior; their supported ranges live in [../decisions/tooling.md](../decisions/tooling.md).

### Containers

- `.utoc`/`.ucas` pairs mean IoStore. A small `.pak` usually ships alongside.
- `.sig` files mean signed paks. Signing exists "to prevent tampering of the data" (Epic Project Settings, "Signing"). Do not modify or replace signed containers on a client that connects to live services.
- Container directory listings are often readable without a key; file payloads are not. A container that lists but will not extract is encrypted.

### Protections

Record presence, not internals:

- **Anti-cheat** — `EasyAntiCheat/` with `start_protected_game.exe`-style launchers, `BattlEye/`, or another vendor directory. Presence sets the runtime scope per [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md#anti-cheat-decides-runtime-scope).
- **Packers/protectors** — non-standard PE section names and high-entropy code sections are visible in the PE header. A packed executable changes what static analysis can see: REA's guidance is that "decompiling an unpacking stub does not recover the unpacked program". Unpacking a protected executable is circumvention of a technical protection measure and is out of scope here.
- **Debug data** — a shipped `.pdb` is rare and valuable; load it into Ghidra before other analysis. Ghidra's PDB help says to load the PDB "prior to other analysis".

### Online Backends

- DLLs under `Binaries/Win64` and `Engine/Binaries/ThirdParty` name the SDKs: Epic Online Services, Steamworks, vendor telemetry, voice chat, crash reporting.
- Cooked config (`DefaultEngine.ini`) names the NetDriver class, online subsystem, and often backend hostnames. Extract it with the asset tools ([asset-extraction.md](asset-extraction.md)).
- Strings in the executable give URL templates and REST paths. Search for `https://` and the backend hostnames, then page.

### Agent Role

Fingerprinting is fully static and agent-safe. An agent can run the shell steps in the recipe, call `inspect_pe_resources`, and draft `fingerprint.md`. A human confirms the anti-cheat finding, because it gates every runtime decision after it.

## Common Mistakes And Forbidden Patterns

- Analyzing the launcher instead of the `-Shipping.exe`.
- Recording "no anti-cheat" because one vendor's directory is absent; record "not observed" with what was checked.
- Choosing tool versions before the engine version is pinned.
- Attempting to unpack a protected executable.
- Modifying containers, `.sig` files, or the install directory during recon.

## Verification And Proof

- `builds/<build-id>/fingerprint.md` is complete and every line cites a command output or Evidence ID.
- The anti-cheat line is reviewed by a human and mirrored in `scope.md`.
- `sha256sum -c` against `hashes.sha256` passes for the recorded files.

## Where To Go Next

- [../recipes/fingerprint-a-ue5-build.md](../recipes/fingerprint-a-ue5-build.md) — the procedure.
- [asset-extraction.md](asset-extraction.md) — the next step for content.
- [../recipes/open-the-client-in-rea.md](../recipes/open-the-client-in-rea.md) — the next step for code.
