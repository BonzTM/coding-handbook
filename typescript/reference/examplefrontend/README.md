# examplefrontend

A complete React 19 and Vite widget administration application that proves the TypeScript handbook's frontend composition, data, accessibility, testing, audit, and build rules together.

## What It Is

- React 19 function components with React Router route composition, a lazy detail route, Suspense status UI, router errors, and a final class render-error boundary.
- TanStack Query-owned server state with a resource query-key factory, cursor pagination, bounded retry defaults, precise invalidation, response-driven detail caching, and reversible optimistic create/delete updates.
- A single typed fetch boundary with URL ownership, JSON headers, a ten-second timeout, caller cancellation, one-megabyte response bounds, safe HTTP error mapping, and Zod-parsed responses.
- Wire schemas matching [exampleservice](../exampleservice/)'s widget list, create, get, and delete contract without importing backend implementation types.
- A native create form with Zod validation, associated accessible errors, invalid-field focus, retained input, idempotency keys, and duplicate-submit prevention.
- Jest 30's Babel transform with jsdom, React Testing Library, user-event, and MSW. Tests use accessible roles and names and reject every unhandled request.

## Requirements

- Node.js 24.18.0 (pinned in `.nvmrc`)
- npm and the committed `package-lock.json`
- The sibling [exampleservice](../exampleservice/) on port 3000 for live API development, or the opt-in MSW browser worker

## Setup And Run

```bash
npm ci
cp .env.example .env
npm run dev
```

Vite proxies `/widgets` to `http://localhost:3000` by default. To work entirely offline, set `VITE_ENABLE_MSW=true`; the browser then starts the committed `public/mockServiceWorker.js` and uses the same contract-validating handlers as the Jest suite. Recreate that generated worker after an MSW upgrade with `npx msw init public --save`.

`VITE_API_BASE_URL` may point at another public HTTP(S) API origin. Like every Vite-exposed value, it is public configuration and must never contain a secret.

## Package Map

| Path                               | Responsibility                                                                    |
| ---------------------------------- | --------------------------------------------------------------------------------- |
| `src/app/`                         | router, providers, QueryClient defaults, Suspense, and error-boundary composition |
| `src/routes/`                      | list and lazy detail navigation boundaries, route params, and page titles         |
| `src/features/widgets/api/`        | exampleservice-compatible Zod wire schemas and widget API slice                   |
| `src/features/widgets/hooks/`      | query keys, queries, optimistic mutations, rollback, and invalidation             |
| `src/features/widgets/components/` | accessible create form and cursor-paginated widget list                           |
| `src/components/`                  | shared presentation-only status UI                                                |
| `src/lib/api/`                     | bounded fetch, abort propagation, response parsing, and typed errors              |
| `src/mocks/`                       | shared browser/test MSW handlers                                                  |
| `src/test/`                        | jsdom polyfills, MSW lifecycle, and application render composition                |

## Verification

The canonical offline gate runs formatting, ESLint with React Hooks and jsx-a11y at zero warnings, strict type checking, Jest with at most two workers, the high-severity npm audit policy, and the Vite production build:

```bash
npm run verify
# equivalent shim
make verify
```

Tests cover list loading/error/success and cursor pagination, accessible client validation, optimistic create plus invalidation, delete, navigation and cold deep links, malformed route input, Zod response rejection, caller abort, and HTTP problem mapping.

## Release And Recovery

`npm run build` writes fingerprinted static assets to `dist/`; this exemplar therefore has no Dockerfile. Publish that immutable directory through the approved static host or CDN. Serve HTML with revalidation/no-cache, fingerprinted assets with long immutable caching, and an SPA deep-link fallback that excludes `/widgets` API paths and missing assets. Set CSP and other security headers at the delivery layer, keep source maps private unless a review approves publication, and roll back by restoring the prior artifact digest.

## Related Exemplars And Handbook Docs

- [exampleservice](../exampleservice/) implements the HTTP and PostgreSQL widget contract consumed here.
- [exampleworker](../exampleworker/) projects widget-domain events with bounded at-least-once processing.
- [React applications](../../services/react-applications.md) governs components, routing, forms, accessibility, errors, and assets.
- [Frontend data and state](../../services/frontend-data-and-state.md) governs QueryClient defaults, keys, cancellation, optimistic updates, and invalidation.
- [Testing](../../quality/testing.md), [linting](../../quality/linting.md), and [deployment](../../operations/deployment.md) define the verification and static-delivery proof.
