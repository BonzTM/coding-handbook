import { describe, expect, it } from "@jest/globals";
import { FakeClock } from "./fake-clock.js";

describe("FakeClock", () => {
  it("returns defensive copies and advances deterministically", () => {
    const clock = new FakeClock(new Date("2026-08-08T12:00:00.000Z"));
    const first = clock.now();
    first.setUTCFullYear(1999);

    clock.advance(1_000);

    expect(clock.now().toISOString()).toBe("2026-08-08T12:00:01.000Z");
  });
});
