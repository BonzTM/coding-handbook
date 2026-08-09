# Data Modeling

Per-type decisions for identifiers, values, entities, optional fields, and collections so domain code carries its own invariants.

## Default Approach

Make invalid state difficult to construct. Parse untrusted input at the adapter, then create plain typed domain values that do not depend on Pydantic, FastAPI, SQLAlchemy, or transport representation.

### Value Objects Are Frozen Dataclasses

Use `@dataclass(frozen=True, slots=True)` by default for domain values. Frozen state prevents ordinary reassignment; slots rejects accidental attributes and keeps the instance shape explicit. Validate invariants in `__post_init__` for direct construction or expose a classmethod factory when parsing/conversion can fail in multiple ways.

```python
@dataclass(frozen=True, slots=True)
class Money:
    minor_units: int
    currency: Currency

    def __post_init__(self) -> None:
        if self.minor_units < 0:
            raise ValueError("minor_units must be non-negative")
```

Do not perform I/O in construction. If an invariant needs a database or network call, it belongs in a use case that constructs the value after obtaining evidence.

### Values Versus Entities

A value object is identified by all its fields and is replaced rather than mutated. An entity has stable identity across state changes; equality and persistence semantics follow that identity deliberately. Do not let a database row object define the domain's equality, lifecycle, or mutation rules.

Entities may use a dataclass but keep mutation behind named methods that enforce transitions. Return new values for snapshots crossing task or layer boundaries.

### Enums Are Stable Values

Use `Enum` for internal finite states. Use `StrEnum` for wire-adjacent values whose stable string representation is part of a contract. Assign every member value explicitly; never derive a persisted value from member order or `auto()`.

Unknown external enum values are handled at the boundary according to the compatibility policy: inbound commands reject them; tolerant external consumers preserve or map them without pretending they are known. Adding an enum member is a compatibility review, not merely a code edit.

### Identifiers Are Distinct Types

Use `NewType` for identifiers with the same runtime representation but different meaning:

```python
OrderId = NewType("OrderId", str)
CustomerId = NewType("CustomerId", str)
```

Validate format/non-emptiness before construction. Do not build a dataclass wrapper solely to make an ID distinct unless it also owns real behavior or invariants that justify runtime identity.

### Construction Enforces Invariants

Use `__post_init__` for cheap, deterministic checks over already typed fields. Use a named factory such as `Sku.parse(raw: str) -> Self` when normalization and parse failures are part of the contract. Raise specific domain construction errors when callers branch; use `ValueError` for programmer-facing primitive invariant failures that never cross the wire unchanged.

A valid object never has a partially initialized phase. Avoid setters that require fields to be assigned in a particular order.

Factories return a complete value or raise one documented construction error; they never return a partially useful object alongside an error flag.

### Optional Values And Sentinels

Use `T | None` only when absence is a domain value. If `None` is itself meaningful and omission differs from explicit null, use a private typed sentinel at the boundary and map it into a domain command that represents the three states explicitly. Do not overload empty strings, zero, or false to mean absent unless the domain makes them identical.

| Model | Use when | Cost |
|---|---|---|
| `T | None` | value or genuine absence | every consumer must narrow |
| explicit enum/value state | absence has domain behavior | extra type, clearest contract |
| private sentinel | omitted/null/value differ during boundary parsing | must not leak beyond mapping |
| default value | omitted and default are genuinely equivalent | loses presence information |

### Collections Are Immutable At Boundaries

Expose `tuple[T, ...]`, `frozenset[T]`, or `Mapping[K, V]` from domain contracts when callers do not own mutation. Copy incoming mutable collections before retaining them. A frozen dataclass containing a list is still transitively mutable; frozen means assignment protection, not deep immutability.

Use `field(default_factory=list)` or another factory for owned mutable dataclass fields. Mutable default arguments are forbidden; Ruff `B006` enforces the common case. Return copies or immutable views rather than internal lists/dicts.

### Money Is Integer Minor Units

Represent money as integer minor units plus an explicit currency code. Never use binary `float` for monetary values. Define rounding at the conversion boundary and test it. Escalate to `decimal.Decimal` only when fractional minor units, accounting rules, or a contract require decimal arithmetic; set and localize the rounding/context policy.

### Domain And Persistence Shapes Stay Separate

SQLAlchemy models and rows map explicitly to domain objects in repositories. Pydantic DTOs map explicitly in transport adapters. Do not add serialization aliases, ORM session state, or validation-framework methods to a domain type for adapter convenience.

## Common Mistakes And Forbidden Patterns

- Mutable domain values shared across callers or tasks.
- A frozen dataclass containing exposed mutable collections and treated as deeply immutable.
- Mutable default arguments or dataclass fields initialized with one shared list/dict.
- `auto()`/ordinal enum values persisted or placed on the wire.
- Stringly typed IDs that callers can swap without a type error.
- `None`, empty string, and omitted conflated without an explicit contract.
- Construction that permits a half-valid object or performs I/O.
- Money represented as `float`, or currency omitted from a monetary value.
- Pydantic/SQLAlchemy models crossing into core as domain entities.

## Verification And Proof

```bash
uv run ruff check .
uv run mypy .
uv run pytest -k "model or domain or value"
make verify
```

Data modeling is done when constructors prove valid and invalid paths; equality/identity semantics are tested; enum and identifier mappings round-trip; optional/omitted/null states are explicit; mutation cannot escape through collections; money rounding is tested; and strict mypy rejects swapped identifiers in the relevant type test.

Related: [typing discipline](typing-discipline.md), [serialization](serialization.md), and [time](time.md).
