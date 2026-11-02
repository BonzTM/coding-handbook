# Template: Engagement Scope (`scope.md`)

Copy to the workspace root as `scope.md` and fill it before any analysis. The rules behind each field live in [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md). An unfilled field means the related activity is out of scope.

```markdown
# Scope — <game> reverse-engineering engagement

Owner: <name>
Date: <yyyy-mm-dd>
Review by: <yyyy-mm-dd>   # re-read after every major patch or ToS change

## Purpose
<One of: interoperability | preservation (servers discontinued) | modding of
single-player/offline content | authorized security research | personal education.
State the concrete goal, e.g. "document the item DataTable schema for a fan wiki".>

## Target
- Game: <name>, store app id <id>
- Build(s): <build-id list>
- Anti-cheat present: <none | EasyAntiCheat | BattlEye | other: ...>  (from fingerprint)
- Terms reviewed: <ToS/EULA URL>, retrieved <yyyy-mm-dd>
- Relevant ToS clauses: <quote the reverse-engineering, modding, and third-party
  software clauses verbatim>

## Authorization
- Bug bounty / VDP: <program URL and in-scope assets, or "none">
- Written permission from the publisher: <reference, or "none">
- Jurisdiction notes: <e.g. US 17 USC 1201(f) interoperability; EU 2009/24/EC Art. 6>

## Allowed Activities
- [ ] Static analysis of locally installed files
- [ ] Asset extraction for personal analysis (no redistribution)
- [ ] Passive capture of my own client's traffic
- [ ] TLS interception of my own client's HTTP traffic
- [ ] Runtime instrumentation (requires: no anti-cheat, OR offline/private server,
      OR written authorization) — see checklists/runtime-session-gate.md
- [ ] Private/emulated server for interoperability or preservation

## Never In Scope
- Anti-cheat, DRM, or integrity-check circumvention
- Automation, botting, or any advantage over other players on live servers
- Sending crafted or replayed traffic to live servers
- Accessing other players' accounts or data
- Redistributing game binaries, assets, keys, or decrypted content

## Accounts
- Account used for runtime work: <throwaway | none>; owner accepts ban risk: <yes/no>
- Credentials are entered by the human only; never stored in the workspace.

## Data Handling
- Captures may contain other players' names, chat, and IDs. Retention: <days>.
- Redaction before sharing: <rule>.

## Disclosure
- Security issues found go to: <security contact / VDP>, not public.
```
