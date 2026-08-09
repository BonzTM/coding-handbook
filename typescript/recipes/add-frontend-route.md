# Recipe: Add Frontend Route

Use this when adding a React Router page boundary.

## Files To Touch

- `src/routes/<name>-route.tsx`
- router composition under `src/app/`
- owning feature under `src/features/<feature>/`
- navigation components and route tests
- API query definitions under the feature or `src/lib/`

## Steps

1. Define the path, params, search contract, title, authorization UX, and not-found behavior.
2. Parse URL values before feature use and keep shareable state in the URL.
3. Lazy-load at the route boundary when it creates a meaningful bundle split.
4. Supply loading and recoverable error boundaries with accessible status.
5. Use TanStack Query for server state and pass its cancellation signal.
6. Treat client auth guards as UX only; the server still authorizes resources.
7. Restore page title and a useful focus target after navigation.
8. Add navigation, direct deep-link, back/forward, forbidden, and not-found tests.
9. Verify deployment fallback does not rewrite API or asset failures to HTML success.

```bash
npm test -- --runInBand src/routes/<name>-route.test.tsx
npm run build
npm run verify
```

## Invariants To Preserve

- Navigation side effects do not run during render.
- Route params and search values are untrusted until parsed.
- Loading does not discard usable stale data without reason.
- Error UI exposes no stack or raw server response.
- Deep links work from a cold browser load.
- Route modules do not become business-logic containers.

## Proof

- Memory-router tests prove navigation and URL-state behavior.
- A deployed or preview-server smoke test proves a cold deep link loads.
- Tests cover loading, success, forbidden, not-found, and boundary recovery.
- Production bundle inspection confirms the intended lazy chunk.
- `npm run verify` is green.
