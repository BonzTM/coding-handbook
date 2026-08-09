# HTTP Services

Fastify service rules for validated, bounded, observable, and evolvable HTTP boundaries.

## Default Approach

Use Fastify v5 with its Zod type provider, thin route adapters, and explicit plugin scope.

### Application Shape

Build the Fastify application in a function that receives config, logger, core ports, and lifecycle dependencies. Register plugins in deliberate scope and order. `src/index.ts` starts listening and owns signals; application construction remains injectable for tests.

Routes parse transport values, authorize, map into domain inputs, call one use case, and map the result. Business decisions do not live in handlers, hooks, or decorators.

Use Express 5 only for inherited middleware or platform compatibility and record the constraint. Do not run both frameworks in one new service.

A route plugin keeps the provider and schemas in its Fastify scope:

```ts
import Fastify, { type FastifyPluginAsync } from "fastify";
import type { ZodTypeProvider } from "fastify-type-provider-zod";
import { serializerCompiler, validatorCompiler } from "fastify-type-provider-zod";
import { z } from "zod";

export function buildHttpApp() {
  const app = Fastify();
  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);
  return app;
}

const createBody = z.strictObject({ name: z.string().min(1).max(100) });
const widgetResponse = z.strictObject({ id: z.uuid(), name: z.string() });

type WidgetService = Readonly<{
  create(input: { readonly name: string }, signal: AbortSignal): Promise<z.infer<typeof widgetResponse>>;
}>;

export const widgetRoutes: FastifyPluginAsync<{ widgets: WidgetService }> = async (
  app,
  options,
) => {
  const typed = app.withTypeProvider<ZodTypeProvider>();
  typed.post("/widgets", {
    schema: { body: createBody, response: { 201: widgetResponse } },
  }, async (request, reply) => {
    const controller = new AbortController();
    const abort = (): void => controller.abort(new Error("client disconnected"));
    request.raw.once("aborted", abort);
    try {
      const widget = await options.widgets.create({ name: request.body.name }, controller.signal);
      return reply.code(201).send(widget);
    } finally {
      request.raw.off("aborted", abort);
    }
  });
};
```

Application construction installs both compilers before registering routes.

### Schemas And Limits

Declare Zod schemas for params, query, headers, body, and response. Type inference follows runtime schemas; it never substitutes for parsing. Set explicit body, header, query, upload, and response-size limits before expensive work.

Reject unsupported content types and malformed encodings. Normalize only what the contract allows. Never pass a Fastify request, raw payload, or asserted DTO into core logic.

### Errors And Problem Details

Map typed failures to RFC 9457 problem details with stable type URIs, status, title, and safe extensions. Validation errors are client-safe summaries; internal stacks, SQL, dependency details, secrets, and rejected payloads remain private.

Install one final unknown-error handler. Log once with request and trace context, then return a generic internal problem. Cancellation, timeout, dependency failure, conflict, and not-found remain distinct.

```ts
app.setErrorHandler((error, request, reply) => {
  const problem = error instanceof AppError
    ? toProblem(error)
    : { type: "/problems/internal", title: "Internal error", status: 500 };
  if (!(error instanceof AppError)) request.log.error({ err: error }, "request failed");
  void reply.type("application/problem+json").code(problem.status).send({
    ...problem,
    instance: request.id,
  });
});
```

### Authentication And Authorization

Authenticate at the earliest owned hook, but authorize the concrete action and resource before core mutation. Default deny. Pass a typed principal inward, not raw token claims.

CSRF protection is mandatory for cookie-authenticated state changes. CORS is an allowlist policy, not authentication. Security headers and proxy trust are configured for the actual deployment topology.

### Idempotent Writes

State-changing endpoints that clients or infrastructure may retry define an idempotency contract.

- Require a validated key, scoped to principal, operation, and target.
- Store a request fingerprint and durable outcome atomically with the effect where feasible.
- Replay the same status, headers, and safe response for a matching duplicate.
- Reject key reuse with a different request fingerprint.
- Define retention, expiry, in-progress behavior, concurrency control, and maximum key length.
- Do not cache authorization decisions beyond the request; re-establish the principal before replay policy allows a response.

Exactly-once delivery is not claimed. Tests prove duplicate concurrent requests create one durable effect and that partial failure is recoverable.

### Outbound And Async Work

Pass the request `AbortSignal` through owned work. Bound dependency calls by a budget shorter than the server deadline. Do not detach work from the response unless a durable queue owns it.

Streams honor backpressure and abort. Uploads are size-limited, content-validated, and never trusted by file extension alone.

### Telemetry And Health

Use Pino child bindings and OpenTelemetry request spans. Record route templates, status class, and bounded error identity; never use raw paths, user IDs, or request IDs as metric attributes.

Expose `/livez` for process liveness and `/readyz` for ability to accept work. Readiness checks owned dependencies with a short bound and does not turn every optional downstream outage into process death.

### Tests

Use `app.inject()` for route behavior. Test exact schema rejection, authentication, authorization, problem mapping, idempotency, limits, content type, cancellation, and safe logging. Add one start/shutdown smoke test for real listener lifecycle.

```ts
it("creates a validated widget", async () => {
  const app = buildApp({ widgets: fakeWidgets });
  try {
    const response = await app.inject({
      method: "POST",
      url: "/widgets",
      payload: { name: "Meter" },
    });
    const body: unknown = response.json();
    expect(response.statusCode).toBe(201);
    expect(body).toEqual({ id: expect.any(String), name: "Meter" });
  } finally {
    await app.close();
  }
});
```

## Common Mistakes And Forbidden Patterns

- Domain behavior or SQL in handlers and hooks.
- Type-provider inference treated as runtime validation without schema execution.
- Framework request objects passed into core.
- Unbounded bodies, streams, pagination, or fan-out.
- CORS treated as authorization or trusted proxy settings copied blindly.
- A retryable write without an idempotency contract.
- Fire-and-forget work started after the response.
- Raw URLs or identities used as metric dimensions.

## Verification And Proof

- `app.inject()` tests cover success plus malformed, unauthenticated, forbidden, conflict, timeout, and unknown-error paths.
- Schemas constrain every request and documented response location.
- Concurrent duplicate idempotency tests yield one durable effect and stable replay.
- Body, upload, response, pagination, and deadline bounds are exercised.
- Logs and problem details contain no secret or rejected sensitive payload.
- `/livez`, `/readyz`, startup, and bounded shutdown pass smoke tests.

Related: [../foundations/serialization.md](../foundations/serialization.md), [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md), and [../operations/security.md](../operations/security.md). Provider anchor: [fastify-type-provider-zod](https://github.com/turkerdev/fastify-type-provider-zod#how-to-use).
