# Recipe: Add Config Key

Use this when a service or frontend adds one configuration value.

## Files To Touch

- `src/config/schema.ts` and config tests
- `src/index.ts` composition or the adapter receiving a narrow slice
- `.env.example`
- deployment values and secret references
- project README, changelog, and runbook when operator-visible

## Steps

1. Name the key, owner, source, safe default, sensitivity, unit, and bounds.
2. Add it to the Zod startup schema; parse strings explicitly.
3. Select the owned environment name before `z.strictObject` parsing.
4. Expose the result as readonly and inject only the slice a consumer needs.
5. Add a safe placeholder and description to `.env.example`.
6. Update every deployment environment and secret grant before rollout.
7. Keep server secrets out of `VITE_*`; frontend values are public.
8. Add valid, absent, empty, malformed, defaulted, and boundary tests.
9. Document rotation, restart, or dynamic-refresh behavior where applicable.

```bash
npm test -- --runInBand src/config
npm run typecheck
npm run verify
```

## Invariants To Preserve

- Only `src/config/` reads `process.env`.
- Parsing completes before listeners, workers, pools, or clients start.
- Errors name invalid keys but never reveal secret values.
- Boolean and numeric strings are parsed, not coerced by truthiness.
- Defaults are safe and deterministic.
- `.env.example`, deployment, schema, and runbook names agree.

## Proof

- Config tests cover every accepted and rejected representation.
- Startup smoke tests fail before accepting work when the key is invalid.
- Search proves no new ambient environment read.
- Artifact and log inspection finds no secret or unintended frontend value.
- `npm run verify` is green.
