# Recipe: Add React Hook

Use this when reusable stateful behavior belongs in a hook.

## Files To Touch

- `src/hooks/use-<name>.ts` for genuinely shared behavior, or the owning `src/features/<feature>/` directory
- colocated `use-<name>.test.tsx`
- call sites and their behavior tests
- `src/lib/` adapter code when the hook calls an external boundary

## Steps

1. State the hook's single owned behavior; do not extract it merely to shorten JSX.
2. Keep derived values in render and event-driven work in event handlers.
3. Put only external synchronization in effects.
4. Include every reactive dependency; restructure instead of suppressing analysis.
5. Create and abort owned requests in cleanup, or pass TanStack Query's supplied signal.
6. Return the smallest stable readonly result and typed actions.
7. Avoid module state and mutable singleton caches.
8. Test initial render, rerender with changed inputs, action, error, and unmount.
9. Confirm Strict Mode setup/cleanup repetition cannot duplicate an irreversible effect.

```bash
npm test -- --runInBand src/hooks/use-<name>.test.tsx
npm run lint
npm run typecheck
```

## Invariants To Preserve

- Hooks are called only at component or hook top level.
- Server state remains owned by TanStack Query.
- Every subscription, timer, observer, and request has cleanup.
- No `isMounted` flag substitutes for cancellation.
- Promise outcomes are awaited, returned, or owned and observed.
- Hook results do not leak mutable internal state.

## Proof

- Tests prove rerender behavior and dependency changes.
- Unmount tests prove abort and cleanup exactly once per setup.
- Error and cancellation outcomes remain distinguishable.
- Hooks lint rules pass without suppression.
- `npm run verify` is green.
