# MMO Client/Server Model

How an Unreal Engine 5 online game splits work between client and server, how to classify a target's architecture before deep analysis, and what a client can and cannot reveal about server behavior.

## Default Approach

Classify the architecture first. It decides which techniques pay off. A game that uses Unreal's built-in replication for everything is analyzed through the reflection system and the NetDriver. A game that runs its MMO backend on a custom protocol is analyzed through its socket layer and serialization library, and Unreal's replication may only carry a thin slice of traffic or none at all.

### The Server Is Authoritative

Unreal's networking model is server-authoritative. Epic's networking overview says the server "holds the one, true, authoritative game state" (`dev.epicgames.com/documentation/en-us/unreal-engine/networking-overview-for-unreal-engine`). For an MMO this has two consequences for the analyst:

- The client contains the **client half** of every system: input, prediction, presentation, request messages, and the parsing of server replies. It does not contain server-side validation, loot tables that never ship, or server-only code paths.
- Client data can **describe** server rules (a DataTable of item stats shipped for UI tooltips) without **being** the server rule. Treat shipped tables as the client's copy, an inference about server behavior, not an observation of it.

A client finding never proves what the server does. Server behavior is only observed through recorded server responses to normal play.

### Unreal's Built-In Replication

When a game uses Unreal's own networking, the moving parts are documented in Epic's networking overview and API reference:

- **NetDriver** — `UNetDriver` (`Engine/NetDriver.h`) owns the connections and the available channel types (`ChannelDefinitions`). The default UDP implementation is `UIpNetDriver`.
- **NetConnection** — one per remote peer; on a client, one connection to the server.
- **Channels** — a control channel, and actor channels. Epic's API reference describes `UActorChannel` as "a channel for exchanging actor and its subobject's properties and RPCs" (`Engine/ActorChannel.h`).
- **Bunches and packets** — channels send bunches, which the engine packs into UDP packets with its own reliability and ordering. The format is bit-packed, not byte-aligned. Networking Insights names each bunch "after the channel it belongs to" (`dev.epicgames.com/documentation/unreal-engine/networking-insights-in-unreal-engine`).
- **Property replication** — `UPROPERTY(Replicated)` fields sent server→client when they change.
- **RPCs** — the networking overview lists three types: "Server (only runs on the server)", "Client (only runs on the Actor's owning client)", and NetMulticast, which runs on every connected machine. Server RPCs are the client→server request surface.
- **Packet encryption** — optional. `FAESGCMHandlerComponent` is documented as "AES256 GCM block encryption component", and `FNetDelegates::OnReceivedNetworkEncryptionToken` is "fired when an encryption key is required by the network layer" (Epic API reference, `AESGCMHandlerComponent.h`, `Engine/NetworkDelegates.h`). An encrypted NetDriver stream is opaque on the wire; analysis moves into the client ([../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md#unreal-netdriver-traffic)).
- **Iris** — Unreal's newer replication system. Epic's Iris introduction page labels it "Experimental" and enables it per project with `bUseIris = true;`, `SetupIrisSupport(Target);`, and `net.Iris.UseIrisReplication=1` (`dev.epicgames.com/documentation/en-us/unreal-engine/introduction-to-iris-in-unreal-engine`). Check whether the target ships Iris before assuming the classic replication path ([../../techniques/static-binary-analysis.md](../../techniques/static-binary-analysis.md#find-the-networking-layer)).

Reflection makes this surface enumerable. Every RPC is a `UFunction` with network flags, so a reflection dump lists the client→server request surface by name ([reflection-dumping.md](reflection-dumping.md)).

### Custom MMO Backends

Many MMOs keep Unreal for rendering, input, and local simulation and run the persistent world on a separate backend. Typical signals:

- Long-lived TCP or WebSocket connections to hosts that are not the UDP game-server host.
- An embedded serialization library: Protocol Buffers, FlatBuffers, MessagePack, Cap'n Proto, or a hand-written binary format. Protocol Buffers leaves serialized descriptors and `.proto` file names in the binary that tools can recover ([../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md#identify-the-serialization-format)).
- HTTPS REST or gRPC for login, character select, store, social, and matchmaking.
- Third-party online SDKs (Epic Online Services, Steamworks, PlayFab, AWS GameLift, custom). Their DLLs and strings fingerprint them.

Hybrid designs are common: HTTPS for login and meta-game, a custom TCP protocol for the world server, and Unreal UDP replication inside an instance or zone.

### Classify Before You Dig

Run this classification as the first network question and record each answer as a ledger finding:

| Question | Evidence | Tool |
|---|---|---|
| Which hosts and ports does the client talk to during login, idle in world, and zone change? | Recorded capture of one scripted session | `inspect_pcap_capture` summary (endpoints, DNS, SNI, protocol distribution) |
| Is UDP traffic Unreal NetDriver traffic? | NetDriver strings and config in the binary and `DefaultEngine.ini` in the containers; packet-size patterns | `search_strings`, asset export of config |
| Is there a serialization library? | Library strings, embedded descriptors, `.proto` names | `search_strings`, descriptor extraction tools |
| What runs over HTTPS? | Intercepted flows of the login and store sequences | `inspect_web_network_capture` + `api-surface` view |
| Which online SDKs ship? | DLL names in `Binaries/Win64`, import tables, strings | file listing, `inspect_pe_resources`, `search_strings` |

### What The Client Can Tell You

| Question | Client evidence can answer | Client evidence cannot answer |
|---|---|---|
| Message formats | Yes: the client writes requests and parses responses | Fields the client ignores in responses |
| Request surface | Yes: Server RPC names, opcodes, REST routes the client calls | Server-only admin routes |
| Game data | Client-shipped DataTables, strings, curves | Server-only tables, drop rates not shipped |
| Validation | Client-side checks only | Server validation, except as seen in recorded rejections |
| Anti-abuse | Out of scope ([../../foundations/scope-legal-and-ethics.md](../../foundations/scope-legal-and-ethics.md)) | — |

## Common Mistakes And Forbidden Patterns

- Assuming all traffic is Unreal replication because the engine is Unreal.
- Reporting a shipped DataTable value as "the server's drop rate" or "the server's formula".
- Probing server validation by sending crafted requests to a live server. Server behavior is learned from recorded responses to normal play, or on a private server you run.
- Starting packet-level decoding before the endpoint and transport classification is in the ledger.

## Verification And Proof

- `findings.md` contains the classification table's five answers, each with an Evidence ID from a capture or the binary.
- Every finding about server behavior is typed `inference` and names the recorded responses it rests on.
- The capture used for classification is listed by file name and build ID.

## Where To Go Next

- [unreal-engine-internals.md](unreal-engine-internals.md) — the reflection and packaging internals the classification relies on.
- [../../techniques/network-protocol-analysis.md](../../techniques/network-protocol-analysis.md) — decoding the transports once classified.
- [../../recipes/capture-and-classify-traffic.md](../../recipes/capture-and-classify-traffic.md) — the capture-and-classify procedure.
