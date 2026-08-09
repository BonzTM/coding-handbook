import { describe, expect, it } from "@jest/globals";

import { FakeClock } from "./fake-clock.js";

describe("FakeClock", () => {
  it("returns defensive Date values and advances deterministically", () => {
    const clock = new FakeClock(new Date("2026-08-09T12:00:00.000Z"));
    const first = clock.now();
    first.setUTCFullYear(2000);
    clock.advance(250);

    expect(clock.now()).toEqual(new Date("2026-08-09T12:00:00.250Z"));
  });

  it("rejects invalid advances", () => {
    const clock = new FakeClock(new Date(0));
    expect(() => {
      clock.advance(-1);
    }).toThrow(RangeError);
  });
});
