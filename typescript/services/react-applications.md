# React Applications

React 19 and Vite rules for accessible, testable, and bounded browser applications.

## Default Approach

Use React 19 function components, Vite, React Router, semantic HTML, and feature-owned UI modules.

### Application Shape

Composition creates the router, QueryClient, error boundaries, telemetry, and typed API client. Organize code by feature; keep routes and components thin over feature behavior and boundary adapters.

Prefer server rendering or another framework only when product requirements justify an ADR. Do not add a second router, build system, component runtime, or state store beside the defaults.

### Components And Hooks

Components render state and translate user intent into typed actions. Keep data parsing, network policy, and complex decisions outside JSX. Extract a hook when it owns reusable stateful behavior, not merely to shorten a component.

Obey Rules of Hooks and exhaustive dependency analysis. Effects synchronize with external systems; they are not a default place for derived state, event handling, or server-state fetching.

Every effect cleans up subscriptions, timers, observers, and requests. Use an `AbortController` or the signal supplied by the owning library. Strict Mode development behavior must not duplicate irreversible effects.

A component accepts readonly domain-facing props and exposes user intent through a typed callback:

```tsx
type WidgetCardProps = Readonly<{
  name: string;
  status: "active" | "disabled";
  onSelect: () => void;
}>;

export function WidgetCard({ name, status, onSelect }: WidgetCardProps) {
  return (
    <article>
      <h2>{name}</h2>
      <p>Status: {status}</p>
      <button type="button" onClick={onSelect}>Select widget</button>
    </article>
  );
}
```

Effects synchronize one external subscription and return its cleanup:

```tsx
import { useEffect, useState } from "react";

export function useOnlineStatus(): boolean {
  const [online, setOnline] = useState(() => navigator.onLine);
  useEffect(() => {
    const markOnline = (): void => setOnline(true);
    const markOffline = (): void => setOnline(false);
    window.addEventListener("online", markOnline);
    window.addEventListener("offline", markOffline);
    return () => {
      window.removeEventListener("online", markOnline);
      window.removeEventListener("offline", markOffline);
    };
  }, []);
  return online;
}
```

### Routing

Use React Router with route-level loading/error boundaries and lazy imports at meaningful page boundaries. Deep links must work without prior client navigation. Authentication guards do not replace server authorization.

Define not-found and forbidden behavior explicitly. Preserve accessible focus and page-title updates after navigation. Avoid navigation side effects during render.

### Forms

Use native form semantics first. Parse user input with Zod at the form boundary, show errors beside associated controls, move focus to a useful summary or first invalid control, and preserve user input after rejection.

Disable duplicate submission only as UX; backend idempotency owns correctness. Show pending state without removing context, and surface safe server problem details.

### Accessibility

Use semantic elements, labels, accessible names, keyboard support, visible focus, sufficient contrast, and status announcements. A clickable `div` is not a button. Accessibility findings from jsx-a11y and user-behavior tests block merge.

Manage dialogs, menus, and focus with proven accessible patterns. Do not add ARIA when native HTML already supplies correct semantics.

### Errors And Suspense

Place error boundaries around recoverable route or feature regions and one final application boundary. Log safe context once and offer a user action. Do not render stacks or raw dependency errors.

Use Suspense only where the selected data or code-loading mechanism has a proven contract. Loading UI avoids layout collapse and retains accessible status.

React still requires a class boundary for render errors; keep it small and place Suspense inside the recovery boundary:

```tsx
import { Component, lazy, Suspense, type ErrorInfo, type ReactNode } from "react";
import { reportRenderError } from "../telemetry/report-render-error.js";

const WidgetRoute = lazy(() => import("./widget-route.js"));

class RouteErrorBoundary extends Component<
  Readonly<{ children: ReactNode; onError: (error: Error, info: ErrorInfo) => void }>,
  Readonly<{ failed: boolean }>
> {
  override state = { failed: false };

  static getDerivedStateFromError(): Readonly<{ failed: boolean }> {
    return { failed: true };
  }

  override componentDidCatch(error: Error, info: ErrorInfo): void {
    this.props.onError(error, info);
  }

  override render(): ReactNode {
    if (this.state.failed) return <p role="alert">This route could not load.</p>;
    return this.props.children;
  }
}

export function WidgetPage(): ReactNode {
  return (
    <RouteErrorBoundary onError={reportRenderError}>
      <Suspense fallback={<p role="status">Loading widget…</p>}>
        <WidgetRoute />
      </Suspense>
    </RouteErrorBoundary>
  );
}
```

### Assets And Security

Vite production builds fingerprint assets. Serve HTML with revalidation/no-cache policy and immutable fingerprinted assets with long cache lifetime. Values exposed through Vite are public; never put secrets in frontend configuration.

Avoid unsafe HTML. If product requirements demand it, sanitize through one reviewed adapter and test XSS payloads. CSP and other browser controls live with deployment and security policy.

## Common Mistakes And Forbidden Patterns

- Effects used for derived values or ordinary event handling.
- Fetching server state directly in components instead of TanStack Query.
- Global context storing rapidly changing or unrelated state.
- Missing effect cleanup, ignored promises, or `isMounted` flags instead of abort.
- Authentication-only route guards presented as authorization.
- Inaccessible custom controls or test IDs replacing accessible names.
- Secrets in `VITE_*` variables or unsafe HTML without sanitization.

## Verification And Proof

- RTL and user-event tests cover keyboard and accessible-name behavior.
- Hook tests prove rerender, dependency changes, unmount cleanup, and abort where applicable.
- Route tests cover navigation, deep links, not-found, forbidden, lazy loading, and error boundaries.
- Form tests cover client validation, server rejection, focus, resubmission, and duplicate clicks.
- `eslint-plugin-react-hooks` and `eslint-plugin-jsx-a11y` pass with no warnings.
- Vite production build loads under the deployment base path and contains no secrets.

Related: [frontend-data-and-state.md](frontend-data-and-state.md), [../quality/testing.md](../quality/testing.md), and [../operations/deployment.md](../operations/deployment.md).
