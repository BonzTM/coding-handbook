# Template: Protocol Message Spec

Copy to `protocol/messages/<message-name>.md`, one file per message type. The method for filling it lives in [../techniques/network-protocol-analysis.md](../techniques/network-protocol-analysis.md).

```markdown
# <MessageName>

Direction: client→server | server→client | both
Transport: <UDP UE NetDriver bunch | TCP custom | WebSocket | HTTPS REST>
Channel / route: <actor channel + RPC name | opcode 0x.. | HTTP METHOD /path/{id}>
First seen build: <build-id>; last verified build: <build-id>
Status: draft | verified | changed-in <build-id>

## Trigger
<What player or server action causes it. Observation from capture <evidence_id>.>

## Layout
| Offset/order | Field | Type | Encoding | Source of claim |
|---|---|---|---|---|
| 0 | <field> | <uint16 / FVector_NetQuantize / varint / string> | <LE / bit-packed / zigzag / UTF-8 len-prefixed> | <evidence_id: capture field / decompiled writer> |

## Semantics
- <field>: <meaning, units, range>. Inference from <evidence_ids>.

## Server Validation (Observed)
- <What the server rejects or corrects, seen in recorded responses. Never probed
  against live servers with crafted input.>

## Reproduction
- Capture: captures/<file> frames <n-m> (not committed)
- Dissector: protocol/dissectors/<proto>.lua field <name>

## Unknowns
- U-xxx <question>
```
