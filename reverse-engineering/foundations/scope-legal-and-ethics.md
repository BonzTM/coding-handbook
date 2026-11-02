# Scope, Legal, And Ethics

What this handbook covers, what it never covers, and the legal frame an engagement must be checked against before analysis starts. This is engineering guidance, not legal advice.

## Default Approach

Decide scope before tooling. Every engagement fills [../templates/engagement-scope.md](../templates/engagement-scope.md) and passes [../checklists/engagement-intake.md](../checklists/engagement-intake.md) first. The scope file is what the agent reads to decide whether an action is allowed.

### Allowed Purposes

This handbook supports reverse engineering of a game client you lawfully own, for:

- **Interoperability** — understanding file formats and protocols so independent software can exchange information with the game (fan tools, data viewers, accessibility aids, launchers).
- **Preservation** — keeping a game playable after its servers are shut down.
- **Modding** — of single-player, offline, or self-hosted content where the game's terms allow it.
- **Authorized security research** — inside a published bug bounty or vulnerability disclosure program, or with written permission, with findings reported privately.
- **Education** — learning how a shipped UE5 game is built, using static analysis of files you own.

### Never In Scope

These are excluded regardless of purpose, and an agent following this handbook refuses them:

- Circumventing, disabling, or evading anti-cheat, DRM, integrity checks, or pak signing on a client that connects to live services.
- Developing cheats, bots, or automation that gives an advantage over other players or affects a live service.
- Sending crafted, modified, or replayed traffic to live servers, or testing server validation by abuse.
- Accessing other players' accounts or data, or collecting other players' personal data.
- Redistributing game binaries, assets, encryption keys, decrypted content, or mapping files derived from them.

REA's own disclaimer draws the same line. It states the project does not support "circumventing DRM, license checks, anti-cheat systems or other technical protection measures" or "developing game cheats or tampering with online services" (`github.com/morluto/rea/blob/main/DISCLAIMER.md`).

### Anti-Cheat Decides Runtime Scope

Anti-cheat on Proton is an explicit developer opt-in. Epic's EAC documentation says Linux support requires the developer to "activate a client module for the Linux platform" (`dev.epicgames.com/docs/epic-online-services/trust-and-safety/anti-cheat-interfaces/anti-cheat-interfaces#linux-wine-and-proton-support`). Valve's Proton partner page says of BattlEye that "each title requires a manual configuration step" (`partner.steamgames.com/doc/steamdeck/proton`).

Consequences for the analyst:

- If the fingerprint shows anti-cheat, runtime instrumentation of the live client is out of scope: no injection, no debugger attach, no Frida, no DLL proxies. Static analysis of files on disk and passive capture of your own traffic remain in scope where the ToS allows them.
- Do not analyze how the anti-cheat works or how to avoid it. That analysis is circumvention work regardless of intent.
- Game terms, not the anti-cheat vendor's, usually govern what is a violation. Read the game's ToS/EULA for clauses on reverse engineering, modding, and third-party software, and quote them in `scope.md`.

### The Legal Frame (United States And European Union)

Know the anchors; get advice from a lawyer for anything beyond personal research.

- **17 U.S.C. § 1201(f)** permits circumvention by a person who lawfully obtained a program "for the sole purpose of identifying and analyzing those elements" needed for interoperability of an independently created program, and defines interoperability as the "ability of computer programs to exchange information" (`law.cornell.edu/uscode/text/17/1201`). Its purpose is limited to interoperability.
- **EU Directive 2009/24/EC, Article 6** permits decompilation when "indispensable to obtain the information necessary to achieve the interoperability", with conditions including that it is "confined to the parts of the original program which are necessary" (`legislation.gov.uk/eudr/2009/24/article/6`, the UK-retained copy of the directive text).
- **37 CFR 201.40(b)(19)** (2024 rulemaking) exempts circumvention to preserve games whose owner "has ceased to provide access to an external computer server", where support was discontinued for at least six months. It covers "local gameplay", defined as not through "an online service or facility" (`ecfr.gov/current/title-37/section-201.40`). It does not cover a live MMO.
- **Case law** — *Sega v. Accolade*, 977 F.2d 1510 (9th Cir. 1992), held disassembly for a legitimate reason with no other means of access to be "a fair use of the copyrighted work, as a matter of law"; *Sony v. Connectix*, 203 F.3d 596 (9th Cir. 2000), held intermediate copies made "during the course of its reverse engineering" to be protected fair use. Both are US Ninth Circuit decisions about copyright, not about breach of contract.

The practical reading: static analysis for interoperability is on the strongest footing; preservation exemptions start when servers are gone; nothing above authorizes interfering with a live online service. When the servers are confirmed gone, the preservation path — server re-implementation and archival extraction — is detailed in [../operations/game-preservation.md](../operations/game-preservation.md), including the local-gameplay limit that keeps online-service restoration a grayer, cooperation-dependent question.

### Data You Will Touch

- Captures contain other players' names, chat, positions, and identifiers. Keep them local, set a retention period in `scope.md`, and redact before sharing anything.
- Session tokens and account identifiers appear in captures and logs. Treat them as secrets: never commit, never paste into an agent prompt, never put in a finding.
- Extracted assets and keys are the publisher's property. Findings describe structure; they do not republish content.

### Disclosure

A security issue found on the way — a server that trusts a client-supplied value, an exposed credential, a data leak — goes to the publisher's security contact or VDP privately. It is not tested further against the live service, and it is not published until the publisher has had time to fix it.

## Common Mistakes And Forbidden Patterns

- Starting analysis before `scope.md` exists.
- Treating "I own the game" as authorization for runtime instrumentation of an anti-cheat-protected client.
- Reading the 2024 preservation exemption as covering a live MMO.
- Researching anti-cheat internals "to understand what to avoid".
- Publishing extracted assets, keys, or `.usmap` files derived from the game.
- Keeping captures with other players' data indefinitely or sharing them unredacted.
- Probing a suspected server vulnerability on the live service instead of reporting it.

## Verification And Proof

- `scope.md` exists, is committed, and predates the first Evidence bundle in `git log`.
- `scope.md` quotes the ToS clauses and records the anti-cheat finding from the fingerprint.
- No runtime session on an anti-cheat build exists unless it ran offline, on a private server, or under written authorization recorded in `scope.md`.
- `git ls-files` shows no captures, keys, binaries, or extracted assets.

## Where To Go Next

- [../templates/engagement-scope.md](../templates/engagement-scope.md) — the scope file.
- [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) — the gate for any runtime action.
- [agentic-re-workflow.md](agentic-re-workflow.md#humans-gate-runtime-actions) — how the harness enforces the gate on agents.
