# Serialization

The JSON boundary: how wire shapes are declared, evolved, and proven so internal refactors do not drift contracts.

## Default Approach

Wire formats are explicit adapter-owned contracts. Pydantic v2 models parse and serialize HTTP/config/message boundaries; domain values remain plain typed objects. Map exactly once between them.

### Dedicated DTOs Own The Wire

Define request and response models beside the transport that owns them. Never return a SQLAlchemy model or domain dataclass directly. A mapping function makes field selection, normalization, aliases, and redaction reviewable.

Use stdlib `json` for small internal serialization tasks that do not require schema validation and already operate on JSON-native primitives. Use Pydantic when runtime parsing, aliases, constraints, or generated schema are the contract. Do not mix serializers for the same surface.

### Parse Before Domain Logic

Treat bytes, text, dictionaries, and third-party responses as untrusted. Bound the body or message before decoding, decode with one codec, validate the complete DTO, then map to a domain value. Domain code never receives `dict[str, object]` and never calls `model_validate`.

Pydantic conversion must not silently change contract meaning. Use strict field types where coercing strings into numbers or booleans would hide a producer defect. Normalization such as trimming is a named boundary rule with tests, not a broad validator added for convenience.

Map validation failures into the stable problem/message rejection shape. Never return Pydantic's raw error objects: their locations, URLs, and wording are framework details, not the published contract.

### Unknown Fields Follow Boundary Direction

Inbound commands to a service are strict:

```python
class CreateWidgetRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    display_name: str = Field(alias="displayName")
```

Typos and undeclared input fail rather than disappear. Pydantic's [`extra` configuration](https://docs.pydantic.dev/latest/api/config/#pydantic.config.ConfigDict.extra) defines `forbid`, `ignore`, and `allow` behavior.

Models consuming responses/events from an independently evolving producer are tolerant readers: ignore unknown fields while validating every field the consumer uses. This preserves the additive-evolution posture from [contracts and compatibility](contracts-and-compatibility.md). Never use a tolerant consumer model to accept undeclared command fields on an API you own.

### Field Names Are Explicit Contracts

Python attributes use `snake_case`; JSON uses `camelCase` for new handbook surfaces. Declare aliases or one reviewed alias generator, and configure validation/serialization direction deliberately. Always serialize public DTOs with aliases. A Python rename must not silently rename a wire key.

Do not expose Pydantic's internal field names or framework-generated shape accidentally. Pin exact output keys with golden tests.

Aliases must not collide after generation. Acronyms receive explicit aliases where a generator would produce an ambiguous result. Configure accepted input names and emitted aliases deliberately; accepting both old and new spellings can accidentally publish two contracts.

### Null, Omitted, And Default Differ

`T | None` permits JSON null; it does not by itself make a field optional in every Pydantic declaration. Decide separately whether a field is required, may be null, or has a default. PATCH-like commands that distinguish omitted from explicit null preserve Pydantic's field-set information only in the adapter, then map to an explicit domain command.

Required response fields are emitted even when zero/empty. Omission is a compatibility decision, not a cleanup optimization.

For collections, decide whether empty and null differ. New response collection fields default to a required array and emit `[]` for no values; null requires distinct domain meaning. Preserve order only when promised. Sort otherwise nondeterministic values before serialization.

### Numeric Precision

- Money is integer minor units plus currency; never binary float.
- Use `Decimal` only when the contract requires decimal fractions. Serialize as a documented string when consumer precision is not guaranteed, and parse with explicit bounds/rounding.
- Large integers consumed by JavaScript require a documented string representation when they may exceed exact JSON-number interoperability.
- `float` is limited to approximate measurements where NaN/infinity policy and precision loss are explicit. Public JSON rejects non-finite values.

### Strings And Bytes Are Bounded

Declare length and range limits at the DTO boundary and align them with body limits and persistence constraints. Reject oversized values before expensive normalization or downstream I/O. Decode text with a stated encoding, normally UTF-8, and reject malformed input.

Binary payloads use an explicit content type or documented base64 field with decoded-size bounds. They do not become arbitrary strings in an otherwise JSON contract.

### Datetimes Are Aware UTC Instants

Wire timestamps are timezone-aware ISO 8601 strings normalized to UTC. Naive datetimes are rejected. Parse offsets, convert to `timezone.utc`, and emit a consistent UTC form. Civil-time zone names belong only where the contract is explicitly about local schedules; see [time](time.md).

Durations name their unit on the wire (`timeoutMillis`) or use a documented duration representation. Convert once into `timedelta`; a bare numeric field named `timeout` is forbidden.

### Binary And Unsafe Formats

`pickle` never crosses a trust boundary; it can execute attacker-controlled code during deserialization, as the [Python documentation warns](https://docs.python.org/3/library/pickle.html#warning). Prefer JSON for interoperable data and protobuf for the gRPC contract. Untrusted XML uses `defusedxml`; YAML uses safe loading per [security](../operations/security.md).

### Error Responses Are Contracts

All HTTP failures use the single RFC 9457-style `application/problem+json` mapping described in [errors and logging](errors-and-logging.md) and [HTTP services](../services/http-services.md). Validation problems have stable field paths/codes; 5xx bodies are opaque and carry a request identifier, never exception details.

### Schema Generation Is Reviewed Output

Pydantic/FastAPI JSON Schema matches runtime behavior: required fields, aliases, nullability, formats, bounds, examples, and discriminators. A custom validator whose accepted language cannot be expressed in schema requires explicit documentation and behavior tests.

Polymorphic payloads use an explicit discriminated union with stable discriminator values. Shape guessing by field presence is forbidden because additive fields make it ambiguous. A new member is additive only after consumers prove unknown-member behavior.

### Round-Trip And Golden Proof

Round-trip tests prove semantic preservation: DTO -> JSON -> DTO. Golden tests pin exact keys, alias casing, null/omission, enum strings, timestamp form, decimal representation, and error envelopes. Review golden updates as contract changes; never regenerate and accept them blindly.

Generated OpenAPI is additional contract evidence, not a substitute for behavioral serialization tests.

## Common Mistakes And Forbidden Patterns

- Pydantic DTOs, SQLAlchemy models, or JSON dictionaries used as domain entities.
- One model reused for inbound commands and tolerant external consumption despite different unknown-field policies.
- Field aliases left implicit, so an internal rename breaks clients.
- `T | None`, omission, null, and default treated as the same state.
- `model_dump()` returned without explicit alias/exclusion policy.
- Money as float, non-finite floats in JSON, or undocumented decimal/large-integer representation.
- Naive or local-zone datetimes crossing a wire boundary.
- `pickle` or unsafe YAML/XML deserialization on untrusted bytes.
- Ad hoc error JSON or internal exception details in a 5xx response.
- Golden files updated without compatibility review.
- Unbounded bodies, strings, or decoded binary; lossy text repair; expensive normalization before bounds checks.
- Raw Pydantic validation errors exposed as the public error contract.
- Polymorphism inferred from incidental fields instead of a stable discriminator.

## Verification And Proof

```bash
uv run pytest -k "serialization or schema or contract"
make verify
```

Serialization is done when every DTO has round-trip and exact-shape proof; inbound unknown fields fail; external-consumer unknown fields are ignored without weakening known-field validation; aliases, null/omission/default, enums, UTC timestamps, decimals/large integers, and opaque problem responses are exercised; and the OpenAPI diff is reviewed.

Also test body/value limits, malformed text or binary encoding where accepted, collection ordering, coercion policy, discriminated-union unknown members, and alignment between runtime rejection and generated schema.

Related: [data modeling](data-modeling.md), [contracts and compatibility](contracts-and-compatibility.md), and [add HTTP endpoint](../recipes/add-http-endpoint.md).
