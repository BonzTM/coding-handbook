import { describe, expect, it } from "@jest/globals";
import { ZodError } from "zod/v4";
import { loadConfig } from "./index.js";

const required = {
  NODE_ENV: "test",
  DATABASE_URL: "postgres://app:app@localhost:5432/app",
};

describe("loadConfig", () => {
  it("parses defaults into an immutable typed value", () => {
    const config = loadConfig(required);

    expect(config.port).toBe(3000);
    expect(config.authEnabled).toBe(false);
    expect(Object.isFrozen(config)).toBe(true);
  });

  it("rejects missing and malformed configuration", () => {
    expect(() => loadConfig({ NODE_ENV: "test" })).toThrow(ZodError);
    expect(() => loadConfig({ ...required, PORT: "12x" })).toThrow(ZodError);
    expect(() => loadConfig({ ...required, PORT: "70000" })).toThrow(ZodError);
  });

  it("requires a secret token only when authentication is enabled", () => {
    expect(() => loadConfig({ ...required, AUTH_ENABLED: "true" })).toThrow(
      ZodError,
    );
    expect(
      loadConfig({
        ...required,
        AUTH_ENABLED: "true",
        AUTH_TOKEN: "sixteen-byte-token",
      }).authEnabled,
    ).toBe(true);
  });
});
