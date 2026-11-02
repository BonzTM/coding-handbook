# Dynamic Analysis

How to observe a target at runtime through REA when — and only when — the runtime gate allows it: the least-invasive-first ladder, debuggers, instrumentation with Frida, observing data at an encrypt/decrypt boundary, and recording every session as Evidence. Platform-specific mechanics (Linux/Proton, Windows, macOS) live in the matching `platforms/` doc.

## Default Approach

Runtime work answers what static work cannot: actual values, actual call order, a buffer's contents just before it is encrypted or sent. It also carries the risk, because it runs the target. Every session passes [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) first, and a human runs it. Where the target ships a protection that forbids instrumentation while it is connected to a live service, nothing here applies — the gate fails at its first group.

Prefer the least invasive tool that answers the question, in order: a log the target already writes, then a passive capture, then a debugger breakpoint, then an injected hook. Each step down adds risk and authorization cost.

### Where Runtime Work Is Viable

- A target with no instrumentation-blocking protection, run offline or against a test/private/emulated endpoint you control.
- A component the vendor ships for local or self-hosted use (a dedicated-server build, an offline mode, a standalone tool).
- An authorized engagement whose scope names runtime instrumentation.

If none of these hold, stay static. Protections that gate runtime access are out of scope ([../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md)).

### Debuggers

- Set hardware or software breakpoints on functions found statically (serializers, parsers, dispatchers, crypto boundaries), log registers and buffers, and continue. Observe; do not patch code or write memory in a passive session.
- REA can host a GDB session (`start_gdb_session`, `gdb_console`) so the transcript is Evidence. REA's persistent GDB sessions are Linux-only and verify the inferior's output descriptors before detaching (`start_gdb_session` contract). These tools are on the agent's ask list; a human approves each command batch.
- Attach policy, proxy debuggers, and OS gates (for a Windows target under Wine/Proton, `winedbg --gdb` and `ptrace_scope`) are platform-specific — see the matching `platforms/` doc, e.g. [../platforms/linux-proton/dynamic-analysis-under-proton.md](../platforms/linux-proton/dynamic-analysis-under-proton.md).

### Instrumentation With Frida

- REA runs a caller-authored Frida script against a selected process and detaches (`instrument_with_frida`, `start_frida_session`, `load_frida_script`). All are runtime actions on the ask list.
- Frida's support and limits vary sharply by platform and by how the target was loaded; a Windows PE running under Wine/Proton is a known rough edge. Check the platform doc before relying on it, and corroborate any Frida result with a debugger or a capture until the setup is validated on your host.

### Observing Data At A Boundary

When a payload is encrypted or compressed before it leaves the process, or decrypted just after it arrives, the plaintext exists in memory at that boundary. A breakpoint or hook on the encrypt/decrypt or serialize/parse function logs the plaintext buffer. That is observation of your own session's data; it does not modify traffic. It is still runtime instrumentation and needs the gate. It is the standard way past on-the-wire encryption ([network-protocol-analysis.md](network-protocol-analysis.md)) without touching the protocol's protection.

### Record Every Session

- Save debugger transcripts, hook logs, and captures under `captures/` with the build ID and date.
- Import what REA can read (Frida and GDB session output, captures) so findings cite Evidence IDs.
- Mark ledger rows from live sessions `live-runtime`, and from saved artifacts replayed offline `recorded-runtime`.

## Common Mistakes And Forbidden Patterns

- Running any of this against a target whose protection forbids instrumentation while it is connected to a live service.
- Letting the agent start the target, attach a debugger, or load a hook without a human approving that exact action.
- Writing memory, patching code, or calling target functions in a session meant to be passive.
- Trusting instrumentation output on a platform where the setup is unvalidated.
- Leaving debug hooks, injected libraries, or launch-option changes in place after the session.

## Verification And Proof

- The session's gate checklist and action plan are committed before the session.
- After the session, the target's files and launch configuration are back to their pre-session state (compare against recorded hashes).
- Every runtime-derived ledger row cites a saved session artifact and is typed `live-runtime` or `recorded-runtime`.

## Where To Go Next

- [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) — the gate every session passes.
- [static-binary-analysis.md](static-binary-analysis.md) — where the breakpoint targets come from.
- [network-protocol-analysis.md](network-protocol-analysis.md) — where boundary observations usually end up.
- Platform mechanics: [../platforms/linux-proton/dynamic-analysis-under-proton.md](../platforms/linux-proton/dynamic-analysis-under-proton.md).
