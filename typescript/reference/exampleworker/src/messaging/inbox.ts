export type InboxOutcome = "processed" | "duplicate";

export interface InboxStore {
  executeOnce(
    key: string,
    operation: () => Promise<void>,
    signal: AbortSignal,
  ): Promise<InboxOutcome>;
  has(key: string): boolean;
}

export class MemoryInboxStore implements InboxStore {
  readonly #completed = new Set<string>();
  readonly #inFlight = new Map<string, Promise<void>>();

  async executeOnce(
    key: string,
    operation: () => Promise<void>,
    signal: AbortSignal,
  ): Promise<InboxOutcome> {
    signal.throwIfAborted();
    if (this.#completed.has(key)) return "duplicate";
    const current = this.#inFlight.get(key);
    if (current !== undefined) {
      await current;
      signal.throwIfAborted();
      return "duplicate";
    }
    return this.#executeFirst(key, operation);
  }

  has(key: string): boolean {
    return this.#completed.has(key);
  }

  async #executeFirst(key: string, operation: () => Promise<void>): Promise<InboxOutcome> {
    const task = operation();
    this.#inFlight.set(key, task);
    try {
      await task;
      this.#completed.add(key);
      return "processed";
    } finally {
      this.#inFlight.delete(key);
    }
  }
}
