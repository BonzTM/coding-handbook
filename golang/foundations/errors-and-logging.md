# Errors and Logging

Error semantics and structured logging rules that keep failures actionable instead of noisy.

## Default Approach

### Error Semantics

- Wrap with `fmt.Errorf("...: %w", err)` when callers should still be able to inspect the underlying failure.
- Use `errors.Is` for sentinel matching and `errors.As` for typed errors.
- Preserve opaque boundaries where implementation details should stay private.
- Add context at package or subsystem boundaries, not mechanically at every stack frame.

### Error Categories

| Kind | Use for | Example handling |
|---|---|---|
| opaque internal error | callers only need success or failure | log and map to generic transport error |
| sentinel error | caller branches on a stable condition | `errors.Is(err, ErrNotFound)` |
| typed error | caller needs structured data | validation or policy errors with fields |
| transport mapping error | HTTP/gRPC status mapping | convert from domain error to `404`, `409`, `codes.InvalidArgument`, etc. |

### Structured Logging

- Use `log/slog` as the default logger.
- Prefer JSON logs in production and text logs locally.
- Inject a logger into services and adapters; do not hide it behind global state in reusable packages.
- Standard fields should usually include `service`, `operation`, and correlation data such as `request_id` or `trace_id` when available.

## Log Placement Rules

- Handlers and RPC methods log request lifecycle, status, and latency.
- Background workers log lifecycle and actionable failures.
- Repositories and external clients log retries, dependency failures, and unusual latency when that information helps operators.
- Libraries should not spam logs; they return errors and let the application decide.

## Common Mistakes And Forbidden Patterns

- Comparing error strings.
- Logging the same error at every layer.
- Logging secrets, tokens, full auth headers, or raw user payloads by default.
- Replacing structured attributes with preformatted strings everywhere.
- Using `panic` for expected runtime failures.

## Verification And Proof

- Unit or integration tests should prove that important errors remain matchable with `errors.Is` or `errors.As`.
- Review one successful and one failing request path to ensure logs contain stable fields and no sensitive data.
- Search the repo for `fmt.Print`, `log.Print`, and ad hoc logging paths before calling the logging shape consistent.
