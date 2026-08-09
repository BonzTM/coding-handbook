import type { Logger } from "pino";

import type { Clock } from "../core/clock.js";
import type { Metrics } from "../telemetry/metrics.js";
import type { Waiter } from "./backoff.js";
import type { Broker } from "./broker.js";
import { parseEnvelope } from "./envelope.js";

export type OutboxRecord = Readonly<{
  id: string;
  topic: string;
  eventType: string;
  body: string;
  createdAt: Date;
  sentAt?: Date;
}>;

export interface OutboxStore {
  add(record: OutboxRecord, signal: AbortSignal): Promise<void>;
  pending(limit: number, signal: AbortSignal): Promise<readonly OutboxRecord[]>;
  markSent(id: string, sentAt: Date, signal: AbortSignal): Promise<void>;
}

export class MemoryOutboxStore implements OutboxStore {
  readonly #order: string[] = [];
  readonly #records = new Map<string, OutboxRecord>();

  add(record: OutboxRecord, signal: AbortSignal): Promise<void> {
    const aborted = abortFailure(signal, "outbox add aborted");
    if (aborted !== undefined) return Promise.reject(aborted);
    if (!this.#records.has(record.id)) this.#order.push(record.id);
    this.#records.set(record.id, Object.freeze({ ...record }));
    return Promise.resolve();
  }

  pending(limit: number, signal: AbortSignal): Promise<readonly OutboxRecord[]> {
    const aborted = abortFailure(signal, "outbox read aborted");
    if (aborted !== undefined) return Promise.reject(aborted);
    const pending: OutboxRecord[] = [];
    for (let index = 0; index < this.#order.length && pending.length < limit; index += 1) {
      const id = this.#order[index];
      if (id === undefined) throw new Error("outbox order invariant failed");
      const record = this.#records.get(id);
      if (record === undefined) throw new Error("outbox record invariant failed");
      if (record.sentAt === undefined) pending.push(record);
    }
    return Promise.resolve(pending);
  }

  markSent(id: string, sentAt: Date, signal: AbortSignal): Promise<void> {
    const aborted = abortFailure(signal, "outbox mark-sent aborted");
    if (aborted !== undefined) return Promise.reject(aborted);
    const record = this.#records.get(id);
    if (record === undefined) return Promise.reject(new Error(`unknown outbox record ${id}`));
    this.#records.set(id, Object.freeze({ ...record, sentAt }));
    return Promise.resolve();
  }

  pendingCount(): number {
    return this.#order.reduce((count, id) => {
      const record = this.#records.get(id);
      return count + (record?.sentAt === undefined ? 1 : 0);
    }, 0);
  }
}

export type OutboxRelayOptions = Readonly<{
  pollIntervalMs: number;
  batchSize: number;
  store: OutboxStore;
  broker: Broker;
  clock: Clock;
  waiter: Waiter;
  metrics: Metrics;
  logger: Logger;
}>;

const MAX_RELAY_CYCLES = 1_000_000;

export class OutboxRelay {
  readonly #options: OutboxRelayOptions;

  constructor(options: OutboxRelayOptions) {
    this.#options = options;
  }

  async flush(signal: AbortSignal): Promise<number> {
    const records = await this.#options.store.pending(this.#options.batchSize, signal);
    let published = 0;
    for (const record of records) {
      await this.#publish(record, signal);
      published += 1;
    }
    return published;
  }

  async run(signal: AbortSignal): Promise<void> {
    try {
      for (let cycle = 0; cycle < MAX_RELAY_CYCLES && !signal.aborted; cycle += 1) {
        await this.#options.waiter.wait(this.#options.pollIntervalMs, signal);
        await this.#flushOrLog(signal);
      }
    } catch (error: unknown) {
      if (!signal.aborted) throw error;
    }
    if (!signal.aborted) throw new Error("outbox relay cycle bound exceeded");
  }

  async #publish(record: OutboxRecord, signal: AbortSignal): Promise<void> {
    const envelope = parseEnvelope(record.body);
    await this.#options.broker.publish(
      record.topic,
      { deliveryId: record.id, body: record.body },
      signal,
    );
    await this.#options.store.markSent(record.id, this.#options.clock.now(), signal);
    this.#options.metrics.published(envelope.type);
  }

  async #flushOrLog(signal: AbortSignal): Promise<void> {
    try {
      await this.flush(signal);
    } catch (error: unknown) {
      if (!signal.aborted) this.#options.logger.warn({ err: error }, "outbox relay scan failed");
    }
  }
}

function abortFailure(signal: AbortSignal, fallback: string): Error | undefined {
  if (!signal.aborted) return undefined;
  return signal.reason instanceof Error ? signal.reason : new Error(fallback);
}
