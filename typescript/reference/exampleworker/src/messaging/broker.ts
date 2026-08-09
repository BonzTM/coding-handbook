export type Settlement = "ack" | "nack";

export type BrokerMessage = Readonly<{
  deliveryId: string;
  body: string;
  ack(): Promise<void>;
  nack(): Promise<void>;
}>;

export type PublishedMessage = Readonly<{
  deliveryId: string;
  body: string;
}>;

export interface Broker {
  publish(topic: string, message: PublishedMessage, signal: AbortSignal): Promise<void>;
  subscribe(topic: string, signal: AbortSignal): AsyncIterable<BrokerMessage>;
  isHealthy(): boolean;
  close(): Promise<void>;
}

type Deferred<T> = Readonly<{
  promise: Promise<T>;
  resolve(value: T): void;
}>;

interface TopicState {
  readonly messages: PublishedMessage[];
  readonly waiting: Deferred<PublishedMessage | undefined>[];
}

const MAX_DELIVERIES_PER_SUBSCRIPTION = 1_000_000;

export class MemoryBroker implements Broker {
  readonly #topics = new Map<string, TopicState>();
  readonly #settlements = new Map<string, Deferred<Settlement>>();
  readonly #capacity: number;
  #closed = false;
  #healthy = true;

  constructor(capacity = 1_000) {
    if (!Number.isInteger(capacity) || capacity < 1 || capacity > 10_000) {
      throw new RangeError("capacity must be an integer in [1, 10000]");
    }
    this.#capacity = capacity;
  }

  publish(topic: string, message: PublishedMessage, signal: AbortSignal): Promise<void> {
    const failure = publishFailure(signal, this.#closed);
    if (failure !== undefined) return Promise.reject(failure);
    const state = this.#topic(topic);
    const waiter = state.waiting.shift();
    this.#settlements.set(message.deliveryId, deferred<Settlement>());
    if (waiter !== undefined) {
      waiter.resolve(message);
      return Promise.resolve();
    }
    if (state.messages.length >= this.#capacity) {
      throw new Error("broker topic capacity exceeded");
    }
    state.messages.push(message);
    return Promise.resolve();
  }

  async *subscribe(topic: string, signal: AbortSignal): AsyncIterable<BrokerMessage> {
    for (let count = 0; count < MAX_DELIVERIES_PER_SUBSCRIPTION; count += 1) {
      const message = await this.#next(topic, signal);
      if (message === undefined) return;
      yield this.#delivery(topic, message);
    }
    throw new Error("subscription delivery bound exceeded");
  }

  async waitForSettlement(deliveryId: string): Promise<Settlement> {
    const result = this.#settlements.get(deliveryId);
    if (result === undefined) throw new Error(`unknown delivery ${deliveryId}`);
    return result.promise;
  }

  setHealthy(healthy: boolean): void {
    this.#healthy = healthy;
  }

  isHealthy(): boolean {
    return this.#healthy && !this.#closed;
  }

  close(): Promise<void> {
    if (this.#closed) return Promise.resolve();
    this.#closed = true;
    for (const state of this.#topics.values()) {
      for (const waiter of state.waiting.splice(0)) waiter.resolve(undefined);
    }
    return Promise.resolve();
  }

  async #next(topic: string, signal: AbortSignal): Promise<PublishedMessage | undefined> {
    if (this.#closed || signal.aborted) return undefined;
    const state = this.#topic(topic);
    const available = state.messages.shift();
    if (available !== undefined) return available;
    const waiting = deferred<PublishedMessage | undefined>();
    state.waiting.push(waiting);
    return waitForAbort(waiting, state.waiting, signal);
  }

  #delivery(topic: string, message: PublishedMessage): BrokerMessage {
    return Object.freeze({
      ...message,
      ack: () => {
        this.#settle(message.deliveryId, "ack");
        return Promise.resolve();
      },
      nack: () => {
        this.#settle(message.deliveryId, "nack");
        if (!this.#closed) this.#topic(topic).messages.push(message);
        return Promise.resolve();
      },
    });
  }

  #settle(deliveryId: string, settlement: Settlement): void {
    this.#settlements.get(deliveryId)?.resolve(settlement);
  }

  #topic(topic: string): TopicState {
    const existing = this.#topics.get(topic);
    if (existing !== undefined) return existing;
    const created: TopicState = { messages: [], waiting: [] };
    this.#topics.set(topic, created);
    return created;
  }
}

async function waitForAbort(
  waiting: Deferred<PublishedMessage | undefined>,
  queue: Deferred<PublishedMessage | undefined>[],
  signal: AbortSignal,
): Promise<PublishedMessage | undefined> {
  const abort = (): void => {
    const index = queue.indexOf(waiting);
    if (index >= 0) queue.splice(index, 1);
    waiting.resolve(undefined);
  };
  signal.addEventListener("abort", abort, { once: true });
  try {
    return await waiting.promise;
  } finally {
    signal.removeEventListener("abort", abort);
  }
}

function publishFailure(signal: AbortSignal, closed: boolean): Error | undefined {
  if (closed) return new Error("broker is closed");
  if (!signal.aborted) return undefined;
  return signal.reason instanceof Error ? signal.reason : new Error("publish aborted");
}

function deferred<T>(): Deferred<T> {
  let complete: ((value: T) => void) | undefined;
  const promise = new Promise<T>((resolve) => {
    complete = resolve;
  });
  return {
    promise,
    resolve: (value: T): void => {
      if (complete === undefined) throw new Error("deferred was not initialized");
      complete(value);
    },
  };
}
