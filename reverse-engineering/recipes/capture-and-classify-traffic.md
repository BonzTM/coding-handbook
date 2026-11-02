# Recipe: Capture And Classify Traffic

Use this to record one scripted play session of your own client and classify its transports before any protocol decoding. Governed by [../foundations/mmo-client-server-model.md](../foundations/mmo-client-server-model.md) and [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md).

> **Warning:** Capturing your own traffic is passive. It is still a runtime session for the client. Walk [../checklists/runtime-session-gate.md](../checklists/runtime-session-gate.md) first. The human starts the game and logs in; the agent only reads the saved capture.

## Files To Touch

- `captures/<build-id>-<action>-<yyyy-mm-dd>.pcapng` (new, git-ignored)
- `captures/<build-id>-<action>-<yyyy-mm-dd>.timeline.md` (new, git-ignored): wall-clock marks for each scripted step
- `evidence/<build-id>-traffic-<yyyy-mm-dd>.json` (new)
- `findings.md` (classification rows)

## Steps

1. Write the script of player actions with expected timing: launch → login → character select → enter world → idle 60 s → move → open inventory → zone change → log out.
2. Find the capture interface: `ip route get 1.1.1.1` prints the default-route interface.
3. Start the capture before launching the game, filtering only your own host:

   ```bash
   dumpcap -i <iface> -f "host $(ip -4 -o addr show <iface> | awk '{print $4}' | cut -d/ -f1)" \
     -w captures/<build-id>-session-$(date +%F).pcapng
   ```

4. The human launches the game and runs the script. Record each step's wall-clock time in the timeline file.
5. While in world, record the client's sockets: `ss -tunap | grep -i <exe-name>`. Paste the output into the timeline file.
6. Stop the capture after logout.
7. Summarize with REA (agent-safe; reads the file only):

   ```json
   {"tool": "inspect_pcap_capture",
    "path": "/abs/path/captures/<file>.pcapng",
    "detail": "summary"}
   ```

   Page DNS questions, TLS SNI values, endpoint and protocol distributions with `inspect_analysis_view` and the returned `parent_evidence_id`.
8. Classify each endpoint: purpose (login, meta, world, voice, telemetry, CDN), transport (TCP/UDP/QUIC), TLS or not, step(s) in the timeline where it is active.
9. For UDP endpoints active only in world, pull packet sizes and timing with `inspect_packet_capture` (needs `REA_TSHARK_COMMAND`), fields `frame.time_relative`, `udp.srcport`, `udp.dstport`, `udp.length`, filtered by the endpoint.
10. Write the five classification rows from [../foundations/mmo-client-server-model.md](../foundations/mmo-client-server-model.md#classify-before-you-dig) into `findings.md`, each citing the Evidence IDs.
11. Export the bundle to `evidence/<build-id>-traffic-<date>.json`.

## Invariants To Preserve

- Only your own client's traffic is captured; the capture filter names your host.
- No packet is sent, replayed, or modified.
- The capture file never enters git; it can hold session tokens, other players' names and chat.
- Each classification row says which timeline step it was observed in.

## Proof

- The Evidence bundle contains the `inspect_pcap_capture` summary with the capture's SHA-256.
- `findings.md` has a classification row for every endpoint the summary lists, or a row explaining why an endpoint is ignored (for example OS update traffic).
- `git status --ignored captures/` shows the capture as ignored.
