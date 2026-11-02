# Template: RE Workspace Layout

Copy this tree into a new git repository for each engagement. Replace every `<placeholder>`. The rationale for each path lives in [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md#the-workspace-is-the-agents-memory).

```text
<game-slug>-re/
  AGENTS.md                 # from re-workspace-agents.md; CLAUDE.md may symlink to it
  scope.md                  # from engagement-scope.md; filled before any analysis
  findings.md               # from finding-ledger.md; the only record of conclusions
  .gitignore                # see below
  .claude/
    settings.json           # permission deny/ask rules for runtime tools
  builds/
    <build-id>/
      fingerprint.md        # engine version, exe hash, pak/utoc list, protections
      hashes.sha256         # sha256sum of every shipped file you analyzed
      names.json            # confirmed {"0xADDR": "Name"} map, re-applied each REA session
  keys/                     # git-ignored; AES keys and mapping files per build
  sdk/                      # git-ignored or LFS; reflection dumps per build
    <build-id>/
  evidence/                 # exported REA Evidence bundles
    <build-id>-<topic>-<yyyy-mm-dd>.json
  snapshots/                # git-ignored; REA/Ghidra analysis snapshots per build
  exports/                  # git-ignored; FModel/CUE4Parse output
  captures/                 # git-ignored; pcapng, mitmproxy flows, HAR
  protocol/
    messages/<message-name>.md   # from protocol-message-spec.md
    dissectors/<proto>.lua
  scripts/                  # your own tooling; never run from inside exports/
  briefs/                   # subagent briefs from agent-brief.md
```

## `.gitignore`

```gitignore
# Target material: never committed
/keys/
/sdk/
/snapshots/
/exports/
/captures/
*.pak
*.utoc
*.ucas
*.exe
*.dll
*.pdb
*.usmap
# Captures can hold session tokens and other players' data
*.pcap
*.pcapng
*.har
*.mitm
```

## Rules

- Game binaries, assets, keys and captures stay out of git. They are the vendor's copyrighted material or contain credentials and other people's data.
- `builds/<build-id>/hashes.sha256` is committed. It proves which bytes every finding refers to.
- `scripts/` holds your code. Downloaded tools and extracted content live in separate directories and are passed to scripts as paths.
