# Scope, Legal, And Ethics

What this handbook covers, what it never covers, and the legal frame an engagement must be checked against before analysis starts. It applies to any target — native binaries, managed assemblies, apps, firmware, a game client. This is engineering guidance, not legal advice.

## Default Approach

Decide scope before tooling. Every engagement fills [../templates/engagement-scope.md](../templates/engagement-scope.md) and passes [../checklists/engagement-intake.md](../checklists/engagement-intake.md) first. The scope file is what the agent reads to decide whether an action is allowed. Domain-specific rules layer on top of this doc — for games, the anti-cheat and game-terms rules in [../domains/games/anti-cheat-and-scope.md](../domains/games/anti-cheat-and-scope.md).

### Allowed Purposes

This handbook supports reverse engineering of software you lawfully own or are authorized to examine, for:

- **Interoperability** — understanding file formats and protocols so independent software can exchange information with the target (data viewers, converters, accessibility aids, clients, launchers).
- **Preservation** — keeping software usable after the vendor's service or distribution ends.
- **Modding and extension** — of offline or self-hosted content where the target's terms allow it.
- **Authorized security research** — inside a published bug bounty or vulnerability disclosure program, or with written permission, with findings reported privately.
- **Education** — learning how a shipped program is built, using static analysis of files you own.

### Never In Scope

These are excluded regardless of purpose, and an agent following this handbook refuses them:

- Circumventing, disabling, evading, or forging a technical protection measure — anti-cheat, DRM, license checks, integrity or code-signing checks, signed containers — on software that connects to or depends on a live service.
- Developing cheats, bots, or automation that gives an advantage over other users or affects a live service.
- Sending crafted, modified, or replayed traffic to live servers, or testing server validation by abuse.
- Accessing other people's accounts or data, or collecting their personal data.
- Redistributing the target's binaries, assets, encryption keys, decrypted content, or artifacts derived from them (mapping files, dumps).

REA's own disclaimer draws the same line. It states the project does not support "circumventing DRM, license checks, anti-cheat systems or other technical protection measures" or "developing game cheats or tampering with online services" (`github.com/morluto/rea/blob/main/DISCLAIMER.md`).

### The Legal Frame (United States And European Union)

Know the anchors; get advice from a lawyer for anything beyond personal research.

- **17 U.S.C. § 1201(f)** permits circumvention by a person who lawfully obtained a program "for the sole purpose of identifying and analyzing those elements" needed for interoperability of an independently created program, and defines interoperability as the "ability of computer programs to exchange information" (`law.cornell.edu/uscode/text/17/1201`). Its purpose is limited to interoperability, and it reaches the *act* of circumvention, not trafficking in a tool or method (§1201(a)(2), (b)).
- **EU Directive 2009/24/EC, Article 6** permits decompilation when "indispensable to obtain the information necessary to achieve the interoperability", with conditions including that it is "confined to the parts of the original program which are necessary" (`legislation.gov.uk/eudr/2009/24/article/6`, the UK-retained copy of the directive text).
- **37 CFR 201.40(b)(19)** (2024 rulemaking) exempts circumvention to preserve games whose owner "has ceased to provide access to an external computer server", where support was discontinued for at least six months. It covers "local gameplay", defined as not through "an online service or facility" (`ecfr.gov/current/title-37/section-201.40`). It does not cover a live online service.
- **Case law** — *Sega v. Accolade*, 977 F.2d 1510 (9th Cir. 1992), held disassembly for a legitimate reason with no other means of access to be "a fair use of the copyrighted work, as a matter of law"; *Sony v. Connectix*, 203 F.3d 596 (9th Cir. 2000), held intermediate copies made "during the course of its reverse engineering" to be protected fair use. Both are US Ninth Circuit decisions about copyright, not about breach of contract.

The practical reading: static analysis for interoperability is on the strongest footing; preservation exemptions start when a service is gone; nothing above authorizes interfering with a live online service, and no exemption makes a circumvention *tool* lawful to build or share. The game-preservation application of this frame — server re-implementation, archival extraction, and the limits around anti-cheat — is in [../domains/games/game-preservation.md](../domains/games/game-preservation.md).

### Data You Will Touch

- Captures and dumps can contain other people's names, messages, and identifiers. Keep them local, set a retention period in `scope.md`, and redact before sharing anything.
- Session tokens, keys, and account identifiers appear in captures and logs. Treat them as secrets: never commit, never paste into an agent prompt, never put in a finding.
- Extracted assets and keys are the owner's property. Findings describe structure; they do not republish content.

### Disclosure

A security issue found on the way — a server that trusts a client-supplied value, an exposed credential, a data leak — goes to the vendor's security contact or VDP privately. It is not tested further against the live service, and it is not published until the vendor has had time to fix it.

## Common Mistakes And Forbidden Patterns

- Starting analysis before `scope.md` exists.
- Treating "I own it" as authorization to circumvent a protection measure, or to instrument a protected client connected to a live service.
- Reading the 2024 preservation exemption as covering a live online service, or as making a circumvention tool shareable.
- Researching a protection measure's internals "to understand what to avoid".
- Publishing extracted assets, keys, or derived mapping files.
- Keeping captures with other people's data indefinitely or sharing them unredacted.
- Probing a suspected server vulnerability on the live service instead of reporting it.

## Verification And Proof

- `scope.md` exists, is committed, and predates the first Evidence bundle in `git log`.
- `scope.md` records the target's terms and any protection finding from the fingerprint, and the per-activity legal basis.
- No runtime session on a protected build exists unless it ran offline, on a private/test endpoint, or under written authorization recorded in `scope.md`.
- `git ls-files` shows no captures, keys, binaries, or extracted assets.

## Where To Go Next

- [../templates/engagement-scope.md](../templates/engagement-scope.md) — the scope file.
- [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) — the gate for any runtime action.
- [agentic-re-workflow.md](agentic-re-workflow.md#humans-gate-runtime-actions) — how the harness enforces the gate on agents.
- [../domains/games/anti-cheat-and-scope.md](../domains/games/anti-cheat-and-scope.md) — game-specific scope: anti-cheat and game terms.
