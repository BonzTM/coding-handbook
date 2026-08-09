# Recipe: Add React Component

Use this when adding one reusable React 19 component.

Reference implementation: [examplefrontend](../reference/examplefrontend/).

## Files To Touch

- `src/components/<name>/<name>.tsx` for shared UI, or `src/features/<feature>/` for feature-owned UI
- colocated `<name>.test.tsx`
- owned styles or tokens
- the consuming route or feature module
- accessibility documentation when behavior is non-obvious

## Steps

1. Keep the component feature-owned until two stable consumers justify promotion.
2. Define readonly props in domain-facing terms and callbacks as user intent.
3. Use native semantic HTML before ARIA or custom interaction primitives.
4. Render loading, empty, error, disabled, and success states that the contract needs.
5. Keep server state in TanStack Query and complex decisions outside JSX.
6. Give controls accessible names, visible focus, and keyboard behavior.
7. Observe async callbacks through an owned handler; do not float promises.
8. Add React Testing Library tests using roles and accessible names.
9. Verify the component under React Strict Mode behavior where effects exist.

```bash
npm test -- --runInBand src/components/<name>/<name>.test.tsx
npm run lint
npm run build
```

## Invariants To Preserve

- Components are function components unless a render error boundary requires a class.
- Effects synchronize external systems and always clean up.
- No unsafe HTML, secret-bearing browser config, or authorization-only UI guard.
- Test IDs do not replace accessible queries.
- Props do not expose transport, query-cache, or mutable implementation details.
- Accessibility lint findings are merge blockers.

## Proof

- RTL tests cover pointer and keyboard use through accessible queries.
- Tests cover each user-visible state and error recovery action.
- Hook/effect behavior is tested through rerender and unmount when applicable.
- `eslint-plugin-react-hooks` and `eslint-plugin-jsx-a11y` report zero warnings.
- `npm run verify` is green.
