# Errors And Logging

Failure and structured-logging rules that preserve causes without leaking sensitive data.

## Default Approach

Model expected outcomes explicitly, throw typed errors for exceptional failures, and log once where action is possible.

### Error Taxonomy

Define a small repository-owned taxonomy such as validation, not-found, conflict, authorization, dependency, timeout, cancellation, and internal errors. Each type carries stable machine-readable identity, a safe message, and an optional `cause`.

Use discriminated result unions for routine business alternatives a caller must handle. Use exceptions for infrastructure and invariant failures. Never require callers to parse message text.

Catch values are `unknown`. Narrow with `instanceof Error` or a dedicated normalizer. Preserve `cause` when wrapping and keep the original stack available to internal telemetry.

Use a typed base with a stable code and native cause chaining:

```ts
export type ErrorCode =
  | "widget_not_found"
  | "widget_conflict"
  | "dependency_failed";

export class AppError extends Error {
  readonly code: ErrorCode;

  constructor(code: ErrorCode, message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "AppError";
    this.code = code;
  }
}

export class DependencyError extends AppError {
  constructor(message: string, cause: unknown) {
    super("dependency_failed", message, { cause });
    this.name = "DependencyError";
  }
}
```

Codes are closed, machine-readable vocabulary. Messages remain safe summaries and never carry secrets or raw dependency responses.

### Boundary Mapping

Transport adapters map internal failures into RFC 9457 problem details. Publish stable `type`, `title`, `status`, and safe extensions. Do not expose stack traces, SQL, dependency URLs containing credentials, schema internals, or raw rejected input.

Map cancellation separately from timeout. Map dependency failures conservatively; do not convert every exception to a client-visible 400 or every missing row to a server error.

Keep mapping exhaustive and transport-owned:

```ts
type Problem = Readonly<{
  type: string;
  title: string;
  status: number;
}>;

const problems = {
  widget_not_found: { type: "/problems/not-found", title: "Not found", status: 404 },
  widget_conflict: { type: "/problems/conflict", title: "Conflict", status: 409 },
  dependency_failed: { type: "/problems/unavailable", title: "Unavailable", status: 503 },
} as const satisfies Readonly<Record<ErrorCode, Problem>>;

export function toProblem(error: AppError): Problem {
  return problems[error.code];
}
```

Unknown errors map to a separate generic 500 at the final handler; they do not enter this expected-error table.

### Pino Logging

Use Pino JSON and child loggers with stable bindings: service, version, environment, request or trace context, component, and operation. Pass a narrow injected logger to reusable modules; core decisions do not depend on Pino.

Configure redaction at logger construction for authorization, cookies, tokens, passwords, connection strings, and known sensitive fields. Redaction is defense in depth; callers still must not log entire requests, config, rows, or payloads.

Construct redaction once and inject an operation-scoped child logger into the adapter:

```ts
import pino, { type Logger } from "pino";

export const rootLogger = pino({
  level: "info",
  redact: {
    paths: [
      "req.headers.authorization",
      "req.headers.cookie",
      "config.databaseUrl",
      "*.password",
      "*.token",
    ],
    censor: "[REDACTED]",
  },
});

export function widgetLogger(logger: Logger, requestId: string): Logger {
  return logger.child({ component: "widget", requestId });
}
```

Core modules accept no Pino type. An adapter may accept `Logger`; a reusable cross-framework module accepts a smaller repository-owned interface.

### Log Once And Act

Log at the boundary that retries, translates, drops, or terminates. Lower layers return or throw with context and `cause`. Do not log and rethrow at every layer.

Choose levels by operator action: debug for bounded diagnostics, info for lifecycle and material business milestones, warn for handled degradation, error for failed operations, fatal immediately before process termination.

Use stable event names and structured fields. Message templates are readable summaries, not the query interface. IDs may go in logs and traces; do not create high-cardinality metric attributes from them.

### Process Failures

Invalid startup configuration fails fast. `uncaughtException` and `unhandledRejection` are fatal: record a best-effort event, perform only bounded safe cleanup, and exit nonzero. Do not continue from an unknown process state.

## Common Mistakes And Forbidden Patterns

- Throwing strings, plain objects, or errors whose identity is only message text.
- Catching and returning a default value without an explicit fallback contract.
- Logging and rethrowing through multiple layers.
- Raw request bodies, headers, environment objects, database rows, or errors serialized wholesale.
- Dynamic log messages where a stable event and fields are required.
- Client-visible stack traces or internal dependency details.
- Continuing after an unhandled rejection or uncaught exception.

## Verification And Proof

- Mapping tests cover every public error type and safe unknown-error fallback.
- Cause chains survive wrapping and are visible only in internal telemetry.
- Tests capture logs and prove one event, correct level, stable fields, and redaction.
- Problem-details responses contain no stack, SQL, secret, or raw rejected payload.
- Cancellation and timeout produce distinct outcomes and telemetry.
- A controlled fatal-error smoke test exits nonzero after bounded cleanup.

Related: [async-and-cancellation.md](async-and-cancellation.md), [../services/http-services.md](../services/http-services.md), and [../operations/observability.md](../operations/observability.md).
