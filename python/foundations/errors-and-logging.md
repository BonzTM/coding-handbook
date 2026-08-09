# Errors and Logging

Exception semantics and structured logging rules that keep failures actionable instead of noisy.

## Default Approach

### Exception Semantics

- Define a small module-level hierarchy rooted in one domain base exception; add a subtype only when a boundary branches on that condition.
- Catch only where code can map, retry within a bound, compensate, or add subsystem context.
- Wrap at subsystem boundaries with `raise DomainFailure(...) from err`; use `raise` to preserve the active traceback when rethrowing unchanged.
- Catch the narrowest expected types. `except Exception` belongs only in the HTTP exception boundary and worker supervisor; bare `except:` is forbidden.
- Expected business outcomes handled by every caller use explicit return types, not exception-driven control flow.
- `asyncio.CancelledError` is control for cooperative shutdown and must propagate after cleanup; see [concurrency](concurrency-and-asyncio.md).

When `TaskGroup` siblings fail, Python raises an `ExceptionGroup`. Use `except*` only where the owner can handle a specific subgroup; leave unhandled failures grouped and visible. The [Python 3.11 exception-group documentation](https://docs.python.org/3.11/library/exceptions.html#exception-groups) defines the split behavior.

### Wire Error Contract

Exactly one HTTP boundary maps domain exceptions to RFC 9457-style `application/problem+json`. Known conditions receive stable type/status/title fields and safe extensions; unknown failures return an opaque 500 with a request identifier. Never expose exception text or tracebacks. The complete transport contract lives in [HTTP services](../services/http-services.md).

### Structured Logging

Configure stdlib `logging` once with `logging.config.dictConfig` in the composition root. Services emit JSON; local development may select readable output. Modules use `logger = logging.getLogger(__name__)`. Libraries configure no application handlers and may add only `logging.NullHandler`, consistent with the [logging library guidance](https://docs.python.org/3/howto/logging.html#configuring-logging-for-a-library).

Log once at the boundary that can act. Lower layers raise or return structured failures. Use `logger.exception(...)` only inside an active exception handler when the traceback is actionable; use `logger.error(...)` for a failure without an active traceback. Stable fields include service, operation, request/trace identifier, and safe domain identifiers.

## Common Mistakes And Forbidden Patterns

- Bare `except:`, broad catches in ordinary functions, empty handlers, or swallowed cancellation.
- `raise replacement` without `from err`, or parsing exception strings to branch.
- Logging the same exception at every layer.
- Secrets, credentials, auth headers, PII, raw payloads, or stack traces in wire responses.
- `basicConfig` or handlers in a library; logging configured during import.
- `print()` in services; Ruff `T20` makes it a gate failure.
- `logger.exception` outside an exception handler or interpolation that discards structured fields.

## Verification And Proof

```bash
uv run ruff check .
uv run pytest -k "error or logging or problem"
make verify
```

Tests assert exception types and structured properties, explicit chaining, subgroup handling where used, exact problem content type and safe shape, one error log per failure, and redaction. Exercise one successful and failing request plus one cancelled worker path end to end.

Related: [serialization](serialization.md), [observability](../operations/observability.md), and [data handling](../operations/data-handling.md).
