# Package Design

Package boundaries, visibility rules, and dependency direction for Python code that stays maintainable under growth.

## Default Approach

Prefer one distribution with a small, explicit import graph. Folder names communicate ownership; Import Linter makes the dependency direction executable.

### Dependency Direction

| Package | Can depend on | Must not depend on |
|---|---|---|
| `src/<app>/main.py` | config and every adapter needed for composition | business rules or reusable domain behavior |
| `src/<app>/core/` | stdlib, domain values, core-owned ports | FastAPI, Pydantic DTOs, SQLAlchemy, HTTPX, telemetry SDKs |
| `src/<app>/api/http/` | core, config, transport DTOs, telemetry facade | SQLAlchemy sessions, migrations, direct outbound clients |
| `src/<app>/db/` | core ports and domain values, SQLAlchemy | FastAPI routers or HTTP DTOs |
| `src/<app>/clients/` | core ports, HTTPX or external SDKs | HTTP handlers or database implementation details |
| `src/<app>/workers/` | core use cases, broker adapters, telemetry facade | transport DTO reuse or unowned process lifecycle |
| `src/<app>/telemetry/` | logging, metrics, tracing, health integration | business decisions |

`config.py` or `config/` owns settings parsing. `api/http/` owns request/response mapping. `main.py` is the composition root and the only place allowed to know every concrete implementation.

### Enforce The Graph

Declare Import Linter contracts that forbid `core` importing `api`, `db`, `clients`, `config`, `telemetry`, or `workers`; forbid adapter-to-adapter shortcuts; and keep composition at the top. Run `uv run lint-imports` through `make imports`. Ruff `TID` rules govern relative and banned imports but do not replace architectural contracts.

Circular imports are design failures. Move the consumer-required abstraction inward, split a mixed-responsibility module, or inject a callback/port. Do not defer imports inside functions or create `utils.py` to conceal the cycle.

### Ports Are Consumer-Owned Protocols

Core defines narrow structural interfaces with `typing.Protocol`; adapters implement them without inheriting from framework-owned bases.

```python
class WidgetStore(Protocol):
    async def get(self, widget_id: WidgetId) -> Widget | None: ...
```

Define a port only for a real boundary or test seam. Keep it focused, accept protocols where substitution matters, and return concrete domain values. Protocols are the default because Python supports structural subtyping without forcing adapter inheritance; see the [typing specification](https://typing.python.org/en/latest/spec/protocol.html).

### Package Map

- `core/`: domain types, use cases, and consumer-owned ports.
- `api/http/`: FastAPI routers, Pydantic HTTP DTOs, auth adaptation, and problem responses.
- `db/`: engines, sessions, SQLAlchemy mappings, repositories, and Alembic integration.
- `clients/`: lifespan-shared external clients, timeouts, bounded retries, and response mapping.
- `config.py` or `config/`: settings models and startup validation.
- `telemetry/`: logging setup, OpenTelemetry, Prometheus, and health primitives.
- `workers/`: broker settlement, polling/scheduling loops, and drain behavior.

### Visibility And __init__.py

An underscore-prefixed module or name is internal to its owning package. This is a support-policy signal, not a security boundary. Keep `__init__.py` empty by default; re-export only a small, intentional public facade. Every re-export belongs in explicit `__all__`, remains typed and documented, and becomes a compatibility commitment.

Avoid star imports and implicit re-exports. Mypy strict mode's `no_implicit_reexport` behavior makes accidental facades visible.

### One Distribution By Default

A second distribution is justified only when it has independent consumers, versioning, ownership, release cadence, and compatibility testing. A desire to break an import cycle or share a few helpers is not justification. Record the split in an ADR and prove both built artifacts independently.

## Common Mistakes And Forbidden Patterns

- FastAPI, Pydantic DTO, SQLAlchemy model, HTTPX response, or telemetry SDK types in core.
- Routers querying the database or outbound client directly.
- Protocols created for every class before a boundary or substitute exists.
- Circular imports hidden with local imports, `TYPE_CHECKING` branches, or a junk-drawer module.
- Wildcard imports, broad re-export facades, or business behavior in `__init__.py`.
- `utils.py`, `common.py`, `helpers.py`, or `base.py` becoming the real architecture.
- A second distribution created only to simulate layering.

## Verification And Proof

```bash
uv run lint-imports
uv run ruff check .
uv run mypy .
make verify
```

Review each new import by asking who owns the behavior and whether the dependency points inward. For every Protocol, name its consumer and substitute. For every public re-export, name the supported caller. The import contracts, type checker, and tests must all agree.

Related: [shared constructs](shared-constructs.md), [typing discipline](typing-discipline.md), and [framework selection](../decisions/framework-selection.md).
