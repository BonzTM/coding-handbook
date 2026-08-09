import { describe, expect, it } from "@jest/globals";
import pino from "pino";

import { TransientError, PermanentError } from "../core/errors.js";
import { WIDGET_CREATED, type EventProcessor, type WidgetEvent } from "../core/events.js";
import { FakeClock } from "../testutil/fake-clock.js";
import { CounterMetrics } from "../telemetry/metrics.js";
import type { Waiter } from "./backoff.js";
import { MemoryBroker, type BrokerMessage } from "./broker.js";
import { Consumer, type ConsumerOptions } from "./consumer.js";
import { MemoryDeadLetterStore } from "./dlq.js";
import { MemoryInboxStore } from "./inbox.js";

const eventId = "550e8400-e29b-41d4-a716-446655440000";
const widgetId = "7d9d1e2e-7239-4b4f-aa48-769a1a886c5d";

class RecordingWaiter implements Waiter {
  readonly delays: number[] = [];

  wait(delayMs: number, signal: AbortSignal): Promise<void> {
    if (signal.aborted) {
      return Promise.reject(
        signal.reason instanceof Error ? signal.reason : new Error("recording wait aborted"),
      );
    }
    this.delays.push(delayMs);
    return Promise.resolve();
  }
}

class ScriptedProcessor implements EventProcessor {
  readonly events: WidgetEvent[] = [];
  readonly #results: readonly (Error | undefined)[];

  constructor(results: readonly (Error | undefined)[] = []) {
    this.#results = results;
  }

  process(event: WidgetEvent, signal: AbortSignal): Promise<void> {
    if (signal.aborted) {
      return Promise.reject(
        signal.reason instanceof Error ? signal.reason : new Error("scripted processor aborted"),
      );
    }
    const result = this.#results[this.events.length];
    this.events.push(event);
    return result === undefined ? Promise.resolve() : Promise.reject(result);
  }
}

class BlockingProcessor implements EventProcessor {
  #calls = 0;
  readonly #startedWaiters = new Map<number, () => void>();
  readonly #release = deferredVoid();

  async process(): Promise<void> {
    this.#calls += 1;
    this.#startedWaiters.get(this.#calls)?.();
    await this.#release.promise;
  }

  async whenStarted(count: number): Promise<void> {
    if (this.#calls >= count) return;
    await new Promise<void>((resolve) => this.#startedWaiters.set(count, resolve));
  }

  release(): void {
    this.#release.resolve();
  }
}

class AbortAwareProcessor implements EventProcessor {
  readonly started = deferredVoid();

  process(event: WidgetEvent, signal: AbortSignal): Promise<void> {
    if (event.widgetId.length === 0) return Promise.reject(new Error("empty widget id"));
    this.started.resolve();
    if (signal.aborted) return Promise.reject(abortReason(signal));
    return new Promise<void>((_resolve, reject) => {
      signal.addEventListener(
        "abort",
        () => {
          reject(abortReason(signal));
        },
        { once: true },
      );
    });
  }
}

describe("Consumer", () => {
  it("processes and acknowledges a valid event", async () => {
    const processor = new ScriptedProcessor();
    const harness = consumerHarness({ processor });
    const delivery = testMessage("delivery-success", eventBody());

    await harness.consumer.processDelivery(delivery.message, activeSignal());

    expect(processor.events).toHaveLength(1);
    expect(delivery.settlement()).toBe("ack");
    expect(harness.inbox.has(`widget-projector:${eventId}`)).toBe(true);
    expect(harness.deadLetters.entries()).toHaveLength(0);
  });

  it("retries a transient failure and then succeeds", async () => {
    const processor = new ScriptedProcessor([new TransientError("dependency unavailable")]);
    const waiter = new RecordingWaiter();
    const harness = consumerHarness({ processor, waiter });
    const delivery = testMessage("delivery-retry", eventBody());

    await harness.consumer.processDelivery(delivery.message, activeSignal());

    expect(processor.events).toHaveLength(2);
    expect(waiter.delays).toEqual([100]);
    expect(delivery.settlement()).toBe("ack");
  });

  it("dead-letters after the transient retry budget is exhausted", async () => {
    const failure = new TransientError("dependency unavailable");
    const processor = new ScriptedProcessor([failure, failure, failure]);
    const waiter = new RecordingWaiter();
    const harness = consumerHarness({ processor, waiter, maxAttempts: 3 });
    const delivery = testMessage("delivery-exhausted", eventBody());

    await harness.consumer.processDelivery(delivery.message, activeSignal());

    expect(processor.events).toHaveLength(3);
    expect(waiter.delays).toEqual([100, 200]);
    expect(delivery.settlement()).toBe("ack");
    expect(harness.deadLetters.entries()[0]).toMatchObject({
      deliveryId: "delivery-exhausted",
      eventId,
      attempts: 3,
      failureClass: "retry_exhausted",
    });
  });

  it("dead-letters poison and permanent failures without retry", async () => {
    const poison = consumerHarness();
    const poisonDelivery = testMessage("delivery-poison", "{not-json");
    await poison.consumer.processDelivery(poisonDelivery.message, activeSignal());

    const permanent = consumerHarness({
      processor: new ScriptedProcessor([new PermanentError("widget is not eligible")]),
    });
    const permanentDelivery = testMessage("delivery-permanent", eventBody());
    await permanent.consumer.processDelivery(permanentDelivery.message, activeSignal());

    expect(poison.deadLetters.entries()[0]?.failureClass).toBe("validation");
    expect(permanent.deadLetters.entries()[0]?.failureClass).toBe("permanent");
    expect(poisonDelivery.settlement()).toBe("ack");
    expect(permanentDelivery.settlement()).toBe("ack");
  });

  it("collapses duplicate event delivery through the consumer-scoped inbox key", async () => {
    const processor = new ScriptedProcessor();
    const harness = consumerHarness({ processor });
    const first = testMessage("delivery-1", eventBody());
    const second = testMessage("delivery-2", eventBody());

    await Promise.all([
      harness.consumer.processDelivery(first.message, activeSignal()),
      harness.consumer.processDelivery(second.message, activeSignal()),
    ]);

    expect(processor.events).toHaveLength(1);
    expect(first.settlement()).toBe("ack");
    expect(second.settlement()).toBe("ack");
  });

  it("stops intake and drains a message already in flight", async () => {
    const broker = new MemoryBroker();
    const processor = new BlockingProcessor();
    const harness = consumerHarness({ broker, processor });
    const intake = new AbortController();
    const handlers = new AbortController();
    await broker.publish(
      "widget.events",
      { deliveryId: "drain-1", body: eventBody() },
      activeSignal(),
    );

    const runTask = harness.consumer.run({ intake: intake.signal, handlers: handlers.signal });
    let finished = false;
    const observed = runTask.then(() => {
      finished = true;
    });
    await processor.whenStarted(1);
    intake.abort(new Error("draining"));
    expect(finished).toBe(false);
    processor.release();

    await expect(broker.waitForSettlement("drain-1")).resolves.toBe("ack");
    await observed;
    expect(harness.consumer.inFlight()).toBe(0);
  });

  it("nacks in-flight work aborted after the drain budget", async () => {
    const broker = new MemoryBroker();
    const processor = new AbortAwareProcessor();
    const harness = consumerHarness({ broker, processor });
    const intake = new AbortController();
    const handlers = new AbortController();
    await broker.publish(
      "widget.events",
      { deliveryId: "abort-1", body: eventBody() },
      activeSignal(),
    );

    const runTask = harness.consumer.run({ intake: intake.signal, handlers: handlers.signal });
    await processor.started.promise;
    intake.abort(new Error("draining"));
    handlers.abort(new Error("drain deadline exceeded"));

    await expect(broker.waitForSettlement("abort-1")).resolves.toBe("nack");
    await runTask;
    expect(harness.deadLetters.entries()).toHaveLength(0);
    expect(harness.inbox.has(`widget-projector:${eventId}`)).toBe(false);
  });

  it("never exceeds configured concurrent handlers", async () => {
    const broker = new MemoryBroker();
    const processor = new BlockingProcessor();
    const harness = consumerHarness({ broker, processor, concurrency: 2 });
    const intake = new AbortController();
    const handlers = new AbortController();
    const ids = [
      "550e8400-e29b-41d4-a716-446655440001",
      "550e8400-e29b-41d4-a716-446655440002",
      "550e8400-e29b-41d4-a716-446655440003",
    ];
    for (let index = 1; index <= 3; index += 1) {
      const id = ids[index - 1];
      if (id === undefined) throw new Error("test event id invariant failed");
      await broker.publish(
        "widget.events",
        { deliveryId: `bounded-${String(index)}`, body: eventBody(String(index), id) },
        activeSignal(),
      );
    }

    const runTask = harness.consumer.run({ intake: intake.signal, handlers: handlers.signal });
    await processor.whenStarted(2);
    expect(harness.consumer.maximumObservedConcurrency()).toBe(2);
    processor.release();
    await processor.whenStarted(3);
    intake.abort(new Error("test complete"));

    await runTask;
    expect(harness.consumer.maximumObservedConcurrency()).toBe(2);
  });
});

type HarnessOverrides = Partial<Omit<ConsumerOptions, "inbox" | "deadLetters">> &
  Readonly<{
    inbox?: MemoryInboxStore;
    deadLetters?: MemoryDeadLetterStore;
  }>;

function consumerHarness(overrides: HarnessOverrides = {}) {
  const inbox = overrides.inbox ?? new MemoryInboxStore();
  const deadLetters = overrides.deadLetters ?? new MemoryDeadLetterStore();
  const options: ConsumerOptions = {
    topic: "widget.events",
    consumerName: "widget-projector",
    concurrency: 2,
    maxAttempts: 4,
    baseBackoffMs: 100,
    maxBackoffMs: 1_000,
    broker: new MemoryBroker(),
    processor: new ScriptedProcessor(),
    inbox,
    deadLetters,
    clock: new FakeClock(new Date("2026-08-09T12:00:00.000Z")),
    waiter: new RecordingWaiter(),
    random: () => 1,
    metrics: new CounterMetrics(),
    logger: pino({ level: "silent" }),
    ...overrides,
  };
  return { consumer: new Consumer(options), inbox, deadLetters };
}

function eventBody(name = "Meter", id = eventId): string {
  return JSON.stringify({
    id,
    type: WIDGET_CREATED,
    version: 1,
    occurredAt: "2026-08-09T11:00:00.000Z",
    producer: "exampleservice",
    payload: { widgetId, tenantId: "tenant-1", name },
  });
}

function testMessage(deliveryId: string, body: string) {
  let result: "ack" | "nack" | undefined;
  const message: BrokerMessage = {
    deliveryId,
    body,
    ack: () => {
      result = "ack";
      return Promise.resolve();
    },
    nack: () => {
      result = "nack";
      return Promise.resolve();
    },
  };
  return { message, settlement: () => result };
}

function activeSignal(): AbortSignal {
  return new AbortController().signal;
}

function abortReason(signal: AbortSignal): Error {
  return signal.reason instanceof Error ? signal.reason : new Error("handler aborted");
}

function deferredVoid() {
  let resolve: (() => void) | undefined;
  const promise = new Promise<void>((complete) => {
    resolve = complete;
  });
  return {
    promise,
    resolve: (): void => {
      if (resolve === undefined) throw new Error("deferred invariant failed");
      resolve();
    },
  };
}
