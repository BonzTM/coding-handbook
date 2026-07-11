# Serialization

The JSON encoding boundary: how wire shapes are declared, evolved, and proven, so the contract never drifts with an internal refactor.

## Default Approach

The wire format is a contract, not a reflection of your types. Declare it explicitly on dedicated transport DTOs, control naming, omission, and unknown-member policy deliberately, and keep all encoding configuration at the boundary. The default codec is `System.Text.Json` with **source generation** — `Newtonsoft.Json` is forbidden in new code without an ADR.

This doc governs the JSON boundary. For the shape of the data behind it see [data-modeling.md](data-modeling.md); for the rules that make a shape safe to change over time see [contracts-and-compatibility.md](contracts-and-compatibility.md), which mandates stable request/response shapes and new-optional-fields as the evolution mode — both rest on the conventions specified here.

### Source-Generated Contexts At Trust Boundaries

Every trust boundary serializes through a `JsonSerializerContext` — the source generator emits the (de)serialization code at compile time, so the wire surface is a closed, reviewable set of types instead of whatever reflection finds at runtime. It is faster, trimming/AOT-safe, and — the contract point — **adding a type to the wire is a visible diff** on the context.

```csharp
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    UseStringEnumConverter = true)]
[JsonSerializable(typeof(CreateOrderRequest))]
[JsonSerializable(typeof(CreateOrderResponse))]
[JsonSerializable(typeof(OrderListResponse))]
internal sealed partial class OrdersJsonContext : JsonSerializerContext;
```

Wire it into the host once, in `Program.cs`:

```csharp
builder.Services.ConfigureHttpJsonOptions(options =>
    options.SerializerOptions.TypeInfoResolverChain.Insert(0, OrdersJsonContext.Default));
```

- One context per transport surface (the API's DTOs), not one per type. Serializer options live on the context attribute, in one place — never scattered `new JsonSerializerOptions()` per call site, which silently forks the contract.
- Reflection-based `JsonSerializer.Serialize(value)` with ad hoc options is acceptable only for internal, non-contract uses (debug dumps, tests' expected-value construction). Contract surfaces go through the context.

### Separate DTOs From Domain Types

Wire types (DTOs) are defined next to the endpoints in `Orders.Api`; they are not your `Orders.Core` domain types. Map between them explicitly at the boundary.

```csharp
public sealed record CreateOrderResponse
{
    public required string OrderId { get; init; }
    public required OrderStatus Status { get; init; }
    public required DateTimeOffset CreatedAt { get; init; }
}

internal static class OrderMappings
{
    public static CreateOrderResponse ToResponse(this Order order) => new()
    {
        OrderId = order.Id.ToString(),
        Status = order.Status,
        CreatedAt = order.CreatedAt,
    };
}
```

- Serializing a domain type directly couples the public wire contract to internal refactors: rename a Core property and you silently break clients, or you freeze the domain model to protect the wire. The DTO breaks that coupling — the mapping function is the one place the two shapes meet.
- The mapping is plain code, tested with the endpoint, per [../services/http-services.md](../services/http-services.md) and [../recipes/add-http-endpoint.md](../recipes/add-http-endpoint.md). Do not reach for reflection-based mappers; they reintroduce the implicit coupling the DTO removed.
- The project boundary enforces the direction: `Orders.Core` has no reference to `System.Text.Json` attributes or ASP.NET — wire concerns cannot leak inward.

### Naming Policy: camelCase, Declared Once

The wire convention is **camelCase**, set once via `PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase` on the context (ASP.NET Core's web default agrees). C# properties stay PascalCase; the policy does the mapping, so a property rename is the only thing that can move a key — and that is caught by the golden tests below.

- Use `[JsonPropertyName("legacy_key")]` only to pin a deviation the contract requires (matching an external spec, preserving a historical key). Each one is a visible, deliberate exception.
- Never rely on "no policy" so keys mirror C# names verbatim — that leaks C# naming into the contract and makes every rename a silent wire change.
- One case convention per surface, uniformly. A mixed-case payload is a bug.

### Unknown-Member Policy Is Explicit

`System.Text.Json` **ignores** unknown JSON members by default. That is the right default for public surfaces — it is what lets a new optional field ship without coordinating every caller, per [contracts-and-compatibility.md](contracts-and-compatibility.md) — but it must be a decision, not an accident.

- Public / additive surfaces: keep the default (skip unknown members) and document it.
- Strict internal or versioned surfaces, where a typo or stale client should fail loudly: annotate the DTO with `[JsonUnmappedMemberHandling(JsonUnmappedMemberHandling.Disallow)]` so an unrecognized member fails deserialization.
- Document which mode each surface uses, and prove both with decode tests (below).

### Enums As Strings On The Wire

Numbers are the `System.Text.Json` default for enums, and numbers are the trap: `2` in a payload is undebuggable, and renumbering corrupts consumers. `UseStringEnumConverter = true` on the context makes names the wire form (`"shipped"`, not `2`).

- The enum itself still owes explicit pinned values and a reserved `Unknown = 0`, per [data-modeling.md](data-modeling.md) — a missing member still materializes as the default value.
- `JsonStringEnumConverter` also *reads* integers by default. On strict surfaces, register `new JsonStringEnumConverter(allowIntegerValues: false)` so numeric payloads are rejected instead of quietly accepted.
- An unknown enum *name* fails deserialization — that is the correct behavior; do not "fix" it by mapping unknowns to a live value.

### Numbers Are Numbers; Money Is decimal

`JsonNumberHandling` stays **strict** (the default): numbers serialize as JSON numbers and quoted numbers are rejected on read. Never enable `AllowReadingFromString` on a contract surface — it lets two client populations diverge on the same field.

- **Money is `decimal`, never `double`/`float`.** Binary floating point cannot hold `0.10` exactly; sums drift. `decimal` is exact for decimal quantities and serializes as a JSON number. Carry an explicit currency code beside it ([data-modeling.md](data-modeling.md)); column precision lives in [../services/database.md](../services/database.md).
- `double` is acceptable only for genuinely approximate quantities (ratios, scores, measurements) where lost precision is harmless and documented.
- The 2^53 caveat: JavaScript clients decode every JSON number as an IEEE-754 double, so a `long` above 2^53 silently corrupts. The default identifier sidesteps this — UUIDv7 IDs are strings on the wire. If a surface must carry a large `long` to JS-adjacent consumers, serialize that field as a string via a dedicated converter and document it in the contract; do not weaken number handling globally.

### Timestamps: DateTimeOffset, ISO 8601

Instants cross the wire as `DateTimeOffset`, which `System.Text.Json` writes in the ISO 8601 extended profile (`2026-07-10T14:30:00+00:00`) — offset always present, round-trippable, sortable as text.

- Never put `DateTime` on a DTO: its `Kind` ambiguity means the same value can serialize with or without a trustworthy offset depending on how it was produced. The full instant/zone discipline lives in [time.md](time.md).
- Dates without a time of day are `DateOnly` (`"2026-07-10"`); times without a date are `TimeOnly`. Do not smuggle them in a midnight `DateTimeOffset`.
- Custom timestamp formats require a documented external contract and a dedicated converter — never a repo-wide format override.

### Required Properties

Mark contract-mandatory DTO members `required`. The serializer enforces it: a payload missing a `required` member fails deserialization instead of materializing a silently defaulted object. This is the missing-field guard that Go teams write by hand — here the type system and codec do it.

- `required` answers *presence*. Value validation (ranges, formats, cross-field rules) is a separate boundary step — DataAnnotations on request DTOs via the built-in minimal-API validation, per [../services/http-services.md](../services/http-services.md).
- On responses, do not use nullability or omission to fake optional-ness on a field the contract requires; emit it, even when zero-valued, so consumers see the contract is intact.
- Absent vs `null` vs value is a real tri-state: a `required string?` member must be *present* but may be `null`; a non-required `string?` may be absent. Choose per field what the contract means and prove it with decode tests.

### Polymorphism Via Declared Discriminators

When one wire field carries several shapes, declare the closed set with `[JsonPolymorphic]`/`[JsonDerivedType]` — never roll your own type-sniffing.

```csharp
[JsonPolymorphic(TypeDiscriminatorPropertyName = "type")]
[JsonDerivedType(typeof(CardPayment), "card")]
[JsonDerivedType(typeof(BankTransferPayment), "bank_transfer")]
public abstract record Payment
{
    public required decimal Amount { get; init; }
    public required string Currency { get; init; }
}

public sealed record CardPayment : Payment
{
    public required string Last4 { get; init; }
}

public sealed record BankTransferPayment : Payment
{
    public required string Iban { get; init; }
}
```

- The discriminator values (`"card"`) are wire contract, independent of type names — a C# rename cannot move them. Evolve the set additively per [contracts-and-compatibility.md](contracts-and-compatibility.md).
- The set is closed: an unknown discriminator fails deserialization, and serializing an undeclared derived type fails by default. Both failures are correct — do not register a catch-all.
- Deserializing to `object` or dynamic shapes at a trust boundary is forbidden. If a surface is genuinely schemaless, parse it as `JsonElement`/`JsonNode` at the edge, validate, and convert to typed values before it goes anywhere.

### Decode Defensively

Untrusted input is bounded, parsed strictly per policy, then validated — in that order.

- **Bound the body.** Kestrel enforces a default request-body limit; tighten it per endpoint where the expected payload is small (`[RequestSizeLimit]` on controllers, `IHttpMaxRequestBodySizeFeature` or route-group conventions for minimal APIs). An endpoint that accepts a JSON command has no business reading tens of megabytes.
- **Depth is bounded** by `JsonSerializerOptions.MaxDepth` (default 64) — do not raise it for a trust boundary.
- **Deserialization never validates business rules.** After a successful bind, the built-in minimal-API validation runs DataAnnotations on the request DTO and rejects with a ProblemDetails response; deeper invariants are enforced by domain constructors ([data-modeling.md](data-modeling.md)). The pipeline order and error mapping live in [../services/http-services.md](../services/http-services.md).

### Error Responses

A failure is a wire shape too. Every endpoint returns the **same** error contract: RFC 9457 `ProblemDetails` (`application/problem+json`) — ASP.NET Core's native shape, produced by `AddProblemDetails` and the `IProblemDetailsService`, extended with `requestId` and, for validation failures, the `errors` field→messages map. A bare `{"error":"..."}` string is not a contract — it gives the client one human sentence and nothing to branch on.

```csharp
builder.Services.AddProblemDetails(options =>
    options.CustomizeProblemDetails = context =>
        context.ProblemDetails.Extensions["requestId"] =
            context.HttpContext.TraceIdentifier);
```

```json
{
  "type": "https://api.orders.example/problems/validation-failed",
  "title": "The request has invalid fields.",
  "status": 400,
  "errors": {
    "items[0].qty": ["Must be greater than or equal to 1."]
  },
  "requestId": "0HNC7B2V4KDPM:00000003"
}
```

- **`type` is the machine-readable signal a client branches on** — a stable URI per failure kind — and it is **not** the HTTP status. The status is the coarse channel (`4xx` vs `5xx`); `type` is the precise one. Two `404`s with different `type`s are different failures; a client that switches on the status alone cannot tell them apart, so never make the status the only error signal.
- The set of `type` URIs is a documented part of the wire contract and evolves under the same rules as any other field — add types additively, never repurpose or silently drop one — per [contracts-and-compatibility.md](contracts-and-compatibility.md).
- Validation failures return `TypedResults.ValidationProblem(errors)`: one `errors` entry per offending input, keyed by the field path, so a client can attach messages to form fields without parsing prose.
- **5xx is opaque.** Never put exception messages, stack traces, or connection strings in `detail`. Server faults return a generic problem plus the `requestId`; the detail goes to the logs keyed by that id (see [errors-and-logging.md](errors-and-logging.md)), where an operator can find it. The client gets a correlation handle, not your internals.
- The mapping from a domain error to `(status, type)` happens once, at the boundary — an exception handler plus one result-mapping helper, not hand-built problem objects inline in every endpoint. Where that helper sits in the pipeline lives in [../services/http-services.md](../services/http-services.md).

### Codec Choice: System.Text.Json Is The Mandate

`System.Text.Json` with source generation is the codec. **`Newtonsoft.Json` is forbidden in new code without an ADR** ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)) — it drags in a second options universe (contract resolvers, its own attribute set, its own null semantics), and two codecs on one service means two subtly different contracts. The only defensible reasons are a dependency that demands it or a legacy contract that relies on Newtonsoft-specific behavior; both are ADR material, and the blast radius is contained to the boundary that needs it.

## Common Mistakes And Forbidden Patterns

- Serializing `Orders.Core` domain types directly on the wire, or putting `[JsonPropertyName]`/converter attributes on domain types instead of DTOs.
- `new JsonSerializerOptions { ... }` scattered per call site — the contract forks silently; options live on the context.
- Relying on default (reflection, PascalCase) serialization so a C# rename silently changes a wire key.
- Enums as numbers on the wire, or accepting integer enum values on a strict surface.
- Money as `double`, or `AllowReadingFromString` enabled on a contract surface to paper over a sloppy client.
- `DateTime` on a DTO instead of `DateTimeOffset`/`DateOnly`, or a repo-wide custom timestamp format.
- Optional-by-accident: contract-mandatory members not marked `required`, so a missing field materializes a defaulted object instead of an error.
- Unknown-member policy unstated: strict surfaces silently ignoring typos, or public surfaces set to Disallow so every additive producer change breaks old consumers.
- Hand-rolled polymorphism (type-sniffing on a magic field, deserializing to `JsonNode` and switching) instead of `[JsonPolymorphic]` declared discriminators.
- Returning a bare `{"error":"..."}` string or an ad hoc per-endpoint error shape instead of ProblemDetails.
- Leaking exception text or stack traces into a 5xx body; the problem body must be opaque, with detail in the logs under the `requestId`.
- Treating the HTTP status as the only error signal — omitting a stable `type`, so two distinct `404`s are indistinguishable to the client.
- Adding `Newtonsoft.Json` (or `[JsonObject]`-era patterns) to new code without an ADR.

## Verification And Proof

- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit

The serialization boundary is done when:

- every wire type is registered on the source-generated context, has a **round-trip test** (serialize then deserialize yields an equal value) and a **golden test** asserting exact bytes/keys, so a rename, casing slip, or converter change is caught as a diff.
- decode tests prove the unknown-member policy for each surface: strict surfaces **reject** an unknown member; additive surfaces **ignore** it without error.
- enum tests prove names round-trip, an unknown name is rejected, and (on strict surfaces) an integer payload is rejected.
- `required`-member tests prove a payload missing a mandatory field fails deserialization; tri-state fields (present-null vs absent vs value) decode distinctly where the contract needs the distinction.
- money tests prove `decimal` round-trips exactly; any string-serialized `long` field has an explicit converter test.
- polymorphic surfaces have per-discriminator round-trip tests plus a test that an unknown discriminator fails.
- a body-size cap is enforced and tested (oversized body is rejected, not buffered to OOM).
- a **golden test pins the ProblemDetails envelope** — exact keys, the `type` URI, the `errors` shape, `requestId` present — so a thinning or rename of the error contract shows up as a diff, the same way response goldens do.
- a validation-failure test asserts the response carries field-keyed `errors` entries, not just a top-level title.
- a 5xx test asserts the body carries **no internal detail** — generic `type`/`title`, a `requestId` present, none of the exception text — while the detail is confirmed in the log line for that id.

## Where To Go Next

- [data-modeling.md](data-modeling.md) — the domain shapes behind the DTO, typed IDs, and where validation lives.
- [contracts-and-compatibility.md](contracts-and-compatibility.md) — why stable shapes and additive-only evolution are the rule this doc implements.
- [../services/http-services.md](../services/http-services.md) — where DTOs, binding, validation, and the ProblemDetails mapping live in an endpoint.
- [../recipes/add-http-endpoint.md](../recipes/add-http-endpoint.md) — the end-to-end recipe that wires a DTO to a route.
- [time.md](time.md) — the instant/zone discipline behind `DateTimeOffset` on the wire.
