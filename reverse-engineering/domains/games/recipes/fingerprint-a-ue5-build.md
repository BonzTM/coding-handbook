# Recipe: Fingerprint A UE5 Build

Use this first on every new game and every new build: record exactly what was installed, which engine version built it, how its content is packaged, and which protections ship with it. Governed by [../../../techniques/recon-and-fingerprinting.md](../../../techniques/recon-and-fingerprinting.md).

## Files To Touch

- `builds/<build-id>/fingerprint.md` (new)
- `builds/<build-id>/hashes.sha256` (new)
- `evidence/<build-id>-fingerprint-<yyyy-mm-dd>.json` (new)
- `findings.md` (fingerprint rows)

## Steps

1. Set variables:

   ```bash
   APPID=<appid>; GAME="$HOME/.steam/steam/steamapps/common/<Game>"
   BUILD=$(awk -F'"' '/"buildid"/{print $4}' "$HOME/.steam/steam/steamapps/appmanifest_${APPID}.acf")
   mkdir -p builds/$BUILD
   ```

2. Hash what you will analyze (executables, DLLs, containers, config), not the whole install:

   ```bash
   (cd "$GAME" && find . -type f \( -iname '*.exe' -o -iname '*.dll' -o -iname '*.pak' \
     -o -iname '*.utoc' -o -iname '*.ucas' -o -iname '*.sig' -o -iname '*.ini' \) -print0 \
     | sort -z | xargs -0 sha256sum) > builds/$BUILD/hashes.sha256
   ```

3. Find the game module: `find "$GAME" -path '*Binaries/Win64/*-Shipping.exe'`. The directory above `Binaries` is the project name.
4. Read the engine branch string without running anything:

   ```bash
   EXE="<path to -Shipping.exe>"
   { strings -el "$EXE"; strings -a "$EXE"; } | grep -m5 -E '^\+\+[A-Za-z0-9_]+\+[^ ]*Release'
   ```

   On the reference host this returned `++UE5+Release-5.4` for a stock-branch UE5 game and `++Project+SN2-Release` for a studio branch; a third UE5 game returned nothing. A studio branch name does not state the engine minor version.
5. When the branch string is missing or custom, narrow the version from other evidence and mark it as inference:
   - REA `inspect_pe_resources` on the `-Shipping.exe` for the version resource (`FileVersion`, `ProductVersion`, `CompanyName`).
   - Container format: `.utoc`/`.ucas` (IoStore) means UE5's Zen loader path is in use; the pak and IoStore header versions narrow the range. repak's README maps pak versions to engine ranges (for example pak v11 to "4.26-5.3").
   - Asset tools: open one container in FModel with candidate engine versions; the version that parses cleanly is evidence, not proof.
   - Later, the reflection dump or UE4SS log (gated runtime) states the version outright.
6. List content containers: `find "$GAME" -name '*.pak' -o -name '*.utoc' -o -name '*.ucas' -o -name '*.sig'`. `.utoc`/`.ucas` pairs mean IoStore; `.sig` files next to paks mean signed containers.
7. Detect protections by file presence and import names:
   - Anti-cheat: `EasyAntiCheat/` directory or `EasyAntiCheat_EOS*.exe`/`start_protected_game.exe` launchers; `BattlEye/` directory or `BEService*`; other vendor directories.
   - Packers and protectors: section names in the PE header (`objdump -h <exe>` via REA `inspect_with_objdump` if configured), and REA `search_strings` for vendor markers after the snapshot exists.
   - Debug symbols: any `*.pdb` shipped beside the executable.
8. Record online SDKs: DLL names under `Binaries/Win64` (EOS, Steamworks, PlayFab, vendor SDKs).
9. Fill `builds/$BUILD/fingerprint.md`:

   ```markdown
   # Fingerprint <build-id>
   - Steam build id: <id>; captured <yyyy-mm-dd>
   - Game module: <Project>/Binaries/Win64/<Project>-Win64-Shipping.exe (sha256 <...>)
   - Engine: <++UE5+Release-5.x string>; version resource <FileVersion>
   - Containers: <IoStore | pak>; count <n>; signed: <yes/no>; encrypted: <unknown until key test>
   - Anti-cheat: <none | vendor> (evidence: <path>)
   - Protector/packer: <none observed | name> (evidence: <...>)
   - PDB shipped: <yes/no>
   - Online SDKs: <list>
   ```

10. Add ledger rows for engine version, container format, and anti-cheat presence, each citing the Evidence ID from step 4 or the hash file line.
11. If anti-cheat is present, mark every runtime technique for this build as blocked in `scope.md` unless the session is offline, on a private server, or authorized.

## Invariants To Preserve

- Nothing in the game directory is executed or modified.
- `hashes.sha256` is committed; binaries are not.
- An absent marker is recorded as "not observed", not as "absent".

## Proof

- `sha256sum -c builds/$BUILD/hashes.sha256` (run from `$GAME` with adjusted paths) passes.
- `fingerprint.md` has no unfilled `<placeholder>`.
- The engine version in `fingerprint.md` matches the REA string search in [../../../recipes/open-a-target-in-rea.md](../../../recipes/open-a-target-in-rea.md).
