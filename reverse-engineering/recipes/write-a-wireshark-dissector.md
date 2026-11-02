# Recipe: Write A Wireshark Dissector

Use this when a classified custom protocol (TCP or WebSocket world traffic) has at least one message spec in draft and you want Wireshark, TShark, and REA to decode it. Governed by [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md#dissectors-rea-can-use).

## Files To Touch

- `protocol/dissectors/<proto>.lua` (new, from [../templates/dissector-skeleton.lua](../templates/dissector-skeleton.lua))
- `protocol/messages/<message>.md` (update: dissector field names)
- `evidence/<build-id>-dissector-<yyyy-mm-dd>.json` (new)
- `findings.md` (spec verification rows)

## Steps

1. Copy the skeleton to `protocol/dissectors/<proto>.lua`. Set `PROTO_NAME`, `PROTO_DESC`, and the header layout from the framing finding.
2. Add one `opcode_names` entry and one `decoders` function per message spec. Name fields `<proto>.<message>.<field>`.
3. Link it into TShark's personal Lua plugin folder. Find the folder with `tshark -G folders` (line "Personal Lua Plugins"), then:

   ```bash
   ln -sf "$PWD/protocol/dissectors/<proto>.lua" ~/.local/lib/wireshark/plugins/<proto>.lua
   ```

4. Check it loads: `tshark -G fields | grep -c '^F.*<proto>\.'` prints the number of registered fields.
5. Decode a capture directly:

   ```bash
   tshark -r captures/<file>.pcapng -d tcp.port==<port>,<proto> \
     -T fields -e frame.number -e <proto>.opcode -e _ws.malformed | head
   ```

6. From the agent, confirm REA sees the protocol with `inspect_packet_fields` (search `<proto>`), then extract fields:

   ```json
   {"tool": "inspect_packet_capture",
    "path": "/abs/path/captures/<file>.pcapng",
    "fields": ["<proto>.opcode", "<proto>.<message>.<field>"],
    "decode_as": ["tcp.port==<port>,<proto>"],
    "display_filter": "<proto>"}
   ```

   REA needs `REA_TSHARK_COMMAND` set to the absolute `tshark` path in its environment. TShark loads the personal plugin from the REA process's `$HOME`.
7. Compare decoded values against the scripted actions in the capture's timeline and against the writer or reader Evidence. Update each spec's Layout table with the dissector field name and both Evidence IDs.
8. Export the Evidence bundle and update `findings.md`.

## Invariants To Preserve

- The dissector only reads captures; nothing it produces is sent anywhere.
- Field names match the message specs exactly.
- An opcode without a verified spec decodes as raw payload, not guessed fields.

## Proof

- `tshark -r <capture> -d tcp.port==<port>,<proto> -Y '<proto> && _ws.malformed'` prints nothing.
- `inspect_packet_capture` returns the documented fields for every captured instance of each verified message.
- This procedure was validated on the reference host: the unmodified skeleton, loaded from a personal plugin folder, decoded a synthetic message through `rea inspect-packet-capture` with `--decode_as 'tcp.port==7777,mmoproto'`.
