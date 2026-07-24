# Multiplayer

This doc owns the high-level multiplayer API: peer and transport selection, `@rpc` contracts, server-authoritative design, and the replication scaffolding nodes. Signals wire nodes inside one process; RPCs are the only mechanism that crosses the network boundary — the in-process rules live in [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md).

## Default Approach

Godot 4.x. Use the engine's high-level multiplayer API: a `MultiplayerAPI` managed by the SceneTree over a pluggable `MultiplayerPeer`, assigned via `multiplayer.multiplayer_peer` (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`). The server's peer ID is always 1; clients get random positive integer IDs from `multiplayer.get_unique_id()`.

- **Server-authoritative by default.** The server simulates the game; clients send inputs and render results. Client-hosted (listen-server) and pure P2P topologies are architecture decisions that require an ADR, because they change every trust assumption in [Server Authority](#server-authority).
- **One system owns the network lifecycle.** A single session manager creates the peer, connects the lifecycle signals, and exposes typed signals (`player_joined(id: int)`, not raw peer events) to the rest of the game. If it is an autoload, it goes through [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md) — a session system that owns its data is the legitimate autoload case.
- **Handle every lifecycle signal.** `peer_connected` / `peer_disconnected` on all peers; `connected_to_server`, `connection_failed`, and `server_disconnected` on clients (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`). An unhandled `server_disconnected` is a client that silently plays against nobody. Check the `Error` return of `create_server()` / `create_client()` and surface the failure; never assume the port bound.
- **Gate connections.** Set `auth_callback` on `SceneMultiplayer` so peers are verified before they join — with an empty callback "peers will be automatically accepted as soon as they connect" — and set `auth_timeout` above `0.0` so authentication cannot hang forever (`docs.godotengine.org/en/stable/classes/class_scenemultiplayer.html`). Flip `refuse_new_connections` once a session is full or in progress.

## Peer And Transport Selection

| Transport | Default for | Notes |
|---|---|---|
| `ENetMultiplayerPeer` | Desktop and mobile client-server | The default; UDP via "a modified version of ENet which allows for full IPv6 support" |
| `WebSocketMultiplayerPeer` | Web exports | The web platform offers WebSockets and WebRTC only — ENet/UDP is unavailable there |
| `WebRTCMultiplayerPeer` | Peer-to-peer, including web | Requires external signaling; P2P topology needs the ADR above |

Source: `docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`.

- Pick the transport per export target, not per preference: shipping to web forces WebSocket or WebRTC.
- A web-multiplayer target also constrains language choice — C# projects cannot export to the web platform; see [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md).
- Keep the transport behind the session manager so gameplay code never touches a `MultiplayerPeer` type directly; swapping ENet for WebSocket must not touch game logic.

## RPC Contracts

RPCs are declared with the `@rpc` annotation; the default is `@rpc("authority", "call_remote", "reliable", 0)` (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`).

- **Mode:** keep the `"authority"` default for everything the server pushes to clients. `"any_peer"` is reserved for client-to-server input RPCs and marks a trust boundary — every `"any_peer"` function follows the rules in [Server Authority](#server-authority).
- **Transfer mode:** `"reliable"` for discrete state changes (spawn, death, inventory); `"unreliable"` for continuous streams where the next update supersedes the last (positions); `"unreliable_ordered"` only on its own transfer channel — channels are separate packet streams that do not interfere, and mixing variable-size traffic on one unreliable-ordered channel causes packet loss.
- **`"call_local"`** only when the sending peer must also run the effect (a hosting server that is also a player). Remote-only is the default because double-execution bugs on the host are the common failure.
- **Signatures are a wire contract.** "Both RPCs must have the same signature which is evaluated with a checksum" across all RPCs — client and server builds must ship from the same script revision, or peers fail to communicate. Treat any change to an `@rpc` function's signature as a protocol change: version the builds together and note it in the release, per [../operations/ci-and-release.md](../operations/ci-and-release.md).
- Type every RPC parameter, same as signal payloads ([../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md)). Payloads are plain values — never objects; see the `allow_object_decoding` ban below.

## Server Authority

The official docs are the rule: "Treat all client input as untrusted", use "server-authoritative logic for gameplay-critical decisions", and "Validate RPC arguments before applying them to the game state" (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`).

- **Clients send intent, not outcomes.** Never let a client decide player position, combat results, inventory changes, or match outcomes — the client sends the input action ([../foundations/input-handling.md](../foundations/input-handling.md)), the server simulates and replicates the result. Movement simulation and any client-side prediction live in [physics-and-movement.md](physics-and-movement.md).
- **Every `"any_peer"` RPC validates first.** Resolve the caller with `multiplayer.get_remote_sender_id()`, assert the sender owns the entity it is acting on, range-check and type-check every argument, and return without mutating state on any failure. The sender ID is the only trustworthy field in the call; an ID inside the payload is attacker-controlled.
- **Rate-limit frequent actions.** "Add safety checks and rate limits to actions that can be triggered frequently" (same doc). Server-side per-peer cooldowns on fire/interact/chat RPCs; a peer that exceeds them gets throttled or disconnected, not trusted.
- **Authority is per-node and narrow.** The server (peer 1) is the default multiplayer authority; `set_multiplayer_authority()` reassigns it per-node. Granting a client authority is limited to nodes that only represent that client's own input or cosmetics — never a node whose state other players consume as truth. Guard authority-side code with `is_multiplayer_authority()`.
- **Ship the server headless.** Use the dedicated-server export mode, which strips visual resources (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`), and run it with `--headless` — export wiring in [../operations/ci-and-release.md](../operations/ci-and-release.md).

## Replication Scaffolding

Use the built-in replication nodes instead of hand-rolled sync RPCs; hand-rolled replication is an ADR-level exception.

- **`MultiplayerSpawner`** "automatically replicates spawnable nodes from the authority to other multiplayer peers" (`docs.godotengine.org/en/stable/classes/class_multiplayerspawner.html`). Register every networked scene with `add_spawnable_scene()`; use `spawn_function` for data-driven spawns — the callable returns a node it does not `add_child()` itself. Set `spawn_limit`: the default `0` means no limit, which is unacceptable for anything a client action can trigger.
- **`MultiplayerSynchronizer`** "synchronizes properties from the multiplayer authority to the remote peers" via a `SceneReplicationConfig` edited in the inspector (`docs.godotengine.org/en/stable/classes/class_multiplayersynchronizer.html`). Prefer on-change replication with a nonzero `delta_interval` for slow-moving state; reserve every-frame sync for state that genuinely changes every frame. `Object`-type properties (including `Resource`) cannot be synchronized — replicate plain values.
- **Replication flows authority-to-peers only.** A synchronizer never carries client input to the server; input goes through an `"any_peer"` RPC and server validation. A client with authority over a synchronized node is the narrow-authority case above, nothing more.
- **Hidden state uses visibility filters.** `public_visibility`, `set_visibility_for()`, and visibility filter callables control which peers receive a node's sync data. State a player must not know (fogged positions, hidden hands) is filtered at the synchronizer — sending it and hiding it in UI is an information leak, not a design.
- **`allow_object_decoding` stays `false`.** The class reference warning is absolute: "Deserialized objects can contain code which gets executed. Do not use this option if the serialized object comes from untrusted sources to avoid potential security threat such as remote code execution" (`docs.godotengine.org/en/stable/classes/class_scenemultiplayer.html`). Every remote peer is an untrusted source. This is the same code-in-data hazard as loading foreign `.tres` files — see [save-and-load.md](save-and-load.md).

## Common Mistakes And Forbidden Patterns

- An `"any_peer"` RPC that mutates game state without validating `get_remote_sender_id()`, ownership, and every argument.
- Trusting a peer ID or "who am I" field inside an RPC payload instead of the transport-provided sender ID.
- Clients authoritative over position, combat, inventory, or match outcome — or a `MultiplayerSynchronizer` pointed client-to-server as an input channel.
- Enabling `allow_object_decoding`, or passing objects/Resources through RPC payloads.
- `"reliable"` transfer for per-frame position streams, or `"unreliable"` for one-shot state changes a peer must not miss.
- Ignoring the `Error` from `create_server()`/`create_client()`, or leaving `connection_failed`/`server_disconnected` unhandled.
- Shipping client and server builds from different revisions of scripts that declare RPCs — the signature checksum makes the mismatch a runtime communication failure.
- `spawn_limit` left at `0` on a spawner that client actions can trigger.
- Replicating hidden state to all peers and masking it in UI instead of using synchronizer visibility.
- Empty `auth_callback` plus no `auth_timeout` on a public server — every socket that connects becomes a peer.
- Gameplay code constructing or holding `MultiplayerPeer` objects outside the session manager.

## Verification And Proof

- two-instance smoke test: one `--headless` dedicated-server instance plus one client from the CLI; join, play, disconnect, and confirm the client handles `server_disconnected`
- a negative test per `"any_peer"` RPC: a forged sender ID, an out-of-range argument, and an action on an entity the sender does not own each leave server state unchanged — framework wiring in [../quality/testing.md](../quality/testing.md)
- a rate-limit test: an RPC spammed past its cooldown is throttled, not applied N times
- `grep -rn "@rpc" --include="*.gd"` — every `"any_peer"` site shows sender validation before mutation, and no site passes object-typed parameters
- `grep -rn "allow_object_decoding" --include="*.gd" --include="*.tscn"` returns nothing
- inspect every `MultiplayerSpawner` in scenes for a nonzero `spawn_limit`, and every `MultiplayerSynchronizer` config for hidden-state properties without a visibility filter
- bandwidth sanity-checked with the profiler under representative peer counts — see [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md)

## Related

- [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) — in-process communication; RPCs replace none of it locally
- [../foundations/input-handling.md](../foundations/input-handling.md) — the action layer that produces the intents clients send
- [physics-and-movement.md](physics-and-movement.md) — server-side simulation and movement rules
- [save-and-load.md](save-and-load.md) — the matching untrusted-data rules for files
- [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md) — language choice constraints for web-multiplayer targets
- [../operations/ci-and-release.md](../operations/ci-and-release.md) — dedicated-server export and lockstep client/server versioning
