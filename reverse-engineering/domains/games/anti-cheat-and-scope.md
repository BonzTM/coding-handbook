# Anti-Cheat And Game-Specific Scope

The game-specific scope rules that layer on top of the general [scope, legal, and ethics](../../foundations/scope-legal-and-ethics.md) doc: how anti-cheat decides what runtime work is possible, and why the game's own terms, not the anti-cheat vendor's, usually govern.

## Default Approach

A game client adds one classification to the general scope decision: does it ship anti-cheat, and is it connected to a live service? That answer gates every runtime technique. Record it in `scope.md` from the fingerprint ([recipes/fingerprint-a-ue5-build.md](recipes/fingerprint-a-ue5-build.md)).

### Anti-Cheat Decides Runtime Scope

Anti-cheat on Proton is an explicit developer opt-in. Epic's EAC documentation says Linux support requires the developer to "activate a client module for the Linux platform" (`dev.epicgames.com/docs/epic-online-services/trust-and-safety/anti-cheat-interfaces/anti-cheat-interfaces#linux-wine-and-proton-support`). Valve's Proton partner page says of BattlEye that "each title requires a manual configuration step" (`partner.steamgames.com/doc/steamdeck/proton`).

Consequences for the analyst:

- If the fingerprint shows anti-cheat, runtime instrumentation of the live client is out of scope: no injection, no debugger attach, no Frida, no DLL proxies. Static analysis of files on disk and passive capture of your own traffic remain in scope where the terms allow them.
- Do not analyze how the anti-cheat works or how to avoid it. That analysis is circumvention work regardless of intent.
- Game terms, not the anti-cheat vendor's, usually govern what is a violation. Read the game's ToS/EULA for clauses on reverse engineering, modding, and third-party software, and quote them in `scope.md`.

### When The Game Is Discontinued

Anti-cheat changes shape once the servers are gone. A server-gated handshake may be answerable by a preservation server; a vendor attestation or local integrity check is not, and forcing it is circumvention the preservation exemption does not clearly reach. The full treatment — the three gates, the act-versus-tool gap, and the legal avenues (publisher end-of-life build, rights-holder permission, institutional framing) — is in [game-preservation.md](game-preservation.md#anti-cheat-at-end-of-life).

## Common Mistakes And Forbidden Patterns

- Treating "I own the game" as authorization for runtime instrumentation of an anti-cheat-protected client connected to live services.
- Researching anti-cheat internals "to understand what to avoid".
- Relying on the anti-cheat vendor's terms instead of the game's own ToS/EULA to decide what is allowed.
- Recording no anti-cheat finding in `scope.md` before the first runtime decision.

## Verification And Proof

- `scope.md` records the anti-cheat finding from the fingerprint, human-reviewed, and quotes the game's relevant ToS clauses.
- No runtime session on an anti-cheat build exists unless it ran offline, on a private/test server, or under written authorization recorded in `scope.md`.

## Where To Go Next

- [../../foundations/scope-legal-and-ethics.md](../../foundations/scope-legal-and-ethics.md) — the general scope and legal frame this layers on.
- [game-preservation.md](game-preservation.md) — anti-cheat when the servers are gone.
- [../../checklists/runtime-session-gate.md](../../checklists/runtime-session-gate.md) — the gate every runtime session passes.
