import { QueryClientProvider, type QueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { RouterProvider, type RouterProviderProps } from "react-router-dom";
import type { WidgetsApi } from "../features/widgets/api/widgets-api.js";
import { WidgetsApiContext } from "../features/widgets/widgets-context.js";

type AppProvidersProps = Readonly<{
  api: WidgetsApi;
  queryClient: QueryClient;
  router: RouterProviderProps["router"];
}>;

export function AppProviders({
  api,
  queryClient,
  router,
}: AppProvidersProps): ReactNode {
  return (
    <QueryClientProvider client={queryClient}>
      <WidgetsApiContext value={api}>
        <RouterProvider router={router} />
      </WidgetsApiContext>
    </QueryClientProvider>
  );
}
