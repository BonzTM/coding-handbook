# Dynamic Analysis Under Proton

How to observe a UE5 client at runtime on Linux when — and only when — the runtime gate allows it: debugging through GDB and winedbg, Frida's current state under Wine, UE4SS as an observation tool, and how REA records each session as Evidence.

## Default Approach

Runtime work answers what static work cannot: actual values, actual call order, decrypted buffers just before they hit the socket. It also carries all the risk. Every session passes [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) first. On a client with anti-cheat connected to live services, nothing in this document applies; the gate fails at its first group.

Prefer the least invasive tool that answers the question: a log file, then a passive capture, then a debugger breakpoint, then an injected hook.

### Where Runtime Work Is Viable

- A client with no anti-cheat, in an offline mode, at the main menu without logging in, or against a private or emulated server.
- A dedicated-server build the publisher ships for self-hosting. Many UE games publish one; it is the server half, runs locally, and is the most useful runtime target for protocol interoperability work. Check its license terms separately.
- An authorized engagement whose scope names runtime instrumentation.

### Debugging With GDB And winedbg

- Valve's Proton debugging guide states attaching GDB to a running `.exe` "works out of the box" and documents `PROTON_WAIT_ATTACH=1 %command%` to hold the game until a debugger attaches; it also covers loading Wine's GDB helper scripts (`github.com/ValveSoftware/Proton/blob/proton_11.0/docs/DEBUGGING-LINUX.md`).
- `winedbg --gdb` runs winedbg "as a proxy for gdb"; `--no-start` prints a `target remote localhost:<port>` line to connect from a separate GDB (`winedbg` man page). This gives GDB Wine-aware module information for PE images.
- Yama `ptrace_scope` decides whether a non-parent debugger may attach ([../foundations/lab-setup-linux-proton.md](../foundations/lab-setup-linux-proton.md#debugger-and-ptrace-policy)).
- Use hardware or software breakpoints on functions found statically (writers, readers, `ProcessEvent`), log registers and buffers, and continue. Do not patch code or write memory.

REA can host a GDB session (`start_gdb_session`, `gdb_console`) so the transcript is Evidence. REA's persistent GDB sessions are Linux-only and verify the inferior's output descriptors before detaching (`start_gdb_session` contract). These tools are on the agent's ask list; a human approves each command batch.

### Frida Under Wine: Unsupported

Frida has no official Wine or Proton support. Reported state as of the date in [../decisions/tooling.md](../decisions/tooling.md):

- Attaching from the host to a Wine process fails with "Unable to locate the libc" (`github.com/frida/frida/issues/3339`). The maintainer suggests preloading a frida-gadget DLL through Wine instead.
- A community write-up reports running a Windows `frida-server` inside a Wine prefix (`suidpit.sh/posts/freat-intro/`); this is not official.
- An open issue on CachyOS reports hooks failing with remote `frida-server` (`github.com/frida/frida/issues/3617`), and another reports `DebugSymbol.fromAddress` not working under Wine (`github.com/frida/frida/issues/2734`).

If Frida is used, run `frida-server.exe` in the lab prefix and point REA's `instrument_with_frida` or `start_frida_session` at it with the `remote` address option. Treat every result as needing corroboration from a debugger or a capture until the setup is validated on your host. Frida injection is a runtime action on the ask list.

### UE4SS As An Observation Tool

Where UE4SS is in scope ([reflection-dumping.md](reflection-dumping.md#when-runtime-dumping-is-in-scope)), its live viewer can "search, view, edit & watch the properties of every loaded object", and Lua can hook UFunctions by name. Use it read-only:

- Watch replicated properties change in response to scripted actions.
- Log RPC calls by hooking the UFunction, recording parameters to a file under `captures/`.
- Never edit properties or call functions on a client connected to a live server.

### Observing Encrypted Traffic At The Boundary

When UDP payloads are encrypted by a packet handler, or TLS hides an HTTP body, the plaintext exists in the client just before encryption and just after decryption. A breakpoint or hook on the handler's encrypt/decrypt boundary logs plaintext buffers. That is observation of your own client's data in your own session; it does not modify traffic. It is still runtime instrumentation and needs the gate.

### Record Every Session

- Save debugger transcripts, hook logs, and captures under `captures/` with the build ID and date.
- Import what REA can read (Frida and GDB session output, captures) so findings cite Evidence IDs.
- Mark ledger rows from live sessions as `live-runtime`, and from saved captures replayed offline as `recorded-runtime`.

## Common Mistakes And Forbidden Patterns

- Running any of this against an anti-cheat-protected client connected to live services.
- Letting the agent start the game, attach a debugger, or load a Frida script without a human approving that exact action.
- Writing memory, patching code, or calling game functions in a live session.
- Trusting Frida output under Wine without corroboration.
- Leaving `PROTON_WAIT_ATTACH`, DLL overrides, or proxy DLLs in place after the session.
- Lowering `ptrace_scope` permanently.

## Verification And Proof

- The session's gate checklist and action plan are committed before the session.
- After the session, Steam launch options and the game's `Win64/` directory are back to their pre-session state (compare against `hashes.sha256`).
- Every runtime-derived ledger row cites a saved session artifact and is typed `live-runtime` or `recorded-runtime`.

## Where To Go Next

- [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) — the gate.
- [network-protocol-analysis.md](network-protocol-analysis.md) — where most runtime observations end up.
- [../foundations/lab-setup-linux-proton.md](../foundations/lab-setup-linux-proton.md) — prefixes, snapshots, and ptrace policy.
