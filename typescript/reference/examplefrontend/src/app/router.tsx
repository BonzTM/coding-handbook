import { lazy, Suspense, type ReactNode } from "react";
import {
  createBrowserRouter,
  createMemoryRouter,
  type RouteObject,
} from "react-router-dom";
import { AsyncStatus } from "../components/async-status.js";
import { WidgetListRoute } from "../routes/widget-list-route.js";
import { AppLayout, RouterErrorPage } from "./app-layout.js";

const LazyWidgetDetailRoute = lazy(
  () => import("../routes/widget-detail-route.js"),
);

function WidgetDetailBoundary(): ReactNode {
  return (
    <Suspense fallback={<AsyncStatus>Loading widget page…</AsyncStatus>}>
      <LazyWidgetDetailRoute />
    </Suspense>
  );
}

const routes: RouteObject[] = [
  {
    path: "/",
    element: <AppLayout />,
    errorElement: <RouterErrorPage />,
    children: [
      { index: true, element: <WidgetListRoute /> },
      { path: "widgets/:widgetId", element: <WidgetDetailBoundary /> },
    ],
  },
];

export function createAppRouter() {
  return createBrowserRouter(routes);
}

export function createTestRouter(initialEntries: readonly string[]) {
  return createMemoryRouter(routes, { initialEntries: [...initialEntries] });
}
