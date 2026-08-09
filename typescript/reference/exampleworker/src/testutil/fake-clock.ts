import type { Clock } from "../core/clock.js";

export class FakeClock implements Clock {
  #epochMs: number;

  constructor(instant: Date) {
    this.#epochMs = instant.getTime();
  }

  now(): Date {
    return new Date(this.#epochMs);
  }

  advance(delayMs: number): void {
    if (!Number.isFinite(delayMs) || delayMs < 0) {
      throw new RangeError("delayMs must be finite and non-negative");
    }
    this.#epochMs += delayMs;
  }
}
