import type { Clock } from "../core/clock.js";

export class FakeClock implements Clock {
  #instant: Date;

  constructor(instant: Date) {
    this.#instant = new Date(instant.getTime());
  }

  now(): Date {
    return new Date(this.#instant.getTime());
  }

  advance(milliseconds: number): void {
    if (!Number.isInteger(milliseconds)) {
      throw new RangeError("clock advance must use whole milliseconds");
    }
    this.#instant = new Date(this.#instant.getTime() + milliseconds);
  }
}
