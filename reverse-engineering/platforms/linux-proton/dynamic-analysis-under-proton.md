# Dynamic Analysis Under Proton

The Linux/Proton-specific mechanics of runtime observation: attaching a debugger to a Windows PE running under Wine, winedbg's gdb proxy, the `ptrace_scope` gate, Frida's rough edges under Wine, and prefix cleanup. The general runtime method, the least-invasive ladder, and the evidence rules live in [../../techniques/dynamic-analysis.md](../../techniques/dynamic-analysis.md); this doc only adds what is specific to a Windows target running under Proton on Linux. Every session still passes [../../checklists/runtime-session-gate.md](../../checklists/runtime-session-gate.md) first, run by a human.

## Default Approach

A Proton game is a Windows PE executing under Wine as an ordinary Linux process tree. That makes host Linux debuggers usable, with Wine-specific wrinkles for PE module layout and the kernel's ptrace policy.

### Debugging With GDB And winedbg

- Valve's Proton debugging guide states attaching GDB to a running `.exe` "works out of the box" and documents `PROTON_WAIT_ATTACH=1 %command%` to hold the game until a debugger attaches; it also covers loading Wine's GDB helper scripts (`github.com/ValveSoftware/Proton/blob/proton_11.0/docs/DEBUGGING-LINUX.md`).
- `winedbg --gdb` runs winedbg "as a proxy for gdb"; `--no-start` prints a `target remote localhost:<port>` line to connect from a separate GDB (`winedbg` man page). This gives GDB Wine-aware module information for the PE images.
- Yama `ptrace_scope` decides whether a non-parent debugger may attach ([lab-setup-linux-proton.md](lab-setup-linux-proton.md#debugger-and-ptrace-policy)). Launch the target under the debugger, or use `PROTON_WAIT_ATTACH`, rather than lowering `ptrace_scope` permanently.

### Frida Under Wine: Unsupported

Frida has no official Wine or Proton support. Reported state as of the date in [../../decisions/tooling.md](../../decisions/tooling.md):

- Attaching from the host to a Wine process fails with "Unable to locate the libc" (`github.com/frida/frida/issues/3339`). The maintainer suggests preloading a frida-gadget DLL through Wine instead.
- A community write-up reports running a Windows `frida-server` inside a Wine prefix (`suidpit.sh/posts/freat-intro/`); this is not official.
- An open issue on CachyOS reports hooks failing with remote `frida-server` (`github.com/frida/frida/issues/3617`), and another reports `DebugSymbol.fromAddress` not working under Wine (`github.com/frida/frida/issues/2734`).

If Frida is used, run `frida-server.exe` in the lab prefix and point REA's `instrument_with_frida` or `start_frida_session` at it with the `remote` address option. Treat every result as needing corroboration until the setup is validated on your host.

### Prefix Hygiene

- Run every runtime session in a dedicated lab prefix or a btrfs snapshot you can roll back ([lab-setup-linux-proton.md](lab-setup-linux-proton.md#snapshot-before-anything-runs)).
- After the session, remove any injected libraries or proxy DLLs, clear `PROTON_WAIT_ATTACH` and `WINEDLLOVERRIDES` from the launch options, and confirm the game directory matches its recorded hashes.

## Common Mistakes And Forbidden Patterns

- Lowering `ptrace_scope` permanently instead of launching under the debugger.
- Trusting Frida output under Wine without corroboration from a debugger or capture.
- Leaving `PROTON_WAIT_ATTACH`, DLL overrides, or injected libraries in place after the session.
- Instrumenting in the game's own Proton prefix instead of the lab prefix.

## Verification And Proof

- After the session, Steam launch options and the game's `Win64/` directory match their pre-session hashes.
- The debugger attached under the host's existing `ptrace_scope`, with no permanent sysctl change left behind.

## Where To Go Next

- [../../techniques/dynamic-analysis.md](../../techniques/dynamic-analysis.md) — the general runtime method and evidence rules.
- [lab-setup-linux-proton.md](lab-setup-linux-proton.md) — prefixes, snapshots, and ptrace policy.
- [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md) — where boundary observations usually end up.
