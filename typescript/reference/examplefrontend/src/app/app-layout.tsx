import { Outlet, useRouteError, Link } from "react-router-dom";
import type { ReactNode } from "react";
import { RouteErrorBoundary } from "./route-error-boundary.js";

export function AppLayout(): ReactNode {
  return (
    <>
      <header>
        <nav aria-label="Main navigation">
          <Link to="/">Widget administration</Link>
        </nav>
      </header>
      <main>
        <RouteErrorBoundary>
          <Outlet />
        </RouteErrorBoundary>
      </main>
    </>
  );
}

export function RouterErrorPage(): ReactNode {
  const error: unknown = useRouteError();
  const notFound = isNotFoundResponse(error);
  return (
    <main>
      <h1>{notFound ? "Page not found" : "Something went wrong"}</h1>
      <p role="alert">
        {notFound
          ? "The requested page does not exist."
          : "The page could not be loaded."}
      </p>
      <Link to="/">Return to widgets</Link>
    </main>
  );
}

function isNotFoundResponse(error: unknown): boolean {
  return error instanceof Response && error.status === 404;
}
