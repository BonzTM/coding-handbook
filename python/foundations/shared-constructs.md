# Shared Constructs

Reusable in-repo building blocks that remove repeated wiring without creating a junk drawer.

## Default Approach

Each shared construct has one owner and a narrow contract.

| Construct | Owns | Do not turn it into |
|---|---|---|
| `telemetry/` | logging `dictConfig`, OpenTelemetry/Prometheus setup, health state, exporter shutdown | vendor wrappers around every call or business metrics policy |
| `config.py` / `config/` | Pydantic settings, defaults, source precedence, startup validation | ambient service locator or scattered env reads |
| `clients/` factory | one lifespan-owned `httpx.AsyncClient`, mandatory `httpx.Timeout`, pool limits, common transport wiring | one universal client hiding domain response/retry semantics |
| `tests/testutil/` | focused fakes, fake clock, builders, boundary harnesses | production behavior or an assertion DSL that hides failures |
| `create_app()` | FastAPI construction, router registration, lifespan composition | process execution or business rules |

### Constructor And Factory Pattern

Pass dependencies explicitly. Long construction calls are acceptable when they reveal the real graph. Group values only when they have one lifecycle or form a cohesive configuration object. The app factory receives or constructs a typed settings object and returns an unstarted app; its lifespan acquires and releases shared async resources.

### HTTP Client Factory

The factory creates one `httpx.AsyncClient` per application lifespan with an explicit `httpx.Timeout` covering connect/read/write/pool behavior and bounded pool limits. Domain-specific clients receive it and own URLs, response parsing, retry eligibility, and core mapping. Tests substitute transports or stub servers without global monkeypatching.

### Naming And Growth

Name constructs for ownership (`telemetry`, `clients`, `testutil`), never generic reuse. A second caller alone does not justify abstraction; the extracted behavior must have one coherent contract. Split a module when it accumulates unrelated owners.

## Common Mistakes And Forbidden Patterns

- `utils.py`, `common.py`, `helpers.py`, or `base.py` accumulating unrelated behavior.
- A global client, settings singleton, mutable registry, or import-time resource creation.
- Telemetry SDKs or config lookups hidden in core.
- Per-request HTTP clients or clients without explicit timeout policy.
- Test helpers that contain production decisions or make assertions opaque.

## Verification And Proof

```bash
uv run lint-imports
make verify
```

A contributor can read composition and name every resource owner. Each construct states what it owns, has focused tests, and can be removed without finding unrelated dependents. No junk-drawer filename or hidden global dependency remains.

Related: [package design](package-design.md), [project setup](project-setup.md), and [observability](../operations/observability.md).
