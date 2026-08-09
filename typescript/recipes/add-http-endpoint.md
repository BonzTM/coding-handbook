# Recipe: Add HTTP Endpoint

Use this when one Fastify route is added or changed.

## Files To Touch

- `src/api/<feature>-route.ts` and its colocated test
- `src/api/schemas/<feature>.ts` for request and response Zod schemas
- `src/core/<feature>.ts` and core tests
- `src/index.ts` or the owned API registration module
- contract artifacts and telemetry when observable behavior changes

## Steps

1. Define strict Zod schemas for params, query, headers, body, success, and problem responses.
2. Define transport DTOs from schema output; map them explicitly to core values.
3. Add or update the focused core use case. Keep Fastify and database types out of `src/core/`.
4. Implement the route handler as parse, authorize, call, map, and reply.
5. Pass `request.signal` or the repository-owned request cancellation signal through every I/O call.
6. Map typed failures to stable RFC 9457 problem details; keep unknown failures internal.
7. Register the route in the correct Fastify plugin scope and confirm hook ordering.
8. Use route templates, status classes, and bounded result values for telemetry.
9. Update OpenAPI or other controlling contract artifacts and compatibility fixtures.

Copy-pasteable proof commands:

```bash
npm test -- --runInBand src/api/<feature>-route.test.ts
npm run typecheck
npm run verify
```

## Invariants To Preserve

- `api -> core <- db`; handlers contain no SQL or business decisions.
- External input is parsed from `unknown` before core use.
- Authentication and resource authorization run server-side.
- Request bodies, pagination, fan-out, time, and response size are bounded.
- Logs contain no raw credentials, bodies, or rejected sensitive values.
- Every documented response status and content type has a schema and test.

## Proof

- `app.inject()` tests cover success, malformed input, unauthenticated, forbidden, and mapped failure.
- A core test proves the business decision without Fastify or PostgreSQL.
- Contract fixtures and generated artifacts have reviewed diffs.
- A composition smoke test proves registration and production error mapping.
- `npm run verify` is green.
