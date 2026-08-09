# Data Modeling

Domain modeling rules for preserving meaning across TypeScript, PostgreSQL, and wire boundaries.

## Default Approach

Model domain facts explicitly and translate them at every external boundary.

### Domain Values

- Use readonly object types for records and discriminated unions for finite states.
- Represent absence deliberately: missing, `undefined`, `null`, and empty are different contracts.
- With `exactOptionalPropertyTypes`, use `field?: T` only when omission is meaningful; use `field: T | null` when present-but-empty is meaningful.
- Construct invariant-bearing values through functions that validate and return a domain value or typed failure.
- Use branded IDs when accidental interchange between identical primitive representations is hazardous.

Domain objects do not carry Fastify requests, database rows, React state, or Zod internals. They represent business meaning after validation.

### Literal Unions Over Enums

Use an `as const` value set when runtime iteration is required and derive the type from it. Use a plain union when runtime values are unnecessary:

```ts
export const ORDER_STATUSES = ["pending", "paid", "cancelled"] as const;
export type OrderStatus = (typeof ORDER_STATUSES)[number];

export type DeliveryChannel = "email" | "sms";
```

Do not introduce a TypeScript `enum` by default. It emits or implies a runtime construct, can obscure the wire representation, and does not replace schema validation. Use an enum only when interoperating with an existing runtime API that requires it; isolate the conversion at that adapter.

### Optionality Under Exact Types

Omission and explicit null are separate domain facts:

```ts
type CreateWidget = Readonly<{
  name: string;
  description?: string;
}>;

type WidgetRecord = Readonly<{
  id: WidgetId;
  description: string | null;
}>;

const omitted: CreateWidget = { name: "Meter" };
const cleared: WidgetRecord = { id: parseWidgetId("550e8400-e29b-41d4-a716-446655440000"), description: null };
```

Under `exactOptionalPropertyTypes`, `description?: string` permits absence, not `{ description: undefined }`. Use a patch union when an update must distinguish keep, set, and clear:

```ts
type DescriptionChange =
  | Readonly<{ kind: "keep" }>
  | Readonly<{ kind: "set"; value: string }>
  | Readonly<{ kind: "clear" }>;
```

### Collections And Mutation

Prefer `ReadonlyArray`, `ReadonlyMap`, and readonly properties across module boundaries. Copy mutable caller-owned input when retaining it. Return new state from domain transitions unless measured performance requires controlled mutation.

Choose a collection by semantics: array for ordered sequences, set for uniqueness, map for keyed lookup. Do not use plain objects as general-purpose maps when prototype keys or arbitrary external keys are possible.

Readonly modeling makes ownership visible:

```ts
type OrderLine = Readonly<{
  sku: string;
  quantity: number;
}>;

type Order = Readonly<{
  id: string;
  lines: ReadonlyArray<OrderLine>;
}>;

export function addLine(order: Order, line: OrderLine): Order {
  return { ...order, lines: [...order.lines, line] };
}
```

Readonly is shallow. Copy nested input before retaining it, and do not expose a mutable alias through a supposedly immutable aggregate.

### Identifiers And Money

Treat identifiers as opaque strings after format validation; never infer authorization or chronology from an identifier unless that is its documented contract.

Represent money as an integer minor-unit amount plus an ISO currency code, or use an approved decimal representation. Binary floating point is forbidden for exact monetary arithmetic. Define rounding at the business boundary.

### State And Results

Use discriminated unions so each state contains only valid fields. Expected business outcomes may return a result union; exceptional infrastructure failures use typed errors. Keep result alternatives finite and exhaustively handled.

Use a result for an expected alternative the caller must branch on:

```ts
type Result<T, E> =
  | Readonly<{ ok: true; value: T }>
  | Readonly<{ ok: false; error: E }>;

type ReserveFailure =
  | Readonly<{ code: "out_of_stock"; available: number }>
  | Readonly<{ code: "unknown_sku" }>
  | Readonly<{ code: "invalid_quantity" }>;

type Reservation = Readonly<{ sku: string; quantity: number }>;

export function reserve(
  sku: string,
  requested: number,
  available: number | null,
): Result<Reservation, ReserveFailure> {
  if (!Number.isSafeInteger(requested) || requested < 1) {
    return { ok: false, error: { code: "invalid_quantity" } };
  }
  if (available === null) {
    return { ok: false, error: { code: "unknown_sku" } };
  }
  if (requested > available) {
    return { ok: false, error: { code: "out_of_stock", available } };
  }
  return { ok: true, value: { sku, quantity: requested } };
}
```

Throw a typed error when PostgreSQL, the network, or an internal invariant prevents the function from evaluating the expected domain alternatives. Do not wrap every return in `Result`; reserve it for finite, actionable outcomes.

### Boundary Models

Maintain separate wire DTO, database-row, and domain models when their meanings differ. Parse rows and payloads with Zod, then map explicitly. Duplication at a trust boundary is cheaper than coupling domain evolution to storage or transport accidents.

### Evolution

Add fields compatibly: readers tolerate allowed older shapes before writers require new ones. Migrations and API changes follow expand/migrate/contract. Define defaulting in one mapper, not differently across consumers.

## Common Mistakes And Forbidden Patterns

- Reusing a database row or API DTO as the domain type.
- `Partial<T>` as a patch contract without rules for omitted, null, and immutable fields.
- Boolean pairs that permit contradictory state instead of one discriminant.
- Mutable arrays or maps shared across owners.
- Floating-point money or locale-formatted values in storage.
- A branded type created solely through unchecked assertion at the boundary.
- Defaults scattered among route, database, and UI code.

## Verification And Proof

- Construction tests cover valid values and every invariant violation.
- State transition tests prove impossible combinations cannot be produced.
- Serialization and database mapping tests cover omission, null, empty, limits, and unknown fields.
- Typecheck proves exhaustive handling of result and state unions.
- Round-trip tests preserve domain meaning rather than incidental object identity.
- Compatibility tests cover old-reader/new-writer and new-reader/old-writer cases where versions overlap.

Related: [type-system.md](type-system.md), [serialization.md](serialization.md), and [../services/database.md](../services/database.md).
