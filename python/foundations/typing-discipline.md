# Typing Discipline

Static typing rules that make Python boundaries explicit and keep dynamic escape hatches visible.

## Default Approach

Mypy strict mode is the repository gate. Type every public and cross-layer boundary; let inference handle obvious local expressions. Typing documents the runtime contract but never replaces validation of external input.

### Strict Means A Defined Gate

Set `strict = true`, `warn_unreachable = true`, and `disallow_any_unimported = true` in `[tool.mypy]`. According to the current [mypy command-line reference](https://mypy.readthedocs.io/en/stable/command_line.html#cmdoption-mypy-strict), strict enables:

- `disallow_any_generics`
- `disallow_subclassing_any`
- `disallow_untyped_calls`
- `disallow_untyped_defs`
- `disallow_incomplete_defs`
- `check_untyped_defs`
- `disallow_untyped_decorators`
- `warn_redundant_casts`
- `warn_unused_ignores`
- `warn_return_any`
- `no_implicit_reexport`
- `strict_equality`
- `extra_checks`

Mypy explicitly notes that the strict set may change. Pin the tool through the lockfile and review that list on upgrades. `warn_unreachable` is not included by strict; the additional `disallow_any_unimported` rejects types degraded to `Any` through unresolved imports.

### Optional Is Explicit

`T | None` means absence is part of the contract. Do not use `None` as a placeholder for incomplete initialization or silence an initialization problem with an optional attribute. Narrow with a guard before use. Default arguments that accept `None` declare it in the type.

Use a private sentinel when omitted, explicit `None`, and a real value have distinct meanings:

```python
_UNSET = object()

def update_name(value: str | None | object = _UNSET) -> None: ...
```

Prefer a dedicated enum/sentinel type when that value crosses a module boundary; raw `object` unions are internal implementation details.

### Any Is A Boundary Defect

Public functions, decorators, ports, DTO mappings, parsed configuration, database results, and client responses expose no implicit `Any`. Validate unknown data as `object`, narrow it, then construct a typed value. A third-party package with missing or weak typing requires a narrow adapter or reviewed stub, not global `ignore_missing_imports`.

### Protocols Own Behavioral Seams

Use consumer-defined `Protocol` for ports and test seams. Use an ABC only when runtime inheritance supplies shared implementation, registration, or an explicit nominal identity. Keep protocols narrow and place them in the consuming core module; see [package design](package-design.md).

### Distinct IDs Use NewType

Use `NewType` when same-representation identifiers can be swapped accidentally:

```python
WidgetId = NewType("WidgetId", str)
TenantId = NewType("TenantId", str)
```

Construct them after boundary validation. `NewType` distinguishes values statically without adding runtime validation; Pydantic/DTO and repository mappings remain responsible for parsing.

### Shape Selection

| Shape | Use when | Do not use for |
|---|---|---|
| `TypedDict` | a typed dictionary shape required by a dictionary-based API | domain invariants or runtime validation |
| dataclass | an owned domain value/entity with behavior and construction rules | untrusted input parsing |
| Pydantic model | HTTP/config/message trust boundary requiring runtime parsing | framework-free domain core |
| Protocol | behavior a consumer requires from an adapter | shared implementation |

Do not pass a `TypedDict` through unrelated layers because it resembles JSON. Map boundary data once into domain values.

### Generics And Self

Use type parameters for containers and algorithms that preserve or relate caller types. Write the concrete implementation first; generalize when a second real use removes duplication. Bound type variables to the operations required. Use `Self` for fluent methods and alternative constructors that return the concrete subclass; do not use it when the method deliberately returns a base abstraction.

### Decorators Preserve Signatures

A decorator uses `ParamSpec` plus a return type variable when it forwards arbitrary parameters, or a `Protocol`/explicit callable signature when it accepts a narrower shape. `Callable[..., Any]` is forbidden at typed boundaries. Use `functools.wraps`; an untyped decorator poisons every decorated function under strict mode.

### Suppressions Are Local Evidence

`cast(T, value)` requires an adjacent comment naming the invariant or deficient third-party type information the checker cannot see. It performs no runtime check; never use it to replace parsing.

Every `# type: ignore[error-code]` includes the narrow mypy code and a justification on the same or preceding line. Bare ignores and file-wide suppression are forbidden. Remove the ignore when the upstream typing defect is fixed; `warn_unused_ignores` makes stale suppressions fail.

### Libraries Ship Typing Metadata

Published typed libraries include a `py.typed` marker in the built wheel as specified by [PEP 561](https://peps.python.org/pep-0561/). Public APIs, re-exports, overloads, and package data are verified from the installed artifact, not only the checkout.

## Common Mistakes And Forbidden Patterns

- `Any`, bare containers, or untyped decorators at public/cross-layer boundaries.
- `Optional` used to defer required initialization or make every field nullable.
- Global `ignore_missing_imports`, bare `type: ignore`, or a suppression with no error code and reason.
- `cast()` used as validation or without stating the hidden invariant.
- An ABC for every adapter, or Protocols defined beside implementations instead of consumers.
- String IDs shared across distinct domains without `NewType`.
- Pydantic models treated as domain entities; `TypedDict` transported through the core.
- A library wheel missing `py.typed` or silently dropping public re-exports.

## Verification And Proof

```bash
uv run mypy .
uv build
make verify
```

Typing is done when strict mypy passes without broad exclusions; every suppression is narrow and justified; negative type fixtures prove IDs cannot be swapped where used; decorator signatures remain visible; and an installed library artifact exposes `py.typed` and its intended public API.

Related: [data modeling](data-modeling.md), [serialization](serialization.md), and [testing](../quality/testing.md).
