# Runtime Session Gate Checklist

Gate for any action that runs, attaches to, injects into, debugs, or intercepts the game client or its traffic. A human walks it before the session; any unchecked box means the session does not run. The rules live in [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md) and [../techniques/dynamic-analysis-under-proton.md](../techniques/dynamic-analysis-under-proton.md).

## Authorization

- [ ] Does `scope.md` list this activity under Allowed Activities, with the purpose stated?
- [ ] Was the game's ToS re-read for this build, with the relevant clauses quoted in `scope.md`?
- [ ] Is the target one of: a client with no anti-cheat, an offline mode, a private or emulated server, or a target covered by written authorization?
- [ ] Does `builds/<build-id>/fingerprint.md` record the anti-cheat check for this exact build?

## Account And Server Exposure

- [ ] Is the session offline, against a private/emulated server, or on a throwaway account whose ban risk the owner accepted in writing in `scope.md`?
- [ ] Is it certain no crafted, modified, or replayed traffic reaches a live server?
- [ ] Is it certain the session gives no gameplay advantage over other players on a live server?

## The Action

- [ ] Is the exact action written down: tool, process, script or breakpoint list, duration, expected output file?
- [ ] Is the action passive (read, log, capture) rather than state-changing (write memory, patch code, send packets)?
- [ ] Will the human, not the agent, start the game and enter any credentials?
- [ ] Is the agent's harness configured to ask before every runtime tool call ([../templates/re-workspace-agents.md](../templates/re-workspace-agents.md))?

## Environment

- [ ] Does the session run in a dedicated lab prefix or snapshot, not the daily-play prefix ([../foundations/lab-setup-linux-proton.md](../foundations/lab-setup-linux-proton.md))?
- [ ] Is a btrfs snapshot (or prefix copy) taken so the prefix can be rolled back?
- [ ] Are captures written under `captures/`, which git ignores?

## Proof

- [ ] The written action plan and this checklist's answers are committed under `briefs/` before the session.
- [ ] After the session, every output file is listed in `findings.md` with its Evidence ID and the type `live-runtime` or `recorded-runtime`.
- [ ] `git status` shows no capture, key, or binary staged.
