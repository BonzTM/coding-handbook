# Maintainer Reference

Purpose: hold the slower-path engagement lifecycle, evidence taxonomy, rationale, and troubleshooting behind the fast-path rules.
Audience: maintainers of this handbook and agents running a long engagement.
Read [AGENTS.md](AGENTS.md) first. Use this file when you need the why.

## Engagement Lifecycle

```text
intake ──► fingerprint ──► static: assets ──► static: binary ──► recorded traffic ──► (gated) runtime
  │            │               │                   │                   │                    │
scope.md   builds/<id>/    exports/ keys/     snapshots/ names     captures/           briefs/ captures/
                                    └──────────────┴──── findings.md + evidence/ ─────────────┘
                                                              │
                                                      patch day: diff-a-patch
```

| Phase | Entry condition | Exit condition | Agent role |
|---|---|---|---|
| Intake | new game or engagement | intake checklist green | drafts scope from the human's purpose; never decides scope |
| Fingerprint | intake done | `fingerprint.md` complete, anti-cheat human-reviewed | runs static shell steps and `inspect_pe_resources` |
| Static: assets | fingerprint done | tracked tables exported by script | runs export scripts through `capture_process_scenario` |
| Static: binary | snapshot built from a shell | anchors and networking layer in ledger | primary analyst through REA |
| Recorded traffic | human captured a scripted session | classification and message specs drafted | analyzes saved files only |
| Runtime (gated) | gate checklist committed | session artifacts saved and ledgered | proposes plans; human executes |
| Patch day | new build ID | ledger re-verified or marked stale | runs pipeline, compares Evidence |

## Evidence Taxonomy

| Type | Example | Strength | Typical tool |
|---|---|---|---|
| static observation | string at address, decompiled statement, exported table row | strong for "the client contains X" | `search_strings`, `analyze_function`, export run |
| recorded-runtime observation | field value in a saved capture | strong for "the client sent/received X" | `inspect_packet_capture`, `inspect_web_network_capture` |
| live-runtime observation | breakpoint log, hook log | strongest for "at runtime X happened"; highest cost | GDB/Frida via REA (gated) |
| inference | "this function writes the move RPC" | as strong as the observations named | ledger row citing observations |
| unknown | "what does field 7 mean" | — | `record_unknown` |

A `high` confidence finding needs two independent evidence types; static plus recorded-runtime is the usual pair. This mirrors REA's rule that static analysis never implies observed execution.

## Why These Defaults

- **REA over raw tools.** Raw Ghidra or TShark output gives an agent text; REA gives it Evidence records with target hashes, coverage, and limitations, retained and pageable. That is what lets a ledger outlive a context window and lets a reviewer re-check a claim.
- **Snapshots built from a shell.** First analysis of a UE5 executable can exceed both REA's default Ghidra deadline and the MCP client's tool-call timeout. REA's own remediation text says the client timeout "is separate and may also need raising". A CLI run sidesteps both.
- **Names in a file, not the snapshot.** REA refuses immutable snapshot saves after metadata edits in the same session. Keeping names in `names.json` makes them reviewable, diffable, and portable across sessions.
- **Twin builds.** The engine half of the executable is reproducible from source; matching against a symbol-rich twin names it with far less effort and far more accuracy than manual work.
- **Static first.** Static work is legal on the strongest footing, risk-free to the account, and agent-safe. Runtime work is reserved for questions static work cannot answer.
- **Two prefixes.** The game prefix is what the game and any anti-cheat observe; tools there change the game's environment.
- **Human-run captures.** Capture requires starting the game and logging in, which the agent never does. Analysis of the saved file is fully agent-safe.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `rea doctor`: "GHIDRA_INSTALL_DIR is not set" | env var missing in the shell or the MCP server's env | set it in both places ([foundations/lab-setup-linux-proton.md](foundations/lab-setup-linux-proton.md#rea-and-ghidra-provider)) |
| Ghidra startup deadline exceeded | UE5-sized binary | raise `REA_GHIDRA_STARTUP_TIMEOUT_MS`, `GHIDRA_HEADLESS_MAXMEM`; build the snapshot from a shell |
| Snapshot will not reopen | different analysis profile (for example changed `REA_GHIDRA_DISABLED_ANALYZERS`) or different binary hash | reopen with the original profile, or rebuild the snapshot |
| Snapshot save fails at close | metadata edits in the session | export Evidence, close without snapshot; names live in `names.json` |
| `inspect_packet_capture` unavailable | `REA_TSHARK_COMMAND` unset | set it to the absolute `tshark` path in REA's environment |
| Custom protocol fields missing in REA | Lua plugin not in the REA process's `$HOME` plugin folder | `tshark -G folders`; symlink the dissector; check with `inspect_packet_fields` |
| FModel lists assets but shows no properties | unversioned properties without `.usmap` | gated dump or static alternatives ([techniques/reflection-dumping.md](techniques/reflection-dumping.md)) |
| Containers list but do not extract | encryption; wrong or changed key | AESDumpster on the current build; re-test key |
| mitmproxy sees no game traffic | wrong mode, kernel below 6.8 for local mode, or process name beyond 16 characters | check mode and process match ([techniques/network-protocol-analysis.md](techniques/network-protocol-analysis.md#tls-and-the-http-backend)) |
| TLS errors with mitmproxy | game in Steam Runtime container not importing host CAs, bundled CA file, or pinning | test `PRESSURE_VESSEL_IMPORT_CA_CERTS=1`; if the client uses its own bundle or pins, stop on live-service clients |
| Frida "Unable to locate the libc" | host Frida attaching to a Wine process | unsupported; see [techniques/dynamic-analysis-under-proton.md](techniques/dynamic-analysis-under-proton.md#frida-under-wine-unsupported) |
| UE4SS does not load under Proton | DLL override not applied, or Proton regression | `WINEDLLOVERRIDES="dwmapi.dll=n,b" %command%`; try another Proton version in the lab prefix |
| No `++UE5+Release` string | studio engine branch | narrow version from other evidence ([recipes/fingerprint-a-ue5-build.md](recipes/fingerprint-a-ue5-build.md)) |

## Where To Go Next

- [AGENTS.md](AGENTS.md) — the fast path.
- [decisions/tooling.md](decisions/tooling.md) — current tool versions.
- [CONTRIBUTING.md](CONTRIBUTING.md) — changing the handbook.
