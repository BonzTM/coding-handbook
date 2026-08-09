# Frontend Data And State

Ownership rules for TanStack Query server state and minimal React client state.

## Default Approach

TanStack Query owns remote server state; React owns local presentation state; the URL owns shareable navigation state.

### State Classification

Before adding state, classify it:

- server state: fetched, cached, invalidated, and reconciled by TanStack Query;
- URL state: route, filter, sort, pagination, and shareable selections;
- local state: transient input, disclosure, selection, and UI coordination;
- derived state: computed during render, not stored;
- cross-feature client state: narrow context only after ownership is proven.

Do not copy query data into local state merely to read or filter it. A dedicated state machine or Redux requires complex client-only transitions and an ADR; neither is the server-state default.

### API Client Boundary

One typed client owns URL resolution, credentials, headers, timeout, abort, response-size policy, problem parsing, and Zod validation. Query functions receive and pass TanStack Query's `signal`.

The client returns domain-facing DTOs or typed problems, never unchecked generic JSON. Authentication refresh is centralized, bounded, and unable to create retry loops.

```ts
import { z } from "zod";

export async function getJson<S extends z.ZodType>(
  url: URL,
  schema: S,
  signal: AbortSignal,
): Promise<z.output<S>> {
  const response = await fetch(url, {
    headers: { accept: "application/json" },
    signal,
  });
  if (!response.ok) throw await parseProblem(response);
  const body: unknown = await response.json();
  return schema.parse(body);
}
```

The production wrapper also composes a timeout, checks content type and response-size bounds, and maps abort separately from an HTTP problem.

### Query Keys

Define query-key factories by resource and normalized parameters. Keys include every input that changes the result and exclude secrets, access tokens, raw objects, and unstable references.

```ts
export const widgetKeys = {
  all: ["widgets"] as const,
  detail: (id: WidgetId) => ["widgets", "detail", id] as const,
};
```

Choose `staleTime`, garbage collection, retry, refetch, and network behavior from product semantics. Defaults are explicit at composition; a retry never applies blindly to authorization, validation, or non-idempotent work.

### Mutations

Mutation functions validate inputs and surface typed problems. On success, update or invalidate only affected queries. Prefer response-driven cache updates when the server returns the canonical resource.

Optimistic updates require a reversible snapshot, conflict policy, failure rollback, and tests. Do not optimistically represent an effect the server may reject for authorization or business rules without a deliberate UX contract.

Prevent duplicate clicks for usability; server idempotency protects correctness. Mutation cancellation and navigation behavior are explicit.

```tsx
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export function useWidget(id: WidgetId) {
  return useQuery({
    queryKey: widgetKeys.detail(id),
    queryFn: ({ signal }) => getJson(widgetUrl(id), widgetSchema, signal),
    staleTime: 30_000,
  });
}

export function useCreateWidget() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateWidget) => createWidget(input),
    onSuccess: async (widget) => {
      queryClient.setQueryData(widgetKeys.detail(widget.id), widget);
      await queryClient.invalidateQueries({ queryKey: widgetKeys.all });
    },
  });
}
```

`createWidget` owns request validation, idempotency, timeout, and response parsing; the hook owns cache effects.

### Loading And Errors

Differentiate initial loading, background refresh, empty success, stale data with warning, recoverable error, and terminal forbidden/not-found states. Do not replace usable stale content with a full-page spinner during background fetch.

Error UI gives a safe message and an actionable retry when appropriate. Log or trace once at the boundary that can act; do not expose raw server internals.

### Persistence And Hydration

Do not persist query caches by default. Persisted browser state requires data classification, versioning, expiry, logout clearing, privacy review, and safe migration. Never place tokens or sensitive records in localStorage.

When hydration is used, prove server/client schema, query keys, and timestamps agree. Treat dehydrated data as untrusted when it crosses an HTML or storage boundary.

## Common Mistakes And Forbidden Patterns

- Server data copied into `useState` or broad context.
- Query keys missing a filter, tenant, locale, or authorization-sensitive scope.
- Infinite retries, auth refresh loops, or retries for client errors.
- Broad cache invalidation after every mutation.
- Optimistic changes without rollback and conflict tests.
- Sensitive data persisted in localStorage or an unversioned cache.
- Loading, empty, and error states collapsed into one ambiguous branch.

## Verification And Proof

- Tests cover loading, empty, success, stale refresh, malformed response, error, retry, and cancellation.
- Query-key tests prove distinct inputs cannot collide or leak between tenants/users.
- Mutation tests cover success, validation failure, conflict, retry policy, invalidation, and optimistic rollback.
- MSW fails unhandled requests and validates request contracts.
- Logout and principal changes clear or segregate sensitive cached data.
- No query function discards the supplied `AbortSignal`.

Related: [react-applications.md](react-applications.md), [../foundations/serialization.md](../foundations/serialization.md), and [../operations/data-handling.md](../operations/data-handling.md).
