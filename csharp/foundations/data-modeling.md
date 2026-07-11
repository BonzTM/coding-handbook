# Data Modeling

The per-type decisions a mature .NET team fixes once: how to model enums, identifiers, optional fields, and collections so the type itself enforces correctness.

## Default Approach

Make illegal states unrepresentable, make every reachable default either usable or obviously invalid, and never let a type lie about what it holds. The type system — records, `required` members, nullable reference types, strongly-typed IDs — is the first line of validation; reach for runtime checks only for invariants the type cannot express.

### Records For Immutable Domain Values

Domain values default to `sealed record` with `required` init-only properties. A record gives value equality, non-destructive mutation (`with`), and a shape that cannot be half-mutated after construction; `required` makes the compiler reject an instance missing a field instead of leaving it silently defaulted.

```csharp
public sealed record Address
{
    public required string Street { get; init; }
    public required string City { get; init; }
    public required string PostalCode { get; init; }
    public string? Unit { get; init; } // genuinely optional: absent means "no unit"
}
```

- `required` + `init` is the default property shape. A settable property on a domain type is a decision, not a habit — it invites mutation from anywhere.
- Use `sealed` unless the type is designed for inheritance. Unsealed records get subtle equality behavior across a hierarchy.
- The exception is EF-tracked entities: aggregate roots that EF Core mutates through change tracking are ordinarily classes with encapsulated state, not records. Records make poor tracked entities (value equality fights identity, `with` creates untracked copies). Persistence shape lives in [../services/database.md](../services/database.md).

### Validation At Construction

If a type has invariants beyond "fields are present," enforce them in exactly one place: a constructor or static factory that is the only way to obtain an instance. Once constructed, the value is valid by definition and nothing downstream re-checks it.

```csharp
public sealed record Money
{
    public decimal Amount { get; }
    public Currency Currency { get; }

    private Money(decimal amount, Currency currency) =>
        (Amount, Currency) = (amount, currency);

    public static Money Create(decimal amount, Currency currency)
    {
        if (amount < 0m)
        {
            throw new ArgumentOutOfRangeException(
                nameof(amount), amount, "Amount must be non-negative.");
        }

        if (!currency.IsValid())
        {
            throw new ArgumentException($"Unknown currency '{currency}'.", nameof(currency));
        }

        return new Money(amount, currency);
    }
}
```

The forbidden middle ground is a type that is *silently half-initialized*: public settable properties that must be assigned in an undocumented order, or a parameterless shape that compiles and runs but produces wrong results. Either make the default state work or make it unobtainable — a private constructor behind a validating factory guarantees no caller can build an invalid instance. Request DTOs are validated separately at the trust boundary ([serialization.md](serialization.md)); domain constructors guard *invariants*, not input parsing.

### Enums: Explicit Values, Reserved Zero, Names On The Wire

A C# enum is an integer in a costume: **any** integer converts to it (`(OrderStatus)42` compiles and runs), and an uninitialized or defaulted field is silently `0`. Model around both traps.

```csharp
/// <summary>Lifecycle state of an order. <see cref="Unknown"/> is never valid input and never persisted.</summary>
public enum OrderStatus
{
    Unknown = 0, // the reachable default: unset / invalid, never a legal stored state
    Pending = 1,
    Shipped = 2,
    Cancelled = 3,
}

public static class OrderStatusExtensions
{
    /// <summary>Reports whether the status is a known, usable value.</summary>
    public static bool IsValid(this OrderStatus status) =>
        status is OrderStatus.Pending or OrderStatus.Shipped or OrderStatus.Cancelled;

    /// <summary>Validates external input at the boundary. Names only; numeric strings are rejected.</summary>
    public static bool TryParse(string value, out OrderStatus status)
    {
        status = value switch
        {
            "pending" => OrderStatus.Pending,
            "shipped" => OrderStatus.Shipped,
            "cancelled" => OrderStatus.Cancelled,
            _ => OrderStatus.Unknown,
        };
        return status is not OrderStatus.Unknown;
    }
}
```

- **Pin every value explicitly.** Positional enums renumber when a member is inserted, corrupting every previously persisted row. The explicit number is the storage contract.
- **Reserve `0` for `Unknown`.** `default(OrderStatus)` is always reachable — a missing JSON member, an unassigned field, a zeroed struct. If `0` were `Pending`, a truncated or version-skewed payload would silently become a valid-looking first state instead of an error. `Unknown` is never valid input and never persisted.
- **Names cross the wire, not numbers.** `2` in a log or payload is undebuggable without the source; `"shipped"` is self-documenting. Serialize enums as strings with `JsonStringEnumConverter` — the wire mechanics live in [serialization.md](serialization.md).
- **Guard the boundary explicitly.** `Enum.TryParse` happily accepts numeric strings (`"2"`) and undefined values (`"99"` parses, then `Enum.IsDefined` is your only defense), so boundary parsing matches on names, as above. Never trust a cast integer without an `IsValid` check.
- Additive enum evolution is a compatibility concern; see [contracts-and-compatibility.md](contracts-and-compatibility.md).

### Wrap Identifiers In Strongly-Typed IDs

Identifiers get a `readonly record struct` wrapper so the compiler stops argument-swapping bugs that `Guid` and `long` cannot catch.

```csharp
public readonly record struct OrderId(Guid Value)
{
    public static OrderId New() => new(Guid.CreateVersion7());
    public override string ToString() => Value.ToString();
}

public readonly record struct CustomerId(Guid Value);

// The compiler rejects TransferAsync(orderId, customerId, ...) — args cannot be swapped.
public interface IOrderTransfers
{
    Task TransferAsync(
        CustomerId from, CustomerId to, OrderId order, CancellationToken cancellationToken);
}
```

A method taking `(Guid, Guid, Guid)` invites a caller to pass them in the wrong order; `(CustomerId, CustomerId, OrderId)` makes the mistake a compile error. A `readonly record struct` is allocation-free, value-equal, and usable as a dictionary key. The same applies to any value with units or domain meaning: prefer a wrapped type over a bare primitive.

The cost is at the boundary: a wrapper does not automatically serialize or persist as its inner value. Register a small `JsonConverter` for the wire ([serialization.md](serialization.md)) and an EF Core value converter for the column ([../services/database.md](../services/database.md)) once per ID type, and the rest of the codebase never sees a bare `Guid`.

### ID Generation: Server-Issued UUIDv7 By Default

The server generates identifiers, and the default format is **UUIDv7** — time-ordered, so new rows land at the hot end of a B-tree index instead of scattering like random v4, while staying globally unique and non-guessable. Use `Guid.CreateVersion7()` from the standard library and store the value as the database's native `uuid` type where one exists (persistence rules per [../services/database.md](../services/database.md)).

Client-supplied IDs are acceptable only when the client genuinely owns the identity — an external system's key, or an idempotent create where the caller names the resource. If you accept a client ID, validate it at the boundary and enforce uniqueness in storage so a duplicate becomes a conflict response, not a silent overwrite ([../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md)).

Never expose sequential integers publicly: they are enumerable (an attacker can walk `/orders/1..n`) and they leak business volume. An identity column may still exist internally (e.g. as a keyset-pagination tiebreaker), but the wire-facing identifier is the UUID.

### Optionality Is Nullability, Not Sentinels

Nullable reference types are the optionality model. With `<Nullable>enable</Nullable>` (mandatory, per [project-setup.md](project-setup.md)), every type declaration answers "can this be absent?" and the compiler enforces the answer.

| Model | Use when | Cost |
|---|---|---|
| Non-nullable `string` / `T` | The value is always present; `required` or the constructor guarantees it | None; the default |
| Nullable `string?` / `T?` (reference) | Absence is meaningful in the domain ("no middle name") | A compiler-enforced check at each use |
| Nullable `int?` / `T?` (value type) | A scalar that distinguishes "not set" from a real `0` | An explicit `.HasValue`/pattern check |
| `bool TryGet(... out T value)` or a small result type | A lookup answers "found and value" in one call | A little ceremony at the call site |

The decision rule: if the domain needs "absent," say so in the type with `?`. If it does not, the member is non-nullable and `required` (or constructor-assigned), and nothing downstream ever checks it.

- **No sentinel values.** `""`, `0`, `Guid.Empty`, `DateTimeOffset.MinValue`, and magic negative numbers as "not set" are forbidden — they are valid-looking values that lie, and the compiler cannot see them. C# does not need Go's zero-value conventions; it has real nullability.
- **No `null!` and no `= default!` in domain code.** The null-forgiveness operator disables exactly the check this model depends on. Legitimate uses are rare and confined to infrastructure edges (and each one carries a comment saying why it is safe).
- Nullable database columns map to nullable properties in `Orders.Infrastructure`, and the mapping decides what null means before the value reaches Core — do not let persistence null-ness leak through the domain unexamined.
- Nullable annotations are only compile-time for callers you do not compile: still validate at trust boundaries ([serialization.md](serialization.md)) before input reaches typed domain members.

### Value Versus Reference Semantics

Choose the kind of type by the semantics the domain needs, not by habit.

- **`readonly record struct`** — small immutable values compared by content: IDs, quantities, small measures. Allocation-free, copied on assignment; keep them small (a few fields) so copying stays cheap.
- **`sealed record` (class)** — immutable domain values with several fields: addresses, money, snapshots, events. Value equality, heap-allocated, cheap to pass.
- **`class`** — entities with identity and lifecycle: two `Order` instances with the same ID are the *same* order, so reference identity plus an ID-based equality is correct, and mutation happens through methods that guard invariants. This is the EF-tracked shape.
- Mutable structs are forbidden. A mutable struct copies at every assignment and method boundary, so writes land on copies and vanish — a classic heisenbug. If it mutates, it is a class.
- Equality is part of the choice: records compare by value, classes by reference. Persisting or dictionary-keying a type fixes its equality contract; change it deliberately, never accidentally.

### Collections: Expose Read-Only, Never Null

- **Never return or accept `null` for a collection.** An empty collection *is* the "no results" value: return `[]`, initialize members to empty, and no caller ever writes a null check before `foreach`. A nullable collection type (`List<T>?`) is a modeling error — absence of items is the empty collection.
- **Expose `IReadOnlyList<T>` (or `IReadOnlyCollection<T>`/`IReadOnlyDictionary<K,V>`), not `List<T>`.** A public `List<T>` — even as a get-only property — hands every caller a live handle to your state: they can `Add`/`Clear` behind your back and break your invariants.
- **Copy defensively on the way in.** A constructor that stores a caller-supplied collection without copying stays aliased to it; the caller's later mutation is your corruption.

```csharp
public sealed class Order
{
    private readonly List<OrderLine> _lines;

    public Order(OrderId id, IEnumerable<OrderLine> lines)
    {
        ArgumentNullException.ThrowIfNull(lines);
        Id = id;
        _lines = [.. lines]; // defensive copy: the caller's collection stays theirs
    }

    public OrderId Id { get; }
    public IReadOnlyList<OrderLine> Lines => _lines;

    public void AddLine(OrderLine line)
    {
        ArgumentNullException.ThrowIfNull(line);
        _lines.Add(line); // mutation goes through the method that guards invariants
    }
}
```

- `IReadOnlyList<T>` is a read-only *view*, not an immutable snapshot — the owner can still mutate through `_lines`, which is exactly the encapsulation intended. When callers must hold a stable copy across time, hand out `ImmutableArray<T>` or a fresh array instead.
- Do not expose `IEnumerable<T>` from properties when the backing store is a materialized list; it invites repeated enumeration and hides `Count`. `IEnumerable<T>` is for streams, `IReadOnlyList<T>` for materialized state.
- Dictionary lookups use `TryGetValue`, never index-and-catch; absent keys are a normal outcome, not an exception.

### Immutability At The Boundary

Prefer value objects that callers cannot mutate behind your back. Once an instance is validated by a constructor, keep state private and expose read-only views; produce a new value (`with`) for changes rather than mutating in place. This composes with the collection rule: a "read-only" property returning the internal `List<T>` is not read-only. For genuinely shared mutable state, guard it with a lock inside the owning type rather than exposing the field — see the concurrency rules in [cancellation-and-async.md](cancellation-and-async.md).

Typed durations and instants are part of this discipline: model spans as `TimeSpan` and instants as `DateTimeOffset`, never raw integers or ambiguous `DateTime`. See [time.md](time.md).

## Common Mistakes And Forbidden Patterns

- Enums with positional (implicit) values that renumber on insert, or without a reserved `Unknown = 0`, so a missing or version-skewed payload silently deserializes to a valid-looking first member.
- Persisting or logging enum numbers instead of names, or trusting `(OrderStatus)someInt` / `Enum.TryParse` without a validity guard.
- Primitive-typed identifiers (`Task Transfer(Guid from, Guid to, Guid order)`) that let callers swap arguments with no compile error.
- Sentinel values — `""`, `Guid.Empty`, `DateTimeOffset.MinValue`, `-1` — standing in for "not set" instead of a nullable type.
- `null!` / `= default!` sprinkled through domain code to silence nullable warnings the model should have answered.
- Nullable collections, or returning `null` for "no results" instead of an empty collection.
- Exposing `List<T>` from a public API, or storing a caller's collection without a defensive copy, letting callers mutate private state.
- Mutable structs, or records used as EF-tracked entities where identity and change tracking want a class.
- Types that are silently half-initialized: public setters that must be assigned in an undocumented order, or a reachable default that runs but is wrong.
- Public sequential integer IDs that are enumerable and leak volume.

## Verification And Proof

- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit; nullable violations are compile errors under the gate

Data modeling is done when:

- every persisted or wire-facing enum has explicit pinned values, a reserved `Unknown = 0`, an `IsValid`/`TryParse` guard, a round-trip test (serialize → deserialize → equal), and a test that an unknown value is rejected, not silently accepted.
- identifiers that could be swapped at a call site are distinct `readonly record struct` types, the swap is a compile error, and each ID type has registered JSON and EF converters with round-trip tests.
- every "might be absent" member is expressed as a nullable annotation justified by domain meaning — zero sentinel values, zero `null!` in domain code.
- every type with invariants is unobtainable in an invalid state (validating constructor or factory), with tests asserting the rejection path.
- no public member returns `List<T>`, `null` collections, or an aliased internal collection; ownership tests prove a caller's post-construction mutation does not leak in.
- entity vs value semantics (class vs record vs record struct) are deliberate per type, and equality behavior is covered where it feeds dictionaries, sets, or persistence.

## Where To Go Next

- [serialization.md](serialization.md) — wire converters for typed IDs and enums, absent-vs-null, and the DTO boundary.
- [contracts-and-compatibility.md](contracts-and-compatibility.md) — evolving enums and payload shapes without breaking consumers.
- [style-and-review.md](style-and-review.md) — immutability idioms, defensive copies, and the review heuristics this doc deepens.
- [time.md](time.md) — modeling spans and instants with `TimeSpan` and `DateTimeOffset`.
- [../services/database.md](../services/database.md) — EF Core value converters, entity shape, and nullable column mapping.
