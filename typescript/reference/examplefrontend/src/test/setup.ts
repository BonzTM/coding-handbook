import "@testing-library/jest-dom/jest-globals";
import { afterAll, afterEach, beforeAll } from "@jest/globals";
import { cleanup } from "@testing-library/react";
import { resetMockWidgets } from "../mocks/handlers.js";
import { server } from "./server.js";

beforeAll(() => {
  server.listen({ onUnhandledRequest: "error" });
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
  resetMockWidgets();
});

afterAll(() => {
  server.close();
});
