# Game Preservation

How to preserve a UE5 online game whose official servers are gone: the legal frame that applies when servers have been discontinued, server re-implementation as the preservation strategy, archival extraction of a client you own, and the honest limits around protections this handbook does not circumvent.

## Default Approach

Preservation is a first-class purpose of this handbook, but it is still bounded by the same scope and legal rules as every other engagement ([../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md)). Preservation changes which activities have a legal basis; it does not turn off the boundaries. The work is overwhelmingly interoperability and archival: document the protocol, re-implement the server, read your own client's content, and keep the result playable. None of that requires defeating an active protection on the client.

### When This Doc Applies

Use it only when all of the following hold and `scope.md` records them:

- The game's official servers are shut down, not merely down for maintenance.
- The discontinuation has lasted and is expected to continue — the 2024 preservation rulemaking uses "at least six months", and requires that support has not "since been restored" (`ecfr.gov/current/title-37/section-201.40`).
- You lawfully own a copy of the client.
- The purpose is preservation, study, or interoperability, not competitive advantage or redistribution.

If the servers are live, this is not preservation; it is a live-service engagement, and the live-service boundaries in [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md#never-in-scope) apply unchanged.

### Read The Legal Frame Precisely

The preservation exemption is narrower than it first appears, and the narrowness matters most for an MMO:

- **37 CFR 201.40(b)(19)** exempts circumvention to preserve games whose owner "has ceased to provide access to an external computer server", but it covers "local gameplay", which the rule defines as play that is *not* "through an online service or facility" (`ecfr.gov/current/title-37/section-201.40`). Restoring an MMO's online multiplayer is the thing the local-gameplay clause excludes. The exemption cleanly supports keeping a single client runnable and readable; it does not by itself bless standing up a replacement online service.
- **Preservation by eligible libraries, archives, and museums** has its own, broader conditions in the same rulemaking. An individual hobbyist and an eligible institution are not in the same position; say which you are in `scope.md`.
- **17 U.S.C. § 1201(f)** supports the interoperability core — making independent software (your preservation server, your tools) "exchange information" with the client you own (`law.cornell.edu/uscode/text/17/1201`). This is the strongest footing for server re-implementation.
- **EU Directive 2009/24/EC, Article 6** permits decompilation "indispensable to obtain the information necessary to achieve the interoperability", "confined to the parts of the original program which are necessary" (`legislation.gov.uk/eudr/2009/24/article/6`).

The practical reading: archival extraction and server re-implementation for a client you own sit on interoperability and preservation footing; restoring an *online service* for others is a grayer area that depends on who you are, your jurisdiction, and ideally the rights holder's cooperation. Record the specific basis you are relying on, and get legal advice before anything leaves a personal lab. This handbook is engineering guidance, not legal advice.

### Server Re-Implementation Is The Strategy

For a dead MMO, almost all preservation value comes from rebuilding the server from the protocol you documented, not from altering the client. This is the canonical, interoperability-grounded preservation path:

1. **Document the protocol from recordings you already have.** Preservation projects lean on captures taken while the game was live. The classification and message-spec work in [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md) and [../foundations/mmo-client-server-model.md](../foundations/mmo-client-server-model.md) is the input. Old captures are often the only surviving record of server→client messages, so treat any you have as irreplaceable primary evidence.
2. **Re-implement the server as independent software.** Build a service that speaks the documented protocol: login, character select, world, and zone servers as the game used them. It reads your protocol specs and dissectors; it does not contain the publisher's server code. This is new interoperable software under § 1201(f).
3. **Point your own client at it through configuration.** UE clients take the server address from configuration or command line (the NetDriver and online-subsystem settings in cooked config, [../foundations/unreal-engine-internals.md](../foundations/unreal-engine-internals.md#cooked-content-and-containers)). Preferring a reachable endpoint you control, for your own client, is interoperability, not tampering. Changing the client's own *code* to do it is a different act — keep redirection in configuration, DNS, or a local resolver you control, not in a client patch, wherever possible.
4. **Fill gaps from the client's own logic, not guesswork.** Where a message's server-side meaning is unknown, the client's reader and the cooked Blueprints/DataTables constrain it ([../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md#writers-and-readers), [../techniques/blueprint-analysis.md](../techniques/blueprint-analysis.md)). Record each reconstructed server rule as an inference citing the client evidence it rests on.

Server behavior reconstructed this way is a hypothesis about what the original server did, validated against recorded traffic — never presented as the publisher's actual implementation.

### Archival Extraction Of Content

Reading your own client's cooked content for preservation is interoperability/archival extraction, and the mechanics are already in [../techniques/asset-extraction.md](../techniques/asset-extraction.md). The preservation specifics:

- **Container encryption.** A dead game's client still holds the key needed to read its own content; recovering it from your own copy to archive and re-serve that content is the interoperability purpose. Record the § 1201(f)/preservation basis in `scope.md` before recovering a key, per [../techniques/asset-extraction.md](../techniques/asset-extraction.md#encryption-keys). The AES-key and `.usmap` steps are unchanged; what changes is that the legal basis is preservation of a discontinued title rather than interoperability with a live one.
- **What to preserve.** Prioritize what is irreplaceable and non-derivable: the protocol specs, cooked config, DataTables and curves (game rules), string tables and localization, map and spawn data. These let the game be understood and re-served.
- **Keep it a personal archive.** Extracted assets and recovered keys stay out of git and are not redistributed ([../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md#data-you-will-touch)). Preservation justifies reading your own copy; it does not justify publishing the publisher's assets or keys.

### Anti-Cheat At End Of Life

This is where preservation and circumvention are easy to confuse. The honest picture:

- **Server-gated anti-cheat is usually self-resolving.** Most online anti-cheat authenticates against the vendor or game backend at startup. When that backend is gone, the anti-cheat often cannot initialize, so it is a symptom of the dead server, not a separate wall. The preservation fix is the re-implemented server handling (or standing in for) the anti-cheat handshake the client expects — reconstructed, like any other message, from recordings and the client's own logic. That is protocol work under [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md), not tampering with the protection.
- **The legitimate routes when the client still refuses to run:** an official end-of-life build from the publisher with the protection removed, explicit permission from the rights holder, or working with an eligible preservation institution operating under the broader archive exemption. For a game a publisher has abandoned, a direct request for an EOL build or permission is often the fastest and safest path, and it is the route preservation communities and the rights holder can both stand behind.
- **What this handbook does not document:** step-by-step defeat, disabling, patching-out, or evasion of an active anti-cheat or integrity check. That content functions as a circumvention guide whatever the stated purpose, it is excluded by the repo-wide invariants in [../AGENTS.md](../AGENTS.md#repo-wide-invariants) and REA's disclaimer, and even the preservation exemption's local-gameplay limit does not clearly authorize it. If the only way forward is through an active protection, the next step is a rights-holder conversation, not a bypass.

Kernel-mode anti-cheat adds a further reason to stop: preservation of a game you own does not justify loading or defeating a kernel driver on a machine, and such drivers are a security and stability risk in a lab regardless.

### Durability

Preservation is worthless if it rots. Keep, under version control or a durable archive you control:

- the protocol specs and dissectors (text; the durable product);
- the server re-implementation source;
- the finding ledger and exported REA Evidence bundles, so conclusions can be re-checked;
- `builds/<build-id>/` fingerprints and hashes, so the archive states exactly which client it preserves.

Keep the game's own bytes (binaries, assets, keys, captures) in a personal archive outside git, hashed, so their integrity is verifiable without redistributing them.

## Common Mistakes And Forbidden Patterns

- Treating "preservation" as a blanket license; the local-gameplay clause excludes online-service restoration, and the basis must be recorded per activity.
- Standing up a replacement service for others without checking who you are (individual vs. eligible institution), your jurisdiction, and rights-holder cooperation.
- Documenting or performing defeat/disable/evasion of an active anti-cheat or integrity check under a preservation label.
- Patching the client's code to redirect it when configuration, DNS, or a local resolver would do.
- Redistributing game binaries, assets, or recovered keys as part of "preservation".
- Presenting a reconstructed server rule as the publisher's actual implementation instead of an inference from client evidence.
- Loading or circumventing a kernel-mode anti-cheat driver in the lab.

## Verification And Proof

- `scope.md` states the preservation purpose, that servers are discontinued (with how that was established and for how long), whether you are an individual or an eligible institution, and the legal basis per activity.
- The server re-implementation builds from the protocol specs and cites them; it contains no publisher server code.
- Client redirection is in configuration/DNS, not a client-code patch, or the exception is recorded with its basis.
- No ledger row, recipe step, or note describes circumventing an active anti-cheat or integrity check.
- `git ls-files` shows no game binaries, assets, keys, or captures; the durable archive holds them with hashes outside git.

## Where To Go Next

- [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md) — the boundaries preservation operates within.
- [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md) — the protocol documentation a preservation server is built from.
- [../techniques/asset-extraction.md](../techniques/asset-extraction.md) — archival extraction of your own client's content.
- [../recipes/reconstruct-a-preservation-server.md](../recipes/reconstruct-a-preservation-server.md) — the step-by-step reconstruction procedure.
