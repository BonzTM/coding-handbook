# Serialization

The JSON encoding boundary: how wire shapes are declared, evolved, and proven, so the contract never drifts with an internal refactor.

## Default Approach

The wire format is a contract, not a reflection of your structs. Declare it explicitly on dedicated transport types, control omission and precision deliberately, and keep all encoding logic at the boundary. The default codec is the standard library `encoding/json`.

This doc governs the JSON boundary. For the shape of the data behind it see [data-modeling.md](data-modeling.md); for the rules that make a shape safe to change over time see [contracts-and-compatibility.md](contracts-and-compatibility.md), which mandates "stable request/response shapes" and "new optional fields" as the evolution mode — both rest on the tag and omission conventions specified here.

### Always Write Explicit json Tags

Every serialized field carries an explicit `json:` tag. Never rely on Go's field-name default; never leave the key to the whim of a rename.

```go
type CreateOrderResponse struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
```

- The default key is the Go field name verbatim (`OrderID` → `"OrderID"`), which leaks Go naming into the contract and silently changes the wire shape on a field rename. An explicit tag decouples the two: rename the field freely; the wire key stays put.
- Pick one case convention per repo and document it. `snake_case` is the default for new JSON surfaces unless the team has a documented `camelCase` standard (match an existing external API, a mobile client, or a published spec). Apply it uniformly; a mixed-case payload is a bug.
- Unexported fields never serialize — `encoding/json` ignores them entirely. A field that must cross the wire is exported and tagged. A field that must not cross the wire is unexported, or tagged `json:"-"`.
- Tag every field, including ones you intend to drop, so the omission is a visible decision (`json:"-"`) rather than an accident.

### omitempty Is A Trap; Reach For omitzero

`omitempty` does not mean "omit when zero." It omits only empty values as the v1 codec defines empty: `false`, `0`, `""`, nil pointers/interfaces, and empty maps/slices/arrays. It does **not** omit a zero-value struct, and it does **not** omit a zero `time.Time` — a long-standing source of friction where `"created_at":"0001-01-01T00:00:00Z"` leaks into payloads.

It also conflates two distinct facts for scalars: a field genuinely set to `0`/`""` and a field that was never set both disappear. That is wrong whenever zero is a meaningful value.

Go 1.24 added the `omitzero` tag option, which is the modern default when the intent is "omit when zero." Per the Go 1.24 release notes:

> a struct field with the new `omitzero` option in the struct field tag will be omitted if its value is zero. If the field type has an `IsZero() bool` method, that will be used to determine whether the value is zero ... unlike `omitempty`, `omitzero` omits zero-valued `time.Time` values, which is a common source of friction.

```go
type Event struct {
	Name      string    `json:"name"`
	Note      string    `json:"note,omitzero"`       // dropped when ""
	StartedAt time.Time `json:"started_at,omitzero"` // dropped when zero (uses time.Time.IsZero)
	Window    Span      `json:"window,omitzero"`     // dropped when the struct is zero
}
```

- Use `omitzero` (not `omitempty`) wherever the rule is "omit when zero," because it omits zero structs and zero `time.Time` and uses a type's `IsZero()` method.
- Omitting a field is not the same as encoding `null`. If the contract needs a true tri-state — **absent** vs **null** vs **value** — use a pointer (`*string`, `*int64`) and let `nil` map to absent (with `omitzero`/`omitempty`) or to explicit `null`. A pointer is the only stdlib way to distinguish "client did not send this field" from "client sent zero." Decide which the contract requires and prove it with a round-trip test.
- Do not sprinkle `omitempty`/`omitzero` reflexively. On a required response field, omission hides a bug; emit the zero value so consumers see the contract is intact.

### Numeric Precision: Strings For Large Ints, Never Float For Money

JSON has one numeric type. JavaScript and many JSON tooling stacks decode every number into an IEEE-754 `float64`, which represents integers exactly only up to 2^53. An `int64` ID or counter above that silently corrupts on a JS client.

- Serialize `int64` IDs, counters, and any value that can exceed 2^53 as JSON **strings** when the consumer is or may be a JS/JSON client. Use a dedicated string-int boundary type (below) or `json:",string"` for simple cases, and document the choice in the contract.
- **Never represent money as a float.** Floating point cannot hold `0.10` exactly; sums drift. Use integer **minor units** (cents, satoshis) carried as an `int64` (stringified per the rule above) plus an explicit currency code, or a dedicated decimal type. The decimal-library decision is deferred to [../decisions/framework-selection.md](../decisions/framework-selection.md) — state the need there; do not import one ad hoc.
- A bare `float64` is acceptable only for genuinely approximate quantities (ratios, scores, measurements) where lost precision is harmless and documented.

### Custom Marshalers Live On Boundary Types Only

Implement `json.Marshaler`/`json.Unmarshaler` (or `encoding.TextMarshaler`/`TextUnmarshaler`) only on small, dedicated wire types whose sole job is the encoding. Keep each one tiny and unit-tested both directions.

```go
// StringInt64 marshals an int64 as a JSON string to survive float64 clients.
type StringInt64 int64

func (v StringInt64) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(strconv.FormatInt(int64(v), 10))), nil
}

func (v *StringInt64) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return fmt.Errorf("stringint64: %w", err)
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("stringint64: %w", err)
	}
	*v = StringInt64(n)
	return nil
}
```

- Never put encoding logic in a domain or core type. A `MarshalJSON` on a business entity welds the wire format to the domain model and defeats the DTO boundary below.
- A custom unmarshaler must wrap its errors with `%w` (see [errors-and-logging.md](errors-and-logging.md)) and validate, not just parse.

### Separate DTOs From Domain Types

Wire structs (DTOs) are defined in the transport package; they are not your domain or `internal/core` types. Map between them explicitly at the boundary.

- Define request/response structs alongside the handler in `internal/api/...`, per [../services/http-services.md](../services/http-services.md) and [../recipes/add-http-endpoint.md](../recipes/add-http-endpoint.md).
- Serializing a domain struct directly couples the public wire contract to internal refactors: rename a core field and you silently break clients, or you freeze the domain model to protect the wire. The DTO breaks that coupling — the mapping function is the one place the two shapes meet.
- The mapping is plain code (`func toResponse(o core.Order) CreateOrderResponse`), tested as part of the handler. Do not reach for reflection-based struct copiers; they reintroduce the implicit coupling the DTO removed.

### Decode Defensively

Untrusted input is bounded, parsed strictly per policy, then validated — in that order.

```go
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var v T
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB cap
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // strict inputs only; see policy below
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("decode %T: %w", v, err)
	}
	return v, nil
}
```

- Bound the request body with `http.MaxBytesReader` before decoding. An unbounded `json.Decode` will read an attacker-sized body into memory. Set the cap per endpoint.
- Choose an unknown-field policy and apply it consistently. For strict internal or versioned APIs, call `DisallowUnknownFields()` so typos and stale clients fail loudly. For public APIs where additive evolution must not break old senders, **ignore** unknown fields (the default) — this is what lets a new optional field ship without coordinating every caller, per [contracts-and-compatibility.md](contracts-and-compatibility.md). Document which mode each surface uses.
- Decoding never validates business rules. After a successful decode, run validation (required fields present, ranges, enum membership) and reject with a structured error response. See [../services/http-services.md](../services/http-services.md) for the error model.

### Error Responses

A failure is a wire shape too. Every endpoint returns the **same** structured error envelope; a bare `{"error":"..."}` string is not a contract — it gives the client one human sentence and nothing to branch on. The default shape is a stable top-level object:

```go
// ErrorResponse is the single error envelope returned by every endpoint.
type ErrorResponse struct {
	Code      string       `json:"code"`                // machine-readable enum, not the HTTP status
	Message   string       `json:"message"`             // human-readable, safe to surface
	Fields    []FieldError `json:"fields,omitzero"`     // per-field validation failures
	RequestID string       `json:"request_id,omitzero"` // correlation id; see errors-and-logging.md
}

// FieldError carries one validation failure against one input field.
type FieldError struct {
	Field   string `json:"field"`   // dotted path into the request: "items.0.qty"
	Code    string `json:"code"`    // machine-readable enum: "required", "out_of_range"
	Message string `json:"message"` // human-readable detail
}
```

```json
{
  "code": "validation_failed",
  "message": "the request has invalid fields",
  "fields": [
    { "field": "items.0.qty", "code": "out_of_range", "message": "must be >= 1" }
  ],
  "request_id": "01HZ..."
}
```

- The `code` is a **string enum that the client may branch on** — `"validation_failed"`, `"order_not_found"`, `"rate_limited"` — and it is **not** the HTTP status. The status is the coarse channel (`4xx` vs `5xx`); the `code` is the precise one. Two `404`s with different `code`s are different failures; a client that switches on the status alone cannot tell them apart, so never make the status the only error signal.
- The `code` set is a documented part of the wire contract and evolves under the same rules as any other field — add codes additively, never repurpose or silently drop one — per [contracts-and-compatibility.md](contracts-and-compatibility.md). A new `code` is a new optional behavior; removing or re-meaning one is a breaking change.
- Validation failures populate `fields`, one `FieldError` per offending input, so a client can attach messages to form fields without parsing prose. The `field` is a dotted path into the request body; the per-field `code` is its own small enum.
- **5xx is opaque.** Never put internal error text, a wrapped chain, or a stack trace in `message`. Server faults return a generic message (`"internal error"`) plus the `request_id`; the detail goes to the logs keyed by that id (see [errors-and-logging.md](errors-and-logging.md)), where an operator can find it. The client gets a correlation handle, not your internals.
- The mapping from a domain error to `(status, code)` happens once, at the boundary. The handler does not hand-build envelopes inline; it calls one helper that classifies the error — using `errors.Is`/`errors.As` against the sentinel and category types from [errors-and-logging.md](errors-and-logging.md) — and writes the envelope. See [../services/http-services.md](../services/http-services.md) for where that helper sits in the handler and middleware order. A 5xx maps to a generic code and drops detail; a known domain error maps to its documented `(status, code)`.

### Codec Choice: encoding/json (v1) Is The Default

Use the standard library `encoding/json`. It is the mandate.

A second implementation, `encoding/json/v2`, exists but is experimental. Per its package docs: "This package (encoding/json/v2) is experimental, and not subject to the Go 1 compatibility promise. It only exists when building with the GOEXPERIMENT=jsonv2 environment variable set. Most users should use encoding/json." Treat v2 as forward-looking: do not build the contract on a `GOEXPERIMENT` flag, and do not adopt it until it ships as a stable, default-on API. Track it; do not depend on it.

## Common Mistakes And Forbidden Patterns

- Relying on Go field-name defaults instead of writing explicit `json:` tags; a rename then silently changes the wire key.
- Using `omitempty` on a struct or `time.Time` expecting it to disappear — it does not; use `omitzero`.
- Using `omitempty`/`omitzero` to fake a tri-state when the contract needs absent-vs-null-vs-value; use a pointer.
- Representing money as `float64`, or relying on float for any value where precision matters.
- Sending `int64` IDs/counters as JSON numbers to a JS/JSON client; they corrupt above 2^53. Stringify them.
- Putting `MarshalJSON`/`UnmarshalJSON` on a domain type, or serializing `internal/core` structs directly on the wire.
- Decoding an unbounded request body, or leaving the unknown-field policy unstated and inconsistent.
- Building on `encoding/json/v2` behind `GOEXPERIMENT=jsonv2` as if it were stable.
- Returning a bare `{"error":"..."}` string (or a raw `http.Error` line) instead of the structured envelope; the client gets prose and nothing to branch on.
- Leaking internal error text, a wrapped `%w` chain, or a stack trace into a 5xx `message`; the body must be opaque, with detail in the logs under the `request_id`.
- Inventing a per-endpoint, ad-hoc error shape; every endpoint returns the one `ErrorResponse` envelope.
- Treating the HTTP status as the only error signal — omitting the machine-readable `code`, so two distinct `404`s are indistinguishable to the client.

## Verification And Proof

```bash
make verify   # the full gate; runs vet, test, and race
```

The serialization boundary is done when:

- every wire type has a **round-trip test** (`Marshal` then `Unmarshal` yields an equal value) and a **golden test** asserting exact bytes/keys, so a tag or case change is caught as a diff.
- decode tests prove the unknown-field policy for each surface: strict surfaces **reject** an unknown field; additive surfaces **ignore** it without error.
- precision tests cover large `int64` IDs (> 2^53) surviving a marshal/unmarshal round trip, and prove money values use integer minor units or a decimal type — never `float64`.
- absent-vs-null-vs-value behavior is asserted for every tri-state field (pointer present, `null`, and omitted all decode distinctly).
- a body-size cap is enforced and tested (oversized body is rejected, not OOM'd).
- a **golden test pins the error envelope** — exact keys, the `code` value, and the `fields` shape — so a thinning or rename of the error contract shows up as a diff, the same way response goldens do.
- a validation-failure test asserts the response carries field-level `fields` entries (correct `field` path and per-field `code`), not just a top-level message.
- a 5xx test asserts the body carries **no internal detail** — generic `code`/`message`, a `request_id` present, and none of the wrapped error text — while the detail is confirmed in the log line for that id.

## Where To Go Next

- [data-modeling.md](data-modeling.md) — the domain shapes behind the DTO, and where validation lives.
- [contracts-and-compatibility.md](contracts-and-compatibility.md) — why stable shapes and additive-only evolution are the rule this doc implements.
- [../services/http-services.md](../services/http-services.md) — where DTOs, decoding, and the error model live in a handler.
- [../recipes/add-http-endpoint.md](../recipes/add-http-endpoint.md) — the end-to-end recipe that wires a DTO to a route.
- [../decisions/framework-selection.md](../decisions/framework-selection.md) — the decimal-library stance for money.
