import { describe, expect, it } from "@jest/globals";
import pino from "pino";

import { FakeClock } from "../testutil/fake-clock.js";
import { CounterMetrics } from "../telemetry/metrics.js";
import type { Waiter } from "./backoff.js";
import type { Broker, BrokerMessage, PublishedMessage } from "./broker.js";
import { parseEnvelope } from "./envelope.js";
import { MemoryInboxStore } from "./inbox.js";
import { MemoryOutboxStore, OutboxRelay, type OutboxRecord, type OutboxStore } from "./outbox.js";

class NoSleepWaiter implements Waiter {
  wait(delayMs: number, signal: AbortSignal): Promise<void> {
    if (delayMs < 0) return Promise.reject(new RangeError("negative test delay"));
    if (signal.aborted) {
      return Promise.reject(
        signal.reason instanceof Error ? signal.reason : new Error("test wait aborted"),
      );
    }
    return Promise.resolve();
  }
}

class RecordingBroker implements Broker {
  readonly published: PublishedMessage[] = [];
  failPublish = false;

  publish(topic: string, message: PublishedMessage, signal: AbortSignal): Promise<void> {
    if (topic.length === 0) return Promise.reject(new Error("empty test topic"));
    if (signal.aborted) {
      return Promise.reject(
        signal.reason instanceof Error ? signal.reason : new Error("test publish aborted"),
      );
    }
    if (this.failPublish) return Promise.reject(new Error("broker unavailable"));
    this.published.push(message);
    return Promise.resolve();
  }

  subscribe(): AsyncIterable<BrokerMessage> {
    throw new Error("recording broker does not support subscriptions");
  }

  isHealthy(): boolean {
    return true;
  }

  close(): Promise<void> {
    return Promise.resolve();
  }
}

class MarkFailsOnceStore implements OutboxStore {
  readonly #delegate: MemoryOutboxStore;
  #shouldFail = true;

  constructor(delegate: MemoryOutboxStore) {
    this.#delegate = delegate;
  }

  async add(record: OutboxRecord, signal: AbortSignal): Promise<void> {
    await this.#delegate.add(record, signal);
  }

  async pending(limit: number, signal: AbortSignal): Promise<readonly OutboxRecord[]> {
    return this.#delegate.pending(limit, signal);
  }

  async markSent(id: string, sentAt: Date, signal: AbortSignal): Promise<void> {
    if (this.#shouldFail) {
      this.#shouldFail = false;
      throw new Error("process crashed before mark sent");
    }
    await this.#delegate.markSent(id, sentAt, signal);
  }
}

describe("OutboxRelay", () => {
  it("leaves a committed record pending when publish fails and relays it after restart", async () => {
    const store = new MemoryOutboxStore();
    const failingBroker = new RecordingBroker();
    failingBroker.failPublish = true;
    await store.add(outboxRecord(), activeSignal());

    await expect(relay(store, failingBroker).flush(activeSignal())).rejects.toThrow(
      "broker unavailable",
    );
    expect(store.pendingCount()).toBe(1);

    const recoveredBroker = new RecordingBroker();
    await expect(relay(store, recoveredBroker).flush(activeSignal())).resolves.toBe(1);
    expect(recoveredBroker.published).toHaveLength(1);
    expect(store.pendingCount()).toBe(0);
  });

  it("tolerates duplicate publish after a crash between publish and mark-sent", async () => {
    const durableStore = new MemoryOutboxStore();
    const store = new MarkFailsOnceStore(durableStore);
    const broker = new RecordingBroker();
    await store.add(outboxRecord(), activeSignal());
    const outboxRelay = relay(store, broker);

    await expect(outboxRelay.flush(activeSignal())).rejects.toThrow("crashed");
    await expect(outboxRelay.flush(activeSignal())).resolves.toBe(1);
    expect(broker.published).toHaveLength(2);
    expect(broker.published[0]?.body).toBe(broker.published[1]?.body);

    const inbox = new MemoryInboxStore();
    let effects = 0;
    for (const published of broker.published) {
      const envelope = parseEnvelope(published.body);
      await inbox.executeOnce(
        `widget-projector:${envelope.id}`,
        () => {
          effects += 1;
          return Promise.resolve();
        },
        activeSignal(),
      );
    }
    expect(effects).toBe(1);
  });
});

function relay(store: OutboxStore, broker: Broker): OutboxRelay {
  return new OutboxRelay({
    pollIntervalMs: 1_000,
    batchSize: 10,
    store,
    broker,
    clock: new FakeClock(new Date("2026-08-09T12:00:00.000Z")),
    waiter: new NoSleepWaiter(),
    metrics: new CounterMetrics(),
    logger: pino({ level: "silent" }),
  });
}

function outboxRecord(): OutboxRecord {
  return {
    id: "550e8400-e29b-41d4-a716-446655440000",
    topic: "widget.events",
    eventType: "widget.created",
    body: JSON.stringify({
      id: "550e8400-e29b-41d4-a716-446655440000",
      type: "widget.created",
      version: 1,
      occurredAt: "2026-08-09T11:00:00.000Z",
      producer: "exampleservice",
      payload: {
        widgetId: "7d9d1e2e-7239-4b4f-aa48-769a1a886c5d",
        tenantId: "tenant-1",
        name: "Meter",
      },
    }),
    createdAt: new Date("2026-08-09T11:00:00.000Z"),
  };
}

function activeSignal(): AbortSignal {
  return new AbortController().signal;
}
