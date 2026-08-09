import { createContext, useContext } from "react";
import type { WidgetsApi } from "./api/widgets-api.js";

export const WidgetsApiContext = createContext<WidgetsApi | undefined>(
  undefined,
);

export function useWidgetsApi(): WidgetsApi {
  const api = useContext(WidgetsApiContext);
  if (api === undefined) {
    throw new Error("WidgetsApi provider is missing");
  }
  return api;
}
