import type { Logger } from "pino";

import type { Clock } from "../core/clock.js";
import { isAbortError, TransientError } from "../core/errors.js";
import type { EventProcessor } from "../core/events.js";
import type { Metrics } from "../telemetry/metrics.js";
import { fullJitterDelayMs, type RandomSource, type Waiter } from "./backoff.js";
import type { Broker, BrokerMessage } from "./broker.js";
import { createDeadLetter, type DeadLetterStore, type FailureClass } from "./dlq.js";
import { parseEnvelope, toWidgetEvent, type EventEnvelope } from "./envelope.js";
import type { InboxStore } from "./inbox.js";

export type ConsumerOptions = Readonly<{
  topic: string;
  consumerName: string;
  concurrency: number;
  maxAttempts: number;
  baseBackoffMs: number;
  maxBackoffMs: number;
  broker: Broker;
  processor: EventProcessor;
  inbox: InboxStore;
  deadLetters: DeadLetterStore;
  clock: Clock;
  waiter: Waiter;
  random: RandomSource;
  metrics: Metrics;
  logger: Logger;
}>;

export type RunSignals = Readonly<{
  intake: AbortSignal;
  handlers: AbortSignal;
}>;

const MAX_DELIVERIES_PER_RUN = 1_000_000;

export class Consumer {
  readonly #options: ConsumerOptions;
  readonly #inFlight = new Set<Promise<void>>();
  #maximumObserved = 0;

  constructor(options: ConsumerOptions) {
    assertConsumerOptions(options);
    this.#options = options;
  }

  async run(signals: RunSignals): Promise<void> {
    let deliveries = 0;
    for await (const message of this.#options.broker.subscribe(
      this.#options.topic,
      signals.intake,
    )) {
      if (signals.intake.aborted) {
        await this.#nack(message, "intake stopped");
        break;
      }
      if (this.#inFlight.size >= this.#options.concurrency) {
        await Promise.race(this.#inFlight);
      }
      this.#track(this.processDelivery(message, signals.handlers));
      deliveries += 1;
      if (deliveries >= MAX_DELIVERIES_PER_RUN) {
        throw new Error("consumer delivery bound exceeded");
      }
    }
    await Promise.all(this.#inFlight);
  }

  async processDelivery(message: BrokerMessage, signal: AbortSignal): Promise<void> {
    let envelope: EventEnvelope;
    try {
      envelope = parseEnvelope(message.body);
    } catch (error: unknown) {
      await this.#deadLetter(message, undefined, 1, "validation", error, signal);
      return;
    }
    await this.#processParsed(message, envelope, signal);
  }

  inFlight(): number {
    return this.#inFlight.size;
  }

  maximumObservedConcurrency(): number {
    return this.#maximumObserved;
  }

  async #processParsed(
    message: BrokerMessage,
    envelope: EventEnvelope,
    signal: AbortSignal,
  ): Promise<void> {
    let lastError: unknown;
    for (let attempt = 1; attempt <= this.#options.maxAttempts; attempt += 1) {
      try {
        const outcome = await this.#processOnce(envelope, signal);
        this.#options.metrics.consumed(
          envelope.type,
          outcome === "duplicate" ? "duplicate" : "ack",
        );
        await this.#ack(message);
        return;
      } catch (error: unknown) {
        lastError = error;
        const terminal = await this.#handleFailure(message, envelope, attempt, error, signal);
        if (terminal) return;
      }
    }
    await this.#deadLetter(
      message,
      envelope,
      this.#options.maxAttempts,
      "retry_exhausted",
      lastError,
      signal,
    );
  }

  async #processOnce(envelope: EventEnvelope, signal: AbortSignal) {
    const key = `${this.#options.consumerName}:${envelope.id}`;
    return this.#options.inbox.executeOnce(
      key,
      () => this.#options.processor.process(toWidgetEvent(envelope), signal),
      signal,
    );
  }

  async #handleFailure(
    message: BrokerMessage,
    envelope: EventEnvelope,
    attempt: number,
    error: unknown,
    signal: AbortSignal,
  ): Promise<boolean> {
    if (isCancellation(error, signal)) {
      await this.#nack(message, "handler cancelled");
      return true;
    }
    const failureClass = classifyFailure(error);
    if (failureClass === "permanent") {
      await this.#deadLetter(message, envelope, attempt, failureClass, error, signal);
      return true;
    }
    if (attempt >= this.#options.maxAttempts) return false;
    return this.#waitBeforeRetry(message, envelope, attempt, signal);
  }

  async #waitBeforeRetry(
    message: BrokerMessage,
    envelope: EventEnvelope,
    attempt: number,
    signal: AbortSignal,
  ): Promise<boolean> {
    const delayMs = fullJitterDelayMs(attempt, {
      baseDelayMs: this.#options.baseBackoffMs,
      maxDelayMs: this.#options.maxBackoffMs,
      random: this.#options.random,
    });
    this.#options.metrics.consumed(envelope.type, "retry");
    this.#options.logger.warn(
      {
        deliveryId: message.deliveryId,
        eventId: envelope.id,
        eventType: envelope.type,
        attempt,
        delayMs,
      },
      "retrying message",
    );
    try {
      await this.#options.waiter.wait(delayMs, signal);
      return false;
    } catch (error: unknown) {
      await this.#nack(message, errorMessage(error));
      return true;
    }
  }

  async #deadLetter(
    message: BrokerMessage,
    envelope: EventEnvelope | undefined,
    attempts: number,
    failureClass: FailureClass,
    error: unknown,
    signal: AbortSignal,
  ): Promise<void> {
    const eventType = envelope?.type ?? "unknown";
    const letter = createDeadLetter(
      {
        originalTopic: this.#options.topic,
        consumer: this.#options.consumerName,
        deliveryId: message.deliveryId,
        eventId: envelope?.id,
        eventType,
        schemaVersion: envelope?.version,
        attempts,
        failureClass,
        reason: errorMessage(error),
        body: message.body,
      },
      this.#options.clock,
    );
    try {
      await this.#options.deadLetters.add(letter, signal);
      this.#options.metrics.consumed(eventType, "dead_lettered");
      await this.#ack(message);
    } catch (writeError: unknown) {
      this.#options.logger.error(
        { err: writeError, deliveryId: message.deliveryId },
        "DLQ write failed",
      );
      await this.#nack(message, "DLQ write failed");
    }
  }

  async #ack(message: BrokerMessage): Promise<void> {
    try {
      await message.ack();
    } catch (error: unknown) {
      this.#options.logger.error({ err: error, deliveryId: message.deliveryId }, "ack failed");
    }
  }

  async #nack(message: BrokerMessage, reason: string): Promise<void> {
    this.#options.metrics.consumed("unknown", "nack");
    try {
      await message.nack();
    } catch (error: unknown) {
      this.#options.logger.error(
        { err: error, deliveryId: message.deliveryId, reason },
        "nack failed",
      );
    }
  }

  #track(task: Promise<void>): void {
    const owned = task.finally(() => {
      this.#inFlight.delete(owned);
    });
    this.#inFlight.add(owned);
    this.#maximumObserved = Math.max(this.#maximumObserved, this.#inFlight.size);
  }
}

function classifyFailure(error: unknown): "transient" | "permanent" {
  return error instanceof TransientError ? "transient" : "permanent";
}

function isCancellation(error: unknown, signal: AbortSignal): boolean {
  return signal.aborted || error === signal.reason || isAbortError(error);
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "unknown failure";
}

function assertConsumerOptions(options: ConsumerOptions): void {
  if (
    !Number.isInteger(options.concurrency) ||
    options.concurrency < 1 ||
    options.concurrency > 32
  ) {
    throw new RangeError("concurrency must be an integer in [1, 32]");
  }
  if (
    !Number.isInteger(options.maxAttempts) ||
    options.maxAttempts < 1 ||
    options.maxAttempts > 20
  ) {
    throw new RangeError("maxAttempts must be an integer in [1, 20]");
  }
  if (options.baseBackoffMs < 1 || options.maxBackoffMs < options.baseBackoffMs) {
    throw new RangeError("invalid backoff bounds");
  }
}
