# Tooling Decisions

The single home for tool picks, version pins, and supported-engine ranges in this handbook. Other docs name tools; only this file names versions. Every row records when it was checked.

## Default Approach

Pick one default per job, say when to reach for something else, and re-check versions on each new engine minor release or when a tool fails on a new build. A pin is a record of what was verified, not a promise that later versions break.

### Default Tool Per Job

| Job | Default | Reach for something else when |
|---|---|---|
| Agent evidence layer | REA (MCP + `rea` CLI) | the task is ordinary source-repo analysis (REA's skill says to skip it there) |
| Native decompiler for Win64 PE on Linux | Ghidra through REA | an IDA license exists; REA can bind IDA via its MCP adapter |
| Symbol porting from a twin build | Ghidra Version Tracking | the team already uses BinDiff |
| Container browsing and export | FModel (GUI), CUE4Parse (scripts) | IoStore conversion is needed: retoc; legacy pak work: repak |
| Single-asset inspection and Kismet bytecode | UAssetGUI / UAssetAPI | bulk bytecode analysis: kismet-analyzer, after validation |
| AES key recovery from own client | AESDumpster | — (if static recovery fails, stop) |
| Reflection dump and `.usmap` (gated) | RE-UE4SS experimental build | one-shot SDK without a mod runtime: Dumper-7 |
| HTTPS interception | mitmproxy | — |
| Packet capture and decoding | `dumpcap` + TShark, Lua dissectors | interactive work: Wireshark GUI |
| Protobuf descriptor recovery | protodump | pbtk, for its broader extraction and replay tooling |
| Debugger under Proton (gated) | GDB, via `winedbg --gdb` when Wine module info is needed | — |
| Function hooking under Proton (gated) | none by default; Frida is unsupported under Wine | validated lab setup per [../platforms/linux-proton/dynamic-analysis-under-proton.md](../platforms/linux-proton/dynamic-analysis-under-proton.md#frida-under-wine-unsupported) |

### Verified Versions

Checked 2026-11-02 against each project's releases page or the reference host.

| Tool | Version checked | Engine range or note | Source |
|---|---|---|---|
| REA (`rea-agents`) | 6.7.0 | Node.js 22.19+, 24.11+, or 26+ | reference host, `rea --version`; README |
| Ghidra | 12.1.4 | REA's doctor asks for a 12.1.x release | reference host; `rea doctor` remediation text |
| Wireshark CLI (`tshark`) | 4.7.4 | REA packet tools require TShark 4.x | reference host, `pacman -Q wireshark-cli` |
| mitmproxy | 12.3.1 | local capture mode needs Linux 6.8+ | reference host; mitmproxy docs |
| protontricks | 1.14.1 | — | reference host |
| RE-UE4SS | v3.0.1 stable (2024-02-14); `experimental-latest` updated 2026-11-02 | UE 5.5/5.6 support only in the experimental build ("v4.0.0-rc1 / TBD" changelog) | `github.com/UE4SS-RE/RE-UE4SS/releases` |
| Dumper-7 | v7.1.0 (2026-10-16) | "all of UE4 and UE5"; open UE5.6 crash report, issue #490 | `github.com/Encryqed/Dumper-7` |
| FModel | "FModel Nov 2026" (2026-10-31) | built on CUE4Parse | `github.com/4sval/FModel/releases` |
| CUE4Parse | 1.2.2.202611 (2026-11-01) | defines `GAME_UE5_6`; latest `GAME_UE5_9` | `CUE4Parse/UE4/Versions/EGame.cs` |
| retoc | v0.1.5 (2026-01-02) | "5.3+ are well supported" | `github.com/trumank/retoc` |
| repak | v0.2.3 (2026-01-02) | pak versions to "4.26-5.3" tested; later "likely supported but untested" | `github.com/trumank/repak` |
| AESDumpster | 1.3 (2025-08-15) | "5.0 to 5.6 (Supported)" | `github.com/GHFear/AESDumpster/releases/tag/1.3` |
| UAssetGUI | v1.1.0 (2026-05-01) | "Added support for Unreal Engine 5.6 and 5.7" | `github.com/atenfyr/UAssetGUI/releases/tag/v1.1.0` |
| kismet-analyzer | development pre-release (2026-01-11) | supported engine versions undocumented | `github.com/trumank/kismet-analyzer` |
| Unreal Engine 5.6 | released 2025-06-03 | Iris labeled "Experimental" on Epic's Iris page | Epic forum announcement "Unreal Engine 5.6 Released" |

### Recording A New Decision

Changes to the defaults above, and any change to the boundaries in [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md), are recorded per [README.md](README.md) before the docs change.

## Common Mistakes And Forbidden Patterns

- Writing a tool version into a topic doc or recipe instead of this file.
- Updating a version row without re-checking the source and the date.
- Adopting a tool whose primary purpose is anti-cheat bypass, cheat development, or live-service automation.

## Verification And Proof

- `grep -rnE 'v[0-9]+\.[0-9]+\.[0-9]+' reverse-engineering --include=*.md` finds version strings only in this file (quoted tool output elsewhere excepted).
- Each Verified Versions row has a source and a check date.

## Where To Go Next

- [README.md](README.md) — how decisions are recorded.
- [../platforms/linux-proton/lab-setup-linux-proton.md](../platforms/linux-proton/lab-setup-linux-proton.md) — installing the defaults.
