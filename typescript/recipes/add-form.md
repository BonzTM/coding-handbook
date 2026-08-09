# Recipe: Add Form

Use this when a React feature accepts and submits user input.

Reference implementation: [examplefrontend](../reference/examplefrontend/).

## Files To Touch

- the owning form component under `src/features/<feature>/`
- a colocated Zod form schema and mapper
- a TanStack Query mutation hook
- component tests and MSW handlers
- backend contract or idempotency surfaces when submission changes them

## Steps

1. Define the field, omission, null, normalization, and maximum-size contract in Zod.
2. Use a native `form`, labels, controls, buttons, and browser semantics first.
3. Parse at submit; map parsed form values to the API request explicitly.
4. Associate each error with its control and provide a useful error summary.
5. Move focus to the summary or first invalid control after rejection.
6. Preserve input after client or server failure.
7. Use a TanStack Query mutation; invalidate only affected query keys.
8. Disable duplicate clicks for UX and use backend idempotency for correctness.
9. Add MSW tests for accepted, rejected, malformed, aborted, and network outcomes.

```bash
npm test -- --runInBand src/features/<feature>/<name>-form.test.tsx
npm run lint
npm run verify
```

## Invariants To Preserve

- Client validation does not replace server validation or authorization.
- Fields have labels, errors are programmatically associated, and keyboard submit works.
- Raw server internals and rejected sensitive values are never rendered or logged.
- Pending state retains context and prevents accidental duplicate interaction.
- Mutation retries are disabled unless the write is demonstrably idempotent.
- User input is not persisted in browser storage without data review.

## Proof

- RTL tests cover keyboard entry, validation, focus, submit, failure, and resubmit.
- MSW fails any unhandled request and validates method, path, headers, and body.
- A duplicate-click test proves one client request; backend proof covers one effect.
- Accessibility lint and user-behavior review pass.
- `npm run verify` is green.
