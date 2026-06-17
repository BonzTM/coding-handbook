# Data Modeling

The per-type decisions a mature Go team fixes once: how to model enums, identifiers, optional fields, and collections so the type itself enforces correctness.

## Default Approach

Make illegal states unrepresentable, make the zero value either usable or obviously invalid, and never let a type lie about what it holds. The type is the first line of validation; reach for runtime checks only for invariants the type cannot express.

### Enums Are Defined String Types, Not Bare iota

Default to a defined string type with typed constant values for any enum that crosses a boundary — JSON, a database column, a queue payload, a log line, an HTTP query parameter.

```go
// Status is the lifecycle state of an order. The zero value is StatusUnknown,
// which is never valid input and never persisted.
type Status string

const (
	StatusUnknown   Status = ""        // zero value: unset / invalid
	StatusPending   Status = "pending"
	StatusShipped   Status = "shipped"
	StatusCancelled Status = "cancelled"
)

func (s Status) String() string { return string(s) }

// Valid reports whether s is a known, usable status.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusShipped, StatusCancelled:
		return true
	default:
		return false
	}
}

// ParseStatus validates external input at the boundary.
func ParseStatus(v string) (Status, error) {
	s := Status(v)
	if !s.Valid() {
		return StatusUnknown, fmt.Errorf("parse status: unknown value %q", v)
	}
	return s, nil
}
```

Every persisted or wire-facing enum owes the team three things: a `String()` method, a `Valid()` (or `Parse`) function that rejects unknown values, and an explicit zero value that means "unknown/unset" and is never a legal stored state.

Use `iota` only for sets that are internal, non-persisted, and contiguous — a state machine that never leaves the process, a flag table, a switch tag that is never marshaled. The moment an `iota` enum touches JSON, a DB column, or a wire format, it becomes fragile:

- The number carries no meaning. `2` in a log or a database row is undebuggable without the source; `"shipped"` is self-documenting.
- Inserting or reordering a constant silently renumbers everything after it, corrupting every previously persisted row. Defined-string values are stable because the value is the name.
- A missing value deserializes to `0`, which is a *valid-looking* enum member, so a truncated or version-skewed payload silently becomes the first state instead of an error.

If an `iota` enum must be persisted anyway (rare, e.g. a hot column where bytes matter), pin each value explicitly (`StatusPending Status = 1`), never rely on positional `iota`, and still provide `String()` and a parse guard. Additive enum evolution is a compatibility concern; see [contracts-and-compatibility.md](contracts-and-compatibility.md).

### Wrap Identifiers In Named Types

Identifiers get a named type so the compiler stops argument-swapping bugs that `string` and `int64` cannot catch.

```go
type UserID string
type OrderID string

// The compiler rejects Transfer(orderID, userID) — the args cannot be swapped.
func Transfer(ctx context.Context, from, to UserID, order OrderID) error
```

A function taking `(string, string, string)` invites a caller to pass them in the wrong order; `(UserID, UserID, OrderID)` makes the mistake a compile error. The same applies to any value with units or domain meaning: prefer a named type over a bare primitive.

The cost is at the boundary: a named type does not automatically marshal or scan the way you may expect, and crossing JSON, SQL, or flags needs deliberate handling. Keep the underlying type a plain `string`/`int64` (not a struct) so the conversion stays cheap, and centralize marshal/scan behavior on the type. See [serialization.md](serialization.md) for the round-trip rules.

### Optional Fields: Choose By Meaning, Not Habit

There are four ways to model "this field might not have a value." Pick by what *absent* means in the domain, not by reflex. Defaulting everything to `*T` is a forbidden pattern — it sprays nil checks across the codebase and pushes panics to runtime.

| Model | Use when | Cost |
|---|---|---|
| Zero value as absent (`""`, `0`, `false`) | The zero value is genuinely indistinguishable from "not set" *and* the domain treats them the same | None; simplest |
| Pointer `*T` | You must distinguish "set to zero" from "not set" (a real `0` vs no value), e.g. a settable count or a patch payload | Heap/nil-check tax; every read is a guard |
| `sql.Null[T]` / typed null | The value comes from or goes to a nullable DB column | Explicit `.Valid` checks; DB-shaped, not domain-shaped |
| Explicit `(T, bool)` return or a small `Optional[T]` | A function answers "found and value" in one go (map lookup shape) | A little ceremony at the call site |

The decision rule: if the domain genuinely needs to tell "absent" apart from "zero," you must encode presence explicitly — `*T`, `sql.Null[T]`, or an ok-bool. If "absent" and "zero" mean the same thing to the business, use the zero value and keep the type flat. Do not use a pointer merely to make a field "optional-looking" when zero already says everything. When you do reach for `*T`, let it carry only the optional contract; a pointer that also doubles as shared-mutable state conflates two meanings (see [style-and-review.md](style-and-review.md#copying-and-immutability)).

`sql.Null[T]` (Go 1.22+) is the default for nullable database columns; it scans and values cleanly and keeps the null-ness at the persistence edge instead of leaking pointers into the domain. Map the DB-null type to a domain shape (zero value or `*T`) in the repository layer, not throughout core code.

Nil-versus-empty is its own decision and shows up at the JSON boundary: a nil slice and an empty slice both range as zero elements in Go, but marshal differently (`null` vs `[]`), and that difference is a contract. Decide it deliberately per field; the marshaling rules and how to force one or the other live in [serialization.md](serialization.md).

### Design The Zero Value: Usable Or Invalid By Construction

Every struct's zero value is reachable — callers can write `var x T` or get one from a map miss. Two acceptable designs; never a third:

- **Usable zero value.** `bytes.Buffer{}`, `sync.Mutex{}`, and `slog.Logger`-style types work immediately with no constructor. Prefer this when you can; it makes the type composable and zero-value-friendly (the idiom [style-and-review.md](style-and-review.md) gestures at). A zero `time.Duration` is `0`, a zero `Status` is `StatusUnknown` — meaningful by design.
- **Invalid by construction.** If the type needs dependencies or invariants (an open connection, a validated config, an injected clock), make the zero value clearly unusable and force a constructor: `New...` returns the only valid instances, validates inputs, and returns an error if invariants cannot hold.

```go
// Account requires a non-empty owner and a valid currency; the zero Account is unusable.
type Account struct {
	id    AccountID
	owner UserID
	cur   Currency
}

func NewAccount(id AccountID, owner UserID, cur Currency) (*Account, error) {
	if id == "" || owner == "" {
		return nil, errors.New("new account: id and owner are required")
	}
	if !cur.Valid() {
		return nil, fmt.Errorf("new account: invalid currency %q", cur)
	}
	return &Account{id: id, owner: owner, cur: cur}, nil
}
```

The forbidden middle ground is a type that is *silently half-initialized*: a zero value that compiles and runs but produces wrong results, or a struct with public fields that callers must set in a particular undocumented order. Either make zero work or make zero refuse to work. Unexported fields behind a constructor are how you guarantee no caller can build an invalid instance.

### Slices And Maps: Know The Sharp Edges

These are reference-like types with aliasing and ordering behavior that bites teams who model with them casually.

- **nil vs empty.** A `nil` slice and an empty non-nil slice behave identically for `len`, `range`, and `append`; do not write code that depends on telling them apart in Go. The only place the difference is observable is serialization (`null` vs `[]`) — see [serialization.md](serialization.md). Return `nil` for "no results," not a special empty marker.
- **append aliasing.** `append` may return a slice that shares the original backing array, so mutating the result can corrupt the source (and vice versa). When you store a caller-provided slice, or hand a sub-slice to code that may append, copy defensively. Use the three-index slice (`s[a:b:b]`) to cap capacity and force the next `append` to allocate. This is the classic source of `-race` failures and spooky-action bugs.
- **map iteration is randomized.** Go deliberately randomizes map range order. Never depend on it for output, hashing, or equality. When you need deterministic output, collect keys into a slice and `slices.Sort` them. A test that passes only because a map happened to iterate in one order is already broken.
- **pre-size when length is known.** `make([]T, 0, n)` and `make(map[K]V, n)` when `n` is known up front avoids repeated reallocation and rehashing. Cheap, and it documents intent.
- **returning internals leaks mutable state.** Returning a struct's internal slice or map (or accepting one and storing it without copying) hands callers a live handle to your state; they can mutate it behind your back and break your invariants. Either return a copy (`slices.Clone`, `maps.Clone`) or document loudly that the returned value is owned by the caller and must not be retained.

### Immutability At The Boundary

Prefer value objects that callers cannot mutate behind your back. Once an instance is validated by a constructor, keep its fields unexported and expose read-only accessors; produce a new value for changes rather than mutating in place. This composes with the slice/map rule: a "read-only" getter that returns the internal slice is not read-only. For genuinely shared mutable state, guard it with a mutex inside the type rather than exposing the field — and remember a type containing a `sync.Mutex` must not be copied (see the copylocks guidance in [style-and-review.md](style-and-review.md)).

Typed durations and instants are part of this discipline: model spans as `time.Duration` and instants as `time.Time`, never raw integers. See [time.md](time.md).

## Common Mistakes And Forbidden Patterns

- Persisting or serializing a bare `iota` enum, then reordering or inserting a constant and silently corrupting every stored row.
- Enums with no `String()`, no `Valid`/`Parse` guard, or no explicit unknown zero value, so a truncated payload deserializes to a valid-looking first member.
- Stringly-typed identifiers (`func Transfer(from, to, order string)`) that let callers swap arguments with no compile error.
- Reaching for `*T` on every optional field with no decision rule, spraying nil checks and runtime panics across the code.
- Confusing nil and empty: depending on the distinction inside Go logic, or ignoring that it surfaces as `null` vs `[]` at the wire boundary.
- Returning a struct's internal slice or map (or storing a caller's without copying), letting callers mutate private state.
- Depending on map iteration order for output, equality, or hashing.
- Types that are silently half-initialized: a zero value that compiles and runs but is wrong, or public fields that must be set in an undocumented order.

## Verification And Proof

```bash
make verify   # vet + test + race; the full gate
make race     # shared-slice / shared-map aliasing on its own
```

Data modeling is done when:

- every persisted or wire-facing enum has a `String()`, a `Valid`/`Parse` guard, and an explicit unknown zero value, with a round-trip test (`marshal -> unmarshal -> Equal`) and a test that an unknown value is rejected, not silently accepted.
- identifiers that could be swapped at a call site are distinct named types, and the swap is a compile error.
- each optional field has a deliberate model (zero / `*T` / `sql.Null[T]` / ok-bool) justified by the domain meaning of "absent," not chosen by habit.
- the zero value of every exported type is either usable or refused by a constructor; constructor invariant enforcement is covered by tests asserting the error path.
- `go test -race` is clean, including tests that exercise shared-slice/append aliasing and returned-collection ownership.
- no test depends on map iteration order; deterministic output sorts keys.

## Where To Go Next

- [serialization.md](serialization.md) — marshaling named types, nil-vs-empty on the wire, custom `MarshalJSON`/`Scan`/`Value`.
- [contracts-and-compatibility.md](contracts-and-compatibility.md) — evolving enums and payload shapes without breaking consumers.
- [style-and-review.md](style-and-review.md) — value/pointer semantics, copylocks, and the zero-value-friendly baseline this doc deepens.
- [time.md](time.md) — modeling spans and instants with `time.Duration` and `time.Time`.
