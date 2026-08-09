import { describe, expect, it } from "@jest/globals";

import { loadConfig } from "./index.js";

describe("loadConfig", () => {
  it("parses defaults and bounded numeric overrides", () => {
    const config = loadConfig({ NODE_ENV: "test", CONSUMER_CONCURRENCY: "8" });
    expect(config.CONSUMER_CONCURRENCY).toBe(8);
    expect(config.HEALTH_PORT).toBe(3_001);
  });

  it("rejects malformed and inconsistent backoff values", () => {
    expect(() => loadConfig({ NODE_ENV: "test", CONSUMER_CONCURRENCY: "unbounded" })).toThrow();
    expect(() =>
      loadConfig({
        NODE_ENV: "test",
        CONSUMER_BASE_BACKOFF_MS: "200",
        CONSUMER_MAX_BACKOFF_MS: "100",
      }),
    ).toThrow(/CONSUMER_MAX_BACKOFF_MS/);
  });
});
