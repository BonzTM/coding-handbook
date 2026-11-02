-- Template: Wireshark Lua dissector for a length-prefixed, opcode-tagged TCP protocol.
-- Copy to <workspace>/protocol/dissectors/<proto>.lua and replace every <placeholder>.
-- Rules: techniques/network-protocol-analysis.md ("Dissectors REA Can Use").
-- Field names follow <proto>.<message>.<field> so specs, dissectors and REA queries match.

local PROTO_NAME = "mmoproto"           -- <proto>: short lowercase filter name
local PROTO_DESC = "MMO World Protocol" -- <description>
local TCP_PORT = 0                      -- <port> from the capture classification; 0 = register via decode-as only
local HEADER_LEN = 6                    -- <bytes>: uint32 length + uint16 opcode in this skeleton

local proto = Proto(PROTO_NAME, PROTO_DESC)

-- Opcode names come from protocol/messages/*.md; one entry per verified spec.
local opcode_names = {
  -- [0x0001] = "<MessageName>",
}

local f = proto.fields
f.length = ProtoField.uint32(PROTO_NAME .. ".length", "Length", base.DEC)
f.opcode = ProtoField.uint16(PROTO_NAME .. ".opcode", "Opcode", base.HEX, opcode_names)
f.payload = ProtoField.bytes(PROTO_NAME .. ".payload", "Payload")
-- Per-message fields, for example:
-- f.move_x = ProtoField.float(PROTO_NAME .. ".move.x", "Move X")

local ef_short = ProtoExpert.new(PROTO_NAME .. ".short", "Message shorter than header",
  expert.group.MALFORMED, expert.severity.ERROR)
proto.experts = { ef_short }

-- Per-message payload decoders keyed by opcode. Each takes (payload_tvb, subtree).
local decoders = {
  -- [0x0001] = function(tvb, tree) tree:add_le(f.move_x, tvb(0, 4)) end,
}

-- Returns the PDU length so TCP reassembly can split and join messages.
local function get_pdu_len(tvb, pinfo, offset)
  return tvb(offset, 4):le_uint() -- <adjust>: does the length include the header?
end

local function dissect_pdu(tvb, pinfo, tree)
  pinfo.cols.protocol = PROTO_NAME
  local subtree = tree:add(proto, tvb(), PROTO_DESC)
  if tvb:len() < HEADER_LEN then
    subtree:add_proto_expert_info(ef_short)
    return tvb:len()
  end
  subtree:add_le(f.length, tvb(0, 4))
  local opcode = tvb(4, 2):le_uint()
  subtree:add_le(f.opcode, tvb(4, 2))
  pinfo.cols.info = opcode_names[opcode] or string.format("opcode 0x%04x", opcode)
  local payload = tvb(HEADER_LEN):tvb()
  local decoder = decoders[opcode]
  if decoder then
    decoder(payload, subtree)
  elseif payload:len() > 0 then
    subtree:add(f.payload, payload())
  end
  return tvb:len()
end

function proto.dissector(tvb, pinfo, tree)
  dissect_tcp_pdus(tvb, tree, 4, get_pdu_len, dissect_pdu)
  return tvb:len()
end

if TCP_PORT > 0 then
  DissectorTable.get("tcp.port"):add(TCP_PORT, proto)
end
-- Without a port, select it per capture: tshark -d tcp.port==<port>,mmoproto
-- or REA inspect_packet_capture decode_as ["tcp.port==<port>,mmoproto"].
DissectorTable.get("tcp.port"):add_for_decode_as(proto)
