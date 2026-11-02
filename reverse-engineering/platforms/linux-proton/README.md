# Linux / Proton Platform

The host-specific layer for analyzing a Windows target that runs on Linux through Steam Proton (Wine). The general method lives at the handbook root ([../../README.md](../../README.md)); this section adds the lab, debugger, and prefix mechanics specific to Proton on an Arch-based host (CachyOS is the reference).

## Docs

- [lab-setup-linux-proton.md](lab-setup-linux-proton.md) — the lab: native capture tools, a separate Wine tools prefix, REA's Ghidra/TShark wiring, btrfs snapshots, and the `ptrace_scope` policy.
- [dynamic-analysis-under-proton.md](dynamic-analysis-under-proton.md) — the Proton-specific mechanics of runtime observation: attaching GDB to a Wine process, `winedbg --gdb`, Frida's rough edges under Wine, and prefix cleanup. The general runtime method is in [../../techniques/dynamic-analysis.md](../../techniques/dynamic-analysis.md).

## Where To Go Next

- [../../foundations/agentic-re-workflow.md](../../foundations/agentic-re-workflow.md) — the workflow this lab supports.
- [../../techniques/dynamic-analysis.md](../../techniques/dynamic-analysis.md) — the platform-neutral runtime method.
