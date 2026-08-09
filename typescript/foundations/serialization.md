# Serialization

Rules for translating untrusted wire representations into typed domain values without losing meaning.

## Default Approach

Treat every decoded value as `unknown`, parse it with Zod 4, and map it into a domain-owned type.

### JSON Boundaries

Define an explicit schema for request bodies, responses, messages, configuration files, and third-party payloads. Set size limits before parsing. Reject malformed JSON, invalid encodings, excessive nesting where relevant, and values outside documented bounds.

Choose unknown-key behavior deliberately. Public requests should normally reject unexpected keys when they may indicate a client error or security issue; compatibility-oriented readers may strip or preserve them only by contract.

### Wire Types

JSON supports null, booleans, numbers, strings, arrays, and objects. It does not preserve `Date`, `bigint`, `Map`, `Set`, `undefined`, class identity, or numeric precision beyond interoperable bounds.

- Encode timestamps as RFC 3339 UTC strings and parse them explicitly.
- Encode integers outside the safe JavaScript range as canonical decimal strings.
- Encode bytes as a documented base64 variant.
- Encode maps with non-string keys as arrays of entries or a defined object mapping.
- Omit optional fields only when omission is part of the contract; never rely on `JSON.stringify` silently dropping `undefined`.

### Schema Ownership

The producer owns the published wire contract; each consumer validates at ingress. Infer transport types from the Zod schema when that keeps one source of truth, but map them before domain use. An inferred compile-time type does not replace runtime parsing.

### Zod DTO And Domain Mapping

Keep the wire schema, inferred DTO, and explicit domain mapper together. This example accepts an RFC 3339 timestamp with an offset and a canonical non-negative integer string:

```ts
import { z } from "zod";

const decimalBigInt = z
  .string()
  .regex(/^(0|[1-9]\d*)$/)
  .transform((value) => BigInt(value));

export const widgetDtoSchema = z.strictObject({
  id: z.uuid(),
  name: z.string().min(1).max(100),
  createdAt: z.iso.datetime({ offset: true }),
  sequence: decimalBigInt,
  description: z.string().max(500).nullable(),
});

export type WidgetWireDto = z.input<typeof widgetDtoSchema>;
export type WidgetDto = z.infer<typeof widgetDtoSchema>;

export type Widget = Readonly<{
  id: WidgetId;
  name: string;
  createdAt: Date;
  sequence: bigint;
  description: string | null;
}>;

export function mapWidgetDto(dto: WidgetDto): Widget {
  return {
    id: parseWidgetId(dto.id),
    name: dto.name,
    createdAt: new Date(dto.createdAt),
    sequence: dto.sequence,
    description: dto.description,
  };
}

export function parseWidget(input: unknown): Widget {
  return mapWidgetDto(widgetDtoSchema.parse(input));
}

export function encodeWidget(widget: Widget): WidgetWireDto {
  return {
    id: widget.id,
    name: widget.name,
    createdAt: widget.createdAt.toISOString(),
    sequence: widget.sequence.toString(10),
    description: widget.description,
  };
}
```

The transform output makes `WidgetDto["sequence"]` a `bigint`; `WidgetWireDto["sequence"]` remains the JSON decimal string. Encoding maps the domain value back explicitly.

### Unknown-Key Policy In Zod 4

Zod 4 exposes object policy through top-level constructors. Use strict objects for commands and public writes where unexpected keys should identify a client error:

```ts
const createWidgetSchema = z.strictObject({
  name: z.string().min(1).max(100),
});

createWidgetSchema.parse({ name: "Meter", admin: true }); // throws
```

Use a loose object only for an intentionally forward-compatible reader that must retain additive fields:

```ts
const eventEnvelopeSchema = z.looseObject({
  id: z.uuid(),
  type: z.literal("widget.created"),
  occurredAt: z.iso.datetime({ offset: true }),
});

const envelope = eventEnvelopeSchema.parse({
  id: "550e8400-e29b-41d4-a716-446655440000",
  type: "widget.created",
  occurredAt: "2026-08-08T12:00:00Z",
  futureField: "preserved",
});
```

Do not select loose mode merely to avoid updating a schema. The contract must state who owns retained unknown fields and whether they may be re-emitted.

### Mapping And Errors

Keep schema parsing separate from domain construction so failures identify the correct boundary. Return safe, stable problem details to callers; do not echo the entire rejected payload or internal schema diagnostics into production logs.

Normalize only documented equivalences. Avoid silent trimming, case folding, timezone conversion, or numeric coercion that changes user intent. If coercion is accepted for form input, constrain it to that adapter.

### Golden And Round-Trip Proof

Commit small, human-readable golden fixtures for stable public shapes. Round-trip tests are useful when both encoding and decoding are owned, but also assert the exact wire representation so two mutually wrong functions cannot pass together.

Test older supported fixtures against current readers. Contract fixtures contain synthetic data only, never production records or secrets.

A golden test pins the wire representation independently of the decoder:

```ts
import { describe, expect, it } from "@jest/globals";

describe("widget JSON", () => {
  it("uses UTC time and decimal-string sequence", () => {
    const json = JSON.stringify({
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Meter",
      createdAt: "2026-08-08T12:00:00.000Z",
      sequence: "9007199254740993",
      description: null,
    });

    const golden =
      '{"id":"550e8400-e29b-41d4-a716-446655440000","name":"Meter","createdAt":"2026-08-08T12:00:00.000Z","sequence":"9007199254740993","description":null}';

    const decoded: unknown = JSON.parse(json);
    expect(json).toBe(golden);
    expect(parseWidget(decoded)).toMatchObject({
      sequence: 9007199254740993n,
    });
  });
});
```

### Streaming And Large Payloads

Do not buffer unbounded input. Define maximum bytes, record counts, and processing deadlines. For streams, propagate `AbortSignal`, handle partial records explicitly, and release the reader on success, failure, and cancellation.

## Common Mistakes And Forbidden Patterns

- `JSON.parse(value) as T` or trusting generic response typing.
- Serializing domain classes and assuming constructors or methods survive.
- Dates without timezone offsets, locale-specific date strings, or implicit local-time parsing.
- `bigint` passed directly to JSON or large identifiers represented as numbers.
- Logging rejected bodies, authorization headers, or schema values containing secrets.
- Schema coercion that accepts ambiguous or surprising representations.
- Round-trip tests without exact-shape or compatibility fixtures.

## Verification And Proof

- Valid, malformed, oversized, missing, null, extra-key, and boundary-value tests exist.
- Golden fixtures pin field names, omission rules, timestamp form, and large-number representation.
- Supported old fixtures parse with current code; incompatible changes follow the deprecation process.
- Fuzz or property tests target custom parsers with large input spaces.
- Stream tests cover abort, truncated input, size limits, and resource cleanup.
- Logs and returned errors contain no rejected sensitive payload.

Related: [data-modeling.md](data-modeling.md), [contracts-and-compatibility.md](contracts-and-compatibility.md), and [../quality/testing.md](../quality/testing.md). API anchor: [Zod 4 object schemas and ISO datetimes](https://zod.dev/api).
