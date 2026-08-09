import type { Clock } from "../core/clock.js";

export type FailureClass = "validation" | "permanent" | "retry_exhausted";

export type DeadLetter = Readonly<{
  originalTopic: string;
  consumer: string;
  deliveryId: string;
  eventId: string | undefined;
  eventType: string;
  schemaVersion: number | undefined;
  attempts: number;
  failureClass: FailureClass;
  reason: string;
  body: string;
  deadLetteredAt: Date;
}>;

export interface DeadLetterStore {
  add(letter: DeadLetter, signal: AbortSignal): Promise<void>;
}

export class MemoryDeadLetterStore implements DeadLetterStore {
  readonly #entries: DeadLetter[] = [];

  add(letter: DeadLetter, signal: AbortSignal): Promise<void> {
    if (signal.aborted) {
      const error = signal.reason instanceof Error ? signal.reason : new Error("DLQ write aborted");
      return Promise.reject(error);
    }
    this.#entries.push(Object.freeze({ ...letter }));
    return Promise.resolve();
  }

  entries(): readonly DeadLetter[] {
    return this.#entries.map((entry) => Object.freeze({ ...entry }));
  }
}

export function createDeadLetter(
  input: Omit<DeadLetter, "deadLetteredAt">,
  clock: Clock,
): DeadLetter {
  return Object.freeze({ ...input, deadLetteredAt: clock.now() });
}
