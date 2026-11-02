# Lab Setup On Linux With Proton

How to build a reverse-engineering lab on an Arch-based Linux host (CachyOS is the reference) for a Windows UE5 game that runs under Steam Proton: native Linux tools first, Windows-only tools in a separate Wine prefix, REA wired to Ghidra and TShark, and btrfs snapshots for rollback.

## Default Approach

Proton changes less than it seems. The game is still a Windows PE executable, its assets are the same files, and its network traffic leaves through the host's network stack. That gives a Linux host three advantages: native packet capture of the game's sockets, native analysis tools that never run inside the game's environment, and filesystem snapshots. It also adds one cost: Windows-only RE tools need a Wine prefix, and in-process tooling behaves differently under Wine.

### Package Baseline

Install from the distro repositories where possible (`pacman -S`), and record versions in [../../decisions/tooling.md](../../decisions/tooling.md):

| Need | Package or source | Notes |
|---|---|---|
| Packet capture CLI | `wireshark-cli` (provides `tshark`, `capinfos`, `dumpcap`) | REA's pcap tools call these; REA "never captures traffic". |
| Packet capture GUI | `wireshark-qt` | Optional; for interactive dissector work. |
| HTTP(S) interception | `mitmproxy` | Writes native flow files that REA `inspect_web_network_capture` reads. |
| Proton prefix tooling | `protontricks` (provides `protontricks-launch`) | Runs a Windows `.exe` inside a game's prefix. |
| Wine for a separate tools prefix | `wine` or `umu-launcher` (`umu-run`) | Keeps Windows RE tools out of the game prefix. |
| Debugger | `gdb` | Wine also provides `winedbg`. |
| Disassembler/decompiler | Ghidra (official release zip, unpacked under `~/.local/opt/`) | REA's Windows-PE provider on Linux. |
| REA | `npm install --global rea-agents`, then `rea setup` | Node.js 22.19+, 24.11+ or 26+ per the REA README. |
| .NET runtime for FModel-family tools | per tool README | FModel and UAssetGUI are Windows .NET desktop apps; run them in the tools prefix. |

### Know Where Proton Puts Things

- Game files: `~/.steam/steam/steamapps/common/<Game>/`.
- Game prefix: `~/.steam/steam/steamapps/compatdata/<appid>/pfx/` — a normal Wine prefix (`drive_c/`, `user.reg`, `system.reg`).
- Build ID: `~/.steam/steam/steamapps/appmanifest_<appid>.acf`, key `"buildid"`.
- Proton log: set Steam launch options to `PROTON_LOG=1 %command%`; Proton writes `steam-<appid>.log` to `$HOME` (`github.com/ValveSoftware/Proton#runtime-config-options`).
- UE logs: inside the prefix under `drive_c/users/steamuser/AppData/Local/<Project>/Saved/Logs/`. Shipping builds log less than development builds, but the first lines usually include the engine version and command line.

Find the app ID with `protontricks -l` or `protontricks -s <name>`; both are documented in `protontricks --help`.

### Two Prefixes, Never One

Never install RE tools into the game's prefix. The game prefix is what the game, its launcher and any anti-cheat see.

- **Game prefix** — untouched except for what a deliberate, gated runtime session needs ([../../checklists/runtime-session-gate.md](../../checklists/runtime-session-gate.md)).
- **Tools prefix** — a separate Wine prefix (`WINEPREFIX=~/re/tools-prefix`) for FModel, UAssetGUI, and other Windows-only static tools. These tools read files; they never need to share a prefix with the game.

`protontricks-launch --appid <appid> <tool.exe>` runs an executable inside the game's prefix (`protontricks-launch --help`). Reserve it for the rare gated case that needs the game's prefix; it is denied to agents by default in [../../templates/re-workspace-agents.md](../../templates/re-workspace-agents.md).

### Snapshot Before Anything Runs

The CachyOS wiki states "BTRFS is the default and recommended filesystem for CachyOS" (`wiki.cachyos.org/installation/filesystem/`). Its snapshot page warns that "CachyOS's default Snapper configuration only covers the root subvolume (`/`)", and `/home` is a separate subvolume (`wiki.cachyos.org/configuration/btrfs_snapshots/`). Btrfs snapshots are also "not recursive" across nested subvolumes (`btrfs.readthedocs.io/en/latest/Subvolumes.html`). A Steam library under `/home` is therefore not protected by the default root snapshots.

- Create a Snapper config for the subvolume that holds the Steam library, or put the library on a dedicated subvolume.
- Take a manual snapshot before any runtime session: `sudo snapper -c <config> create --type single --description "pre-<build-id>-<action>"`.
- Without a suitable subvolume, copy `compatdata/<appid>/` aside instead.

Snapshots of game files are local only; they are never redistributed.

### Network Visibility

A Proton game runs as an ordinary Linux process tree, and its connections are visible to host tools without anything running in the prefix:

- `ss -tunap | grep -i <exe-name>` lists the client's TCP/UDP sockets and remote endpoints while it runs. Confirm this on your host during the first gated session and record it as an observation.
- `dumpcap -i <iface> -f "host <server-ip>" -w captures/<build-id>-<action>.pcapng` captures them. Capture permissions on Arch come from the `wireshark` group set up by the `wireshark-cli` package; add the user to the group rather than running capture as root.
- HTTPS interception with mitmproxy needs the client to trust mitmproxy's CA. Methods and limits live in [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md#tls-and-the-http-backend).

### Debugger And ptrace Policy

Yama controls ptrace "through sysctls in /proc/sys/kernel/yama". At `ptrace_scope` `1` (restricted), attaching needs "a predefined relationship with the inferior", such as being its parent (`docs.kernel.org/admin-guide/LSM/Yama.html`). The reference CachyOS host reads `1`. Valve's Proton debugging guide says attaching GDB to a running `.exe` "works out of the box" and documents `PROTON_WAIT_ATTACH=1 %command%` to pause the game until a debugger attaches (`github.com/ValveSoftware/Proton/blob/proton_11.0/docs/DEBUGGING-LINUX.md`). Test attach under your host's `ptrace_scope` before relying on it; do not lower `ptrace_scope` permanently. Debugging methods live in [dynamic-analysis-under-proton.md](dynamic-analysis-under-proton.md).

### REA And Ghidra Provider

1. Unpack the Ghidra release under `~/.local/opt/` and install the JDK version its release notes name.
2. Install REA (`npm install --global rea-agents`) and run `rea setup`, choosing the agent clients to register. Setup shows proposed configuration changes and backs up existing files before writing them (REA README, "Set up your agent").
3. Point REA at Ghidra with `GHIDRA_INSTALL_DIR`. `rea doctor` reports the remediation verbatim when it is missing: "Set GHIDRA_INSTALL_DIR to an extracted Ghidra 12.1.x release directory." Set it in two places: the `env` block of the REA MCP server registration (the agent's MCP config), and the shell profile for CLI use. Details: `github.com/morluto/rea/blob/main/docs/installation.md`.
4. Set `REA_GHIDRA_STARTUP_TIMEOUT_MS` high enough for a UE5 executable in the environment that launches the agent. REA's README: "For large binaries, raise the provider's startup deadline".
5. For field-level capture inspection, set `REA_TSHARK_COMMAND` to the absolute path of `tshark`. Without it, `binary_session` reports `inspect_packet_capture` as `provider_missing`, while `inspect_pcap_capture` still works with `tshark` and `capinfos` on PATH.
6. On Linux, REA needs a procps-compatible `ps` on its `PATH` for Ghidra process ownership checks (REA README, "Quick start").
7. Restart the agent client after setup. Verify by calling `binary_session` from the agent and reading `analysis_provider_candidates`.

### Hardware And Time Budget

UE5 shipping executables are large; three UE5 titles on the reference host range from 138 MB to 291 MB. Ghidra's first auto-analysis of a binary that size is long and memory-hungry. REA exposes three settings for it, read from the environment that starts REA (REA `dist/ghidra/GhidraJavaLaunch.js` and `GhidraProviderClient.js`, v6.7.0):

| Setting | Effect |
|---|---|
| `GHIDRA_HEADLESS_MAXMEM` (or `GHIDRA_MAXMEM`) | Ghidra JVM max heap. REA defaults to `2G`, which is too small for a UE5 client; set it from available RAM, for example `24G`. |
| `REA_GHIDRA_STARTUP_TIMEOUT_MS` | Deadline for import plus auto-analysis. REA's remediation text: "Large binaries can need longer". |
| `REA_GHIDRA_DISABLED_ANALYZERS` | Skips named expensive Ghidra analyzers. Changing it changes the analysis profile, so a snapshot made under one value does not reopen under another. |

REA also warns that "an MCP client's own tool-call timeout is separate and may also need raising". Run the first analysis from a shell instead of through the agent's MCP call, so no client timeout applies: `rea decompile <exe> <known-function-or-address> --provider ghidra --snapshot <path>` in a background terminal. Later agent sessions open the same snapshot. Procedure: [../../recipes/open-a-target-in-rea.md](../../recipes/open-a-target-in-rea.md).

## Common Mistakes And Forbidden Patterns

- Installing RE tools, DLL overrides, or mod loaders into the game's prefix "to try something".
- Running Windows RE tools from inside an extracted-asset directory; tools and target data live in separate directories.
- Setting `ptrace_scope` to `0` permanently for convenience.
- Running packet capture as root instead of through the `wireshark` group.
- Letting an agent call `protontricks-launch`, `wine`, or `umu-run` without a human gate.
- Re-running Ghidra auto-analysis each session instead of reusing the snapshot.
- Skipping the btrfs snapshot before a runtime session.

## Verification And Proof

```bash
rea doctor --provider ghidra --json | jq -c '.scope_checks[] | select(.name=="ghidra") | .ok'   # true
command -v tshark capinfos dumpcap mitmproxy protontricks-launch
cat /proc/sys/kernel/yama/ptrace_scope        # expect 1 (unchanged)
findmnt -no FSTYPE /home                     # btrfs on a default CachyOS install
```

The lab is ready when REA reports Ghidra available, the capture tools resolve, a separate tools prefix launches a Windows static tool, and a snapshot of the Steam library exists.

## Where To Go Next

- [../../foundations/agentic-re-workflow.md](../../foundations/agentic-re-workflow.md) — how the agent uses this lab.
- [../../domains/games/recipes/fingerprint-a-ue5-build.md](../../domains/games/recipes/fingerprint-a-ue5-build.md) — the first thing to do with a new build.
- [../../decisions/tooling.md](../../decisions/tooling.md) — pinned tool versions.
