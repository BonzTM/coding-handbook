# Recipe: Reconstruct A Preservation Server

Use this to stand up an independent server that a client you own can connect to, for a UE5 MMO whose official servers are discontinued. Governed by [../operations/game-preservation.md](../operations/game-preservation.md).

> **Warning:** Confirm first, in `scope.md`, that the servers are discontinued (how established, for how long), that you own the client, your status (individual vs. eligible institution), and the legal basis per activity. The local-gameplay limit in [../operations/game-preservation.md](../operations/game-preservation.md#read-the-legal-frame-precisely) means re-serving to others is a cooperation-dependent question; a personal lab is the default scope here.

## Files To Touch

- `scope.md` (preservation basis; updated before any step)
- `protocol/messages/*.md`, `protocol/dissectors/*.lua` (inputs; read)
- `server/` (new: the re-implementation source and its own repo hygiene)
- `server/coverage.md` (new: message-by-message reconstruction status)
- `findings.md` (reconstructed server-rule inferences)

## Steps

1. Confirm the protocol surface is documented enough to build against: login, character select, world/zone entry, and the message loop for one in-world action. If not, finish [write-a-wireshark-dissector.md](write-a-wireshark-dissector.md) and the specs first. Old live-era captures are the primary source; treat them as irreplaceable.
2. Inventory the handshake. List, in `server/coverage.md`, every message the client sends before it is in-world, and every server message it waits for, from the specs. Mark each `documented`, `inferred`, or `unknown`.
3. Build the server as independent software that speaks the documented protocol. It reads `protocol/` specs; it contains no publisher server code. Start with the connection and login handshake, then character select, then world entry — the order the client drives.
4. Reconstruct unknown server rules from the client, not guesswork: the reader function for a message ([../techniques/static-binary-analysis.md](../techniques/static-binary-analysis.md#writers-and-readers)) and the cooked Blueprints/DataTables ([../techniques/blueprint-analysis.md](../techniques/blueprint-analysis.md)) constrain what a valid reply looks like. Record each reconstructed rule as a `findings.md` inference citing the client Evidence.
5. Handle any startup/anti-cheat handshake as protocol, not as tampering: the server answers (or stands in for) what the client expects at connect time, reconstructed from recordings and the client's reader. If the client refuses to run without an active local protection initializing, stop and read [../operations/game-preservation.md](../operations/game-preservation.md#anti-cheat-at-end-of-life) — the next step is a rights-holder request, not a bypass.
6. Point your own client at the server through configuration, DNS, or a local resolver you control — not a client-code patch ([../operations/game-preservation.md](../operations/game-preservation.md#server-re-implementation-is-the-strategy)). Record the method.
7. Validate against recorded traffic: for each implemented message, the server's bytes match the shape the live-era capture showed, and the dissector parses them without malformed errors. Update `server/coverage.md`.
8. Keep the durable products under version control: `server/` source, `protocol/`, `findings.md`, and exported Evidence bundles. Game bytes stay in the personal archive outside git.

## Invariants To Preserve

- The server contains only your code and the documented protocol; no publisher server code.
- Client redirection is configuration/DNS/local-resolver, or the exception is recorded with its basis.
- No step defeats, disables, or evades an active anti-cheat or integrity check.
- Reconstructed server rules are inferences citing client Evidence, never claimed as the original implementation.
- No game binaries, assets, keys, or captures enter git.

## Proof

- `server/coverage.md` lists every pre-in-world message with a status and, for implemented ones, the capture Evidence the output was validated against.
- A client you own, redirected by the recorded method, completes login and world entry against the server.
- `findings.md` reconstructed-rule rows each cite client Evidence IDs.
- `git ls-files` shows source and specs only; no game bytes.
