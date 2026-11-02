# Network Protocol Analysis

How to turn recorded traffic from your own UE5 MMO client into documented message specs: classify transports, intercept the HTTPS backend, identify serialization formats, approach Unreal NetDriver traffic, write Wireshark dissectors that REA can use, and build interoperability tooling without touching live servers.

## Default Approach

Work from recordings. Capture one scripted session per question ([../recipes/capture-and-classify-traffic.md](../recipes/capture-and-classify-traffic.md)), then analyze the saved file as often as needed. Analysis of a saved file is static and agent-safe; capture is a runtime session for a human. Correlate every wire-level claim with a writer or reader found statically ([static-binary-analysis.md](static-binary-analysis.md#writers-and-readers)); a message spec is verified only when both halves agree.

### Classify First

Classification ([../foundations/mmo-client-server-model.md](../foundations/mmo-client-server-model.md#classify-before-you-dig)) decides the path per endpoint:

| Endpoint looks like | Path |
|---|---|
| HTTPS to API hostnames (login, store, social) | [TLS And The HTTP Backend](#tls-and-the-http-backend) |
| Long-lived TCP or WebSocket to world hosts | [Identify The Serialization Format](#identify-the-serialization-format), then a dissector |
| UDP to instance or zone hosts, engine NetDriver strings present | [Unreal NetDriver Traffic](#unreal-netdriver-traffic) |
| UDP/TCP to voice, telemetry, CDN, crash reporting | Record and ignore unless the question needs it |

REA `inspect_pcap_capture` with `detail: "summary"` gives endpoints, DNS, TLS SNI, protocol distribution, and HTTP URIs for a capture, retained for paging with `inspect_analysis_view`. It needs caller-supplied `tshark` and `capinfos`, and "REA never captures traffic" (tool contract).

### TLS And The HTTP Backend

mitmproxy is the default interceptor. Two facts decide whether it works for a Proton game:

1. **How traffic reaches mitmproxy.** mitmproxy's local capture mode "transparently captures traffic from applications running on the same device", selectable by process name or PID (`--mode local:<name>`). On Linux it requires "Linux 6.8 and above", is eBPF-based, needs sudo, captures outbound connections only, and matches process names on their first 16 characters (`docs.mitmproxy.org/stable/concepts/modes/#local-capture`). The reference CachyOS host runs a 7.x kernel.
2. **Whether the client trusts mitmproxy's CA.** Wine's `crypt32` loads host CA bundles, starting with `/etc/ssl/certs/ca-certificates.crt` (`wine-mirror/wine`, `dlls/crypt32/unixlib.c`; Valve's Proton Wine fork has the same list). On Arch, `trust anchor <cert>` plus `update-ca-trust` writes that bundle (`wiki.archlinux.org/title/Transport_Layer_Security`). **But** Proton runs games inside the Steam Linux Runtime container, which mounts its own `/etc/ssl/certs` and imports host CA certificates only when `PRESSURE_VESSEL_IMPORT_CA_CERTS` is set; the default is off, and the source calls it "a developer-facing rather than end-user-facing flag" (`steam-runtime-tools`, `pressure-vessel/runtime.c`). Inference from the source, not yet tested by this handbook: set `PRESSURE_VESSEL_IMPORT_CA_CERTS=1` in the launch options for the session, and verify trust before relying on it.
3. **Whether the client uses the OS store at all.** Unreal's HTTP stack can validate against a bundled `cacert.pem` (`Engine/Content/Certificates/ThirdParty/cacert.pem`, per Epic forum threads on the engine's certificate bundle), and games can pin certificates. If the client does not trust the host store, interception would require modifying the client's files or code. On a live-service client that is tampering; stop and fall back to boundary observation under the runtime gate ([dynamic-analysis-under-proton.md](dynamic-analysis-under-proton.md#observing-encrypted-traffic-at-the-boundary)).

Rules for interception sessions: never modify, replay, or inject requests; save flows under `captures/`; treat bearer tokens and cookies in flows as secrets.

Analyze flows with REA `inspect_web_network_capture` on the native mitmproxy flow file or a HAR export, then `inspect_analysis_view` with `view: {kind: "api-surface"}`. REA groups observed request lines by method, origin, and path template; its guide warns that "format matches are inferences, not handler identity or schemas". Write one message spec per route.

### Identify The Serialization Format

For custom TCP/WebSocket world protocols:

- **Protocol Buffers** — embedded descriptors are recoverable statically. pbtk extracts structures from "binaries containing embedded reflection metadata" (`pbtk-from-binary input_file [output_dir]`, `github.com/marin-m/pbtk`); protodump will "dump all Protobuf file descriptors from a given binary as *.proto files" (`github.com/arkadiyt/protodump`). Run them against the executable on disk. Without descriptors, `protoc --decode_raw` will "read an arbitrary protocol message from standard input and write the raw tag/value pairs" (protoc help).
- **FlatBuffers, MessagePack, Cap'n Proto** — library strings and schema identifiers in the binary; framing visible in captures.
- **Hand-written formats** — find the writer statically and read field order from it. Look for a length prefix and an opcode in the first bytes of each framed message; confirm with several captured instances of the same action.
- **Compression and encryption layers** — high-entropy payloads after a short plaintext header suggest compression (zlib/zstd/LZ4/Oodle markers in the binary) or encryption. Find the layer's call site statically before guessing.

Record the framing (length, opcode, flags) once in a protocol overview, then one spec per message from [../templates/protocol-message-spec.md](../templates/protocol-message-spec.md).

### Unreal NetDriver Traffic

Unreal's UDP protocol is hard to decode on the wire:

- It is bit-packed, with packet headers, acks, and bunch headers whose layout depends on the engine version, and per-property encoding that depends on reflection data and replication state.
- It may be encrypted by a packet handler component (`AESGCMHandlerComponent`), keyed through `OnReceivedNetworkEncryptionToken` ([../foundations/mmo-client-server-model.md](../foundations/mmo-client-server-model.md#unreals-built-in-replication)).
- Iris, when enabled, changes the replication serialization.

Practical approach:

1. Do not start with a wire decoder. Start from the RPC index and the engine source's bunch and property serialization code for the matching tag.
2. Use packet sizes and timing from `inspect_packet_capture` (`udp.length`, `frame.time_relative`) to correlate actions with traffic.
3. Where runtime is in scope, observe RPC parameters at the UFunction or `ProcessEvent` boundary instead of decoding bits ([dynamic-analysis-under-proton.md](dynamic-analysis-under-proton.md)).
4. Build a wire decoder only when interoperability requires it, as a port of the matching engine version's serialization code, validated against boundary observations.

### Dissectors REA Can Use

Write dissectors in Lua so both Wireshark and REA's TShark tools apply them:

- Wireshark Lua dissectors "must be registered to handle a payload type for another protocol" — for example `DissectorTable.get("tcp.port"):add(<port>, proto)` (`wiki.wireshark.org/Lua/Dissectors`; Wireshark developer guide example).
- TShark loads personal Lua plugins automatically. `tshark -G folders` prints the folder; on the reference host it is `~/.local/lib/wireshark/plugins`. Symlink `protocol/dissectors/<proto>.lua` there.
- REA's `inspect_packet_capture` runs TShark with caller-selected `fields`, `display_filter`, and `decode_as` (for example `tcp.port==<port>,<proto>`); it requires `REA_TSHARK_COMMAND`. Confirm REA sees your protocol with `inspect_packet_fields`, which searches installed protocol and field names, before querying fields.
- Start from [../templates/dissector-skeleton.lua](../templates/dissector-skeleton.lua). Name fields `<proto>.<message>.<field>` so specs, dissectors, and REA queries share vocabulary.

### Interoperability Implementations

Purpose-built tools that consume the documented protocol — a data viewer reading your own session's captures, an accessibility overlay, a preservation server after shutdown — are the end product of this work.

- Parsers and dissectors consume recorded traffic. That is always in scope.
- A private or emulated server is in scope for preservation after the official servers are gone (37 CFR 201.40, [../foundations/scope-legal-and-ethics.md](../foundations/scope-legal-and-ethics.md#the-legal-frame-united-states-and-european-union)) or for interoperability research against your own client, with the client pointed at it through its own configuration. Redirecting a live-service client usually means modifying it or its config; check the ToS first.
- Nothing built here sends traffic to live servers. A client of the live service is outside this handbook.

## Common Mistakes And Forbidden Patterns

- Capturing other hosts' traffic, or keeping captures with other players' data longer than `scope.md` allows.
- Replaying, modifying, or crafting messages to live servers.
- Assuming the host trust store reaches a Proton game without checking the runtime container behavior.
- Patching a live-service client's certificate bundle or pinning to enable interception.
- Writing a bit-level NetDriver decoder before checking boundary observation or engine source.
- Documenting a message from capture alone, with no writer or reader correlation.
- Committing captures, flows, HAR files, or tokens.

## Verification And Proof

- Every message spec cites capture Evidence (`inspect_pcap_capture`, `inspect_packet_capture`, or `inspect_web_network_capture`) and static Evidence (writer or reader `analyze_function`), or lists the missing half as an unknown.
- `inspect_packet_fields` returns the custom protocol's fields, proving REA's TShark loads the dissector.
- The dissector decodes every captured instance of a documented message without malformed-packet errors: `tshark -r <capture> -Y '<proto> && _ws.malformed'` returns nothing.

## Where To Go Next

- [../recipes/write-a-wireshark-dissector.md](../recipes/write-a-wireshark-dissector.md) — the dissector procedure.
- [../templates/protocol-message-spec.md](../templates/protocol-message-spec.md) — the spec format.
- [dynamic-analysis-under-proton.md](dynamic-analysis-under-proton.md) — boundary observation when the wire is opaque.
