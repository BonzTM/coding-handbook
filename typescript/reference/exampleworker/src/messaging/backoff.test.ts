import { describe, expect, it } from "@jest/globals";

import { backoffCeilingMs, fullJitterDelayMs } from "./backoff.js";

describe("bounded exponential backoff", () => {
  it("doubles from the base and caps at the maximum", () => {
    expect([1, 2, 3, 4, 5].map((attempt) => backoffCeilingMs(attempt, 100, 500))).toEqual([
      100, 200, 400, 500, 500,
    ]);
  });

  it("applies deterministic full jitter within the ceiling", () => {
    expect(fullJitterDelayMs(3, { baseDelayMs: 100, maxDelayMs: 1_000, random: () => 0.25 })).toBe(
      100,
    );
    expect(fullJitterDelayMs(3, { baseDelayMs: 100, maxDelayMs: 1_000, random: () => 1 })).toBe(
      400,
    );
  });

  it("rejects invalid policy input and random output", () => {
    expect(() => backoffCeilingMs(0, 100, 1_000)).toThrow(RangeError);
    expect(() =>
      fullJitterDelayMs(1, { baseDelayMs: 100, maxDelayMs: 1_000, random: () => 2 }),
    ).toThrow(RangeError);
  });
});
