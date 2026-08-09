# Recipe: Add HTTP Middleware

Use this for a Fastify hook, plugin, decorator, or cross-cutting request policy.

## Files To Touch

- `src/api/plugins/<name>.ts`
- `src/api/plugins/<name>.test.ts`
- the registration module under `src/api/`
- `src/core/` only when a narrow policy port is required
- configuration, telemetry, and runbook surfaces affected by the policy

## Steps

1. Name the lifecycle phase and scope the policy needs: application, plugin subtree, or route.
2. Implement it as a Fastify plugin so encapsulation and registration order are explicit.
3. Parse any header, token, or plugin option before attaching a typed value.
4. Use a unique owned decorator name and add one augmentation file if typing requires it.
5. Prove runtime decoration occurs before any consumer reads the property.
6. Keep authentication, authorization, correlation, and logging responsibilities distinct.
7. Observe async failure; never pass an unowned rejected promise through a void callback.
8. Redact sensitive values at logger construction and emit failures once.
9. Register the plugin at the narrowest scope and document ordering constraints.

```bash
npm test -- --runInBand src/api/plugins/<name>.test.ts
npm run lint
npm run verify
```

## Invariants To Preserve

- Core code imports no Fastify request, reply, plugin, or decorator type.
- Hooks do bounded work and propagate cancellation.
- Authentication never substitutes for resource authorization.
- Request identifiers are validated for length and format or replaced.
- Middleware cannot silently convert an internal failure into success.
- Plugin encapsulation is intentional and covered by a negative scope test.

## Proof

- `app.inject()` tests cover included, excluded, failure, and registration-order cases.
- Negative tests prove missing or invalid credentials and forbidden access remain distinct.
- A lint/type fixture proves augmentation and runtime registration agree.
- Captured logs prove one safe event with expected bindings and redaction.
- `npm run verify` is green.
