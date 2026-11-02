# Game Preservation

How to preserve a UE5 online game whose official servers are gone: the legal frame that applies when servers have been discontinued, server re-implementation as the preservation strategy, archival extraction of a client you own, and the honest limits around protections this handbook does not circumvent.

## Default Approach

Preservation is a first-class purpose of this handbook, but it is still bounded by the same scope and legal rules as every other engagement ([../../foundations/scope-legal-and-ethics.md](../../foundations/scope-legal-and-ethics.md)). Preservation changes which activities have a legal basis; it does not turn off the boundaries. The work is overwhelmingly interoperability and archival: document the protocol, re-implement the server, read your own client's content, and keep the result playable. None of that requires defeating an active protection on the client.

### When This Doc Applies

Use it only when all of the following hold and `scope.md` records them:

- The game's official servers are shut down, not merely down for maintenance.
- The discontinuation has lasted and is expected to continue — the 2024 preservation rulemaking uses "at least six months", and requires that support has not "since been restored" (`ecfr.gov/current/title-37/section-201.40`).
- You lawfully own a copy of the client.
- The purpose is preservation, study, or interoperability, not competitive advantage or redistribution.

If the servers are live, this is not preservation; it is a live-service engagement, and the live-service boundaries in [../../foundations/scope-legal-and-ethics.md](../../foundations/scope-legal-and-ethics.md#never-in-scope) apply unchanged.

### Read The Legal Frame Precisely

The preservation exemption is narrower than it first appears, and the narrowness matters most for an MMO:

- **37 CFR 201.40(b)(19)** exempts circumvention to preserve games whose owner "has ceased to provide access to an external computer server", but it covers "local gameplay", which the rule defines as play that is *not* "through an online service or facility" (`ecfr.gov/current/title-37/section-201.40`). Restoring an MMO's online multiplayer is the thing the local-gameplay clause excludes. The exemption cleanly supports keeping a single client runnable and readable; it does not by itself bless standing up a replacement online service.
- **Preservation by eligible libraries, archives, and museums** has its own, broader conditions in the same rulemaking. An individual hobbyist and an eligible institution are not in the same position; say which you are in `scope.md`.
- **17 U.S.C. § 1201(f)** supports the interoperability core — making independent software (your preservation server, your tools) "exchange information" with the client you own (`law.cornell.edu/uscode/text/17/1201`). This is the strongest footing for server re-implementation.
- **EU Directive 2009/24/EC, Article 6** permits decompilation "indispensable to obtain the information necessary to achieve the interoperability", "confined to the parts of the original program which are necessary" (`legislation.gov.uk/eudr/2009/24/article/6`).

The practical reading: archival extraction and server re-implementation for a client you own sit on interoperability and preservation footing; restoring an *online service* for others is a grayer area that depends on who you are, your jurisdiction, and ideally the rights holder's cooperation. Record the specific basis you are relying on, and get legal advice before anything leaves a personal lab. This handbook is engineering guidance, not legal advice.

### Server Re-Implementation Is The Strategy

For a dead MMO, almost all preservation value comes from rebuilding the server from the protocol you documented, not from altering the client. This is the canonical, interoperability-grounded preservation path:

1. **Document the protocol from recordings you already have.** Preservation projects lean on captures taken while the game was live. The classification and message-spec work in [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md) and [mmo-client-server-model.md](mmo-client-server-model.md) is the input. Old captures are often the only surviving record of server→client messages, so treat any you have as irreplaceable primary evidence.
2. **Re-implement the server as independent software.** Build a service that speaks the documented protocol: login, character select, world, and zone servers as the game used them. It reads your protocol specs and dissectors; it does not contain the publisher's server code. This is new interoperable software under § 1201(f).
3. **Point your own client at it through configuration.** UE clients take the server address from configuration or command line (the NetDriver and online-subsystem settings in cooked config, [unreal-engine-internals.md](unreal-engine-internals.md#cooked-content-and-containers)). Preferring a reachable endpoint you control, for your own client, is interoperability, not tampering. Changing the client's own *code* to do it is a different act — keep redirection in configuration, DNS, or a local resolver you control, not in a client patch, wherever possible.
4. **Fill gaps from the client's own logic, not guesswork.** Where a message's server-side meaning is unknown, the client's reader and the cooked Blueprints/DataTables constrain it ([../../techniques/static-binary-analysis.md](../../techniques/static-binary-analysis.md#writers-and-readers), [blueprint-analysis.md](blueprint-analysis.md)). Record each reconstructed server rule as an inference citing the client evidence it rests on.

Server behavior reconstructed this way is a hypothesis about what the original server did, validated against recorded traffic — never presented as the publisher's actual implementation.

### Archival Extraction Of Content

Reading your own client's cooked content for preservation is interoperability/archival extraction, and the mechanics are already in [asset-extraction.md](asset-extraction.md). The preservation specifics:

- **Container encryption.** A dead game's client still holds the key needed to read its own content; recovering it from your own copy to archive and re-serve that content is the interoperability purpose. Record the § 1201(f)/preservation basis in `scope.md` before recovering a key, per [asset-extraction.md](asset-extraction.md#encryption-keys). The AES-key and `.usmap` steps are unchanged; what changes is that the legal basis is preservation of a discontinued title rather than interoperability with a live one.
- **What to preserve.** Prioritize what is irreplaceable and non-derivable: the protocol specs, cooked config, DataTables and curves (game rules), string tables and localization, map and spawn data. These let the game be understood and re-served.
- **Keep it a personal archive.** Extracted assets and recovered keys stay out of git and are not redistributed ([../../foundations/scope-legal-and-ethics.md](../../foundations/scope-legal-and-ethics.md#data-you-will-touch)). Preservation justifies reading your own copy; it does not justify publishing the publisher's assets or keys.

### Anti-Cheat At End Of Life

This is where preservation and circumvention are easy to confuse, and the line does not depend on your intent — it depends on whether you are implementing a server or altering a protection on the client. Two cases that look alike behave very differently:

- **Game-backend handshake — reconstructable, in scope.** Where the client's startup conversation is with the *game's own* backend (a login, session, or entitlement exchange the game server used to answer), a dead server is the only wall. The preservation fix is to re-implement that server side and answer the exchange, reconstructed like any other message from recordings and the client's reader. You build a server; the client's own code runs and is satisfied; nothing on the client is touched. That is protocol work under [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md), on interoperability footing.
- **Vendor attestation or local integrity — not reconstructable, out of scope.** A third-party anti-cheat (EAC, BattlEye, and the like) usually does not merely talk to the game backend. Its module attests cryptographically against the *vendor's* own servers — which may outlive the game, or be impossible to answer without the vendor's keys — and may also enforce client-side integrity (memory, code-signing, process checks) that no server reply relaxes. You cannot reconstruct a vendor attestation from captures, and forging it, or patching out a local integrity check so the client proceeds, is circumventing the protection, not implementing a server. This handbook does not document that: no step-by-step defeat, disabling, patching-out, forging, or evasion of an active anti-cheat or integrity check. It functions as a circumvention guide whatever the stated purpose, it is excluded by the repo-wide invariants in [../../AGENTS.md](../../AGENTS.md#repo-wide-invariants) and REA's disclaimer, and the preservation exemption's local-gameplay limit does not clearly authorize it.
- **Telling the two apart.** Trace the startup conversation statically ([../../techniques/static-binary-analysis.md](../../techniques/static-binary-analysis.md#find-the-networking-layer)) and in recordings: does the client proceed once a server it talks to returns the right bytes, or does its own code refuse regardless of any reply? The first is a handshake you may answer; the second is a protection you may not touch. Real games mix both, so classify per check, not per game.
- **The legitimate route when a protection — not a dead server — is the wall:** an official end-of-life build from the publisher with the protection removed, explicit permission from the rights holder, or an eligible preservation institution operating under the broader archive exemption. For an abandoned game, a direct request for an EOL build or permission is often the fastest, safest path, and the one a rights holder can stand behind. The next step is that conversation, not a bypass. Why even a permitted act does not let you build or share the bypass is set out in [Limits Of The Exemption](#limits-of-the-exemption-and-the-act-versus-tool-gap).

Kernel-mode anti-cheat is squarely the second case: preservation of a game you own does not justify loading or defeating a kernel driver, and such drivers are a security and stability risk in a lab regardless.

### Limits Of The Exemption, And The Act-Versus-Tool Gap

A protection can gate even *local* play: some anti-cheat initializes at launch and refuses to let the game start, single-player included. When that happens, local-play preservation — the thing the exemption is for — is not achieved just by standing up a local server. The gate still has to open. Three limits decide how far you can take that, and together they explain why this handbook points to avenues rather than handing you a bypass, even when your own act might be permitted.

- **The exemption reaches the act, not the tool.** A triennial exemption can permit *you* to circumvent an access control for a covered purpose, but the Librarian "has no authority to adopt exemptions for the anti-trafficking prohibitions" of 17 U.S.C. § 1201(a)(2) and (b) (`copyright.gov/1201/2024/2024_Section_1201_Registers_Recommendation.pdf`; `law.cornell.edu/uscode/text/17/1201`). Making, sharing, or publishing the circumvention tool or method stays outside any exemption. So even where circumventing a local-play gate for preservation might be a permitted *act*, a document that teaches the bypass, or a patch that performs it, sits in the trafficking zone no exemption reaches. That is a legal reason, on top of this handbook's scope rules, why the mechanism is not here — and why the preserver who is individually permitted to act still "could not obtain such a tool from a third party" lawfully.
- **It covers only non-infringing use, and it did not grow in 2024.** The exemption applies to non-infringing uses, and fair use is not itself a defense to circumvention or trafficking. In the 2024 (ninth) proceeding the Librarian renewed the discontinued-server game exemption but denied the requested expansion, apart from clarifying a single-user limit for preservation institutions, citing market-harm concerns about legacy games (`federalregister.gov/documents/2024/10/28/2024-24563`). Do not plan around a broader exemption than the one on the books.
- **For a pure MMO, "local gameplay" may be nearly empty.** The exemption preserves play that is not "through an online service or facility" (`ecfr.gov/current/title-37/section-201.40`). A game that *is* its online service has little local gameplay to preserve; the part that made it an MMO is the excluded part. The honestly preservable scope is then the artifact, its documentation, extractable content, and whatever offline modes exist — not the live experience.

The resolution follows from those limits. When a protection — not a dead server — gates even local launch, the legal means that close that gate without entering the trafficking zone are to remove the measure at its source or get permission: an official end-of-life build from the publisher with the protection stripped, explicit rights-holder authorization, or, for an eligible institution, the on-premises single-user framing worked out with counsel. Those are the avenues; the bypass itself is neither documented here nor, as a shareable tool, covered by the exemption. Whether a given anti-cheat is even an "access control" to the work, and whether your act is covered, are questions for a lawyer — this is engineering guidance, not legal advice.

### Durability

Preservation is worthless if it rots. Keep, under version control or a durable archive you control:

- the protocol specs and dissectors (text; the durable product);
- the server re-implementation source;
- the finding ledger and exported REA Evidence bundles, so conclusions can be re-checked;
- `builds/<build-id>/` fingerprints and hashes, so the archive states exactly which client it preserves.

Keep the game's own bytes (binaries, assets, keys, captures) in a personal archive outside git, hashed, so their integrity is verifiable without redistributing them.

## Common Mistakes And Forbidden Patterns

- Treating "preservation" as a blanket license; the local-gameplay clause excludes online-service restoration, and the basis must be recorded per activity.
- Assuming that being individually permitted to circumvent a gate lets you build, share, or publish the tool or method that does it; the anti-trafficking provisions are not exemptible.
- Planning around a broader exemption than the one actually on the books, or expecting it to grow.
- Standing up a replacement service for others without checking who you are (individual vs. eligible institution), your jurisdiction, and rights-holder cooperation.
- Documenting or performing forge/defeat/disable/patch-out/evasion of an active anti-cheat or integrity check under a preservation label.
- Treating a vendor attestation or a local integrity check as a "handshake" your server may answer; only a game-backend exchange the client proceeds on is reconstructable.
- Patching the client's code to redirect it when configuration, DNS, or a local resolver would do.
- Redistributing game binaries, assets, or recovered keys as part of "preservation".
- Presenting a reconstructed server rule as the publisher's actual implementation instead of an inference from client evidence.
- Loading or circumventing a kernel-mode anti-cheat driver in the lab.

## Verification And Proof

- `scope.md` states the preservation purpose, that servers are discontinued (with how that was established and for how long), whether you are an individual or an eligible institution, and the legal basis per activity.
- The server re-implementation builds from the protocol specs and cites them; it contains no publisher server code.
- Client redirection is in configuration/DNS, not a client-code patch, or the exception is recorded with its basis.
- Each startup/anti-cheat check is classified as a game-backend handshake or a vendor attestation / local integrity check; only game-backend handshakes are answered in the server, and no ledger row, recipe step, or note describes forging or circumventing an active anti-cheat or integrity check.
- `git ls-files` shows no game binaries, assets, keys, or captures; the durable archive holds them with hashes outside git.

## Where To Go Next

- [../../foundations/scope-legal-and-ethics.md](../../foundations/scope-legal-and-ethics.md) — the boundaries preservation operates within.
- [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md) — the protocol documentation a preservation server is built from.
- [asset-extraction.md](asset-extraction.md) — archival extraction of your own client's content.
- [recipes/reconstruct-a-preservation-server.md](recipes/reconstruct-a-preservation-server.md) — the step-by-step reconstruction procedure.
