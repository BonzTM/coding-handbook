import type { Logger } from "pino";

import type { Config } from "./config/index.js";
import { SystemClock } from "./core/clock.js";
import { WidgetProjector } from "./core/projector.js";
import { buildHealthApp } from "./health/app.js";
import { TimerWaiter } from "./messaging/backoff.js";
import { MemoryBroker } from "./messaging/broker.js";
import { Consumer } from "./messaging/consumer.js";
import { MemoryDeadLetterStore } from "./messaging/dlq.js";
import { MemoryInboxStore } from "./messaging/inbox.js";
import { MemoryOutboxStore, OutboxRelay } from "./messaging/outbox.js";
import { CounterMetrics } from "./telemetry/metrics.js";
import { Readiness } from "./telemetry/readiness.js";

export type WorkerApplication = Readonly<{
  run(signal: AbortSignal): Promise<void>;
}>;

export function buildWorker(config: Config, logger: Logger): WorkerApplication {
  const clock = new SystemClock();
  const broker = new MemoryBroker();
  const metrics = new CounterMetrics();
  const readiness = new Readiness();
  const handlerLifetime = new AbortController();
  const intakeLifetime = new AbortController();
  const relayLifetime = new AbortController();
  const health = buildHealthApp({ readiness, broker, metrics });
  const outbox = new MemoryOutboxStore();
  const relay = buildRelay(config, { broker, outbox, clock, metrics, logger });
  const consumer = buildConsumer(config, { broker, clock, metrics, logger });

  return {
    run: async (signal: AbortSignal) => {
      await health.listen({ host: config.HOST, port: config.HEALTH_PORT });
      readiness.set(true);
      const consumerTask = consumer.run({
        intake: intakeLifetime.signal,
        handlers: handlerLifetime.signal,
      });
      const relayTask = relay.run(relayLifetime.signal);
      let failure: unknown;
      try {
        await supervise(signal, consumerTask, relayTask);
      } catch (error: unknown) {
        failure = error;
      } finally {
        await shutdown({
          config,
          readiness,
          intakeLifetime,
          handlerLifetime,
          relayLifetime,
          consumerTask,
          relayTask,
          relay,
          health,
          broker,
          logger,
        });
      }
      if (failure !== undefined) throw toError(failure);
    },
  };
}

type Infrastructure = Readonly<{
  broker: MemoryBroker;
  outbox: MemoryOutboxStore;
  clock: SystemClock;
  metrics: CounterMetrics;
  logger: Logger;
}>;

function buildRelay(config: Config, infrastructure: Infrastructure): OutboxRelay {
  return new OutboxRelay({
    pollIntervalMs: config.OUTBOX_POLL_INTERVAL_MS,
    batchSize: config.OUTBOX_BATCH_SIZE,
    store: infrastructure.outbox,
    broker: infrastructure.broker,
    clock: infrastructure.clock,
    waiter: new TimerWaiter(),
    metrics: infrastructure.metrics,
    logger: infrastructure.logger,
  });
}

function buildConsumer(config: Config, infrastructure: Omit<Infrastructure, "outbox">): Consumer {
  return new Consumer({
    topic: config.TOPIC,
    consumerName: config.CONSUMER_NAME,
    concurrency: config.CONSUMER_CONCURRENCY,
    maxAttempts: config.CONSUMER_MAX_ATTEMPTS,
    baseBackoffMs: config.CONSUMER_BASE_BACKOFF_MS,
    maxBackoffMs: config.CONSUMER_MAX_BACKOFF_MS,
    broker: infrastructure.broker,
    processor: new WidgetProjector(infrastructure.clock),
    inbox: new MemoryInboxStore(),
    deadLetters: new MemoryDeadLetterStore(),
    clock: infrastructure.clock,
    waiter: new TimerWaiter(),
    random: Math.random,
    metrics: infrastructure.metrics,
    logger: infrastructure.logger,
  });
}

type ShutdownOptions = Readonly<{
  config: Config;
  readiness: Readiness;
  intakeLifetime: AbortController;
  handlerLifetime: AbortController;
  relayLifetime: AbortController;
  consumerTask: Promise<void>;
  relayTask: Promise<void>;
  relay: OutboxRelay;
  health: ReturnType<typeof buildHealthApp>;
  broker: MemoryBroker;
  logger: Logger;
}>;

async function shutdown(options: ShutdownOptions): Promise<void> {
  options.readiness.set(false);
  options.intakeLifetime.abort(new Error("worker intake stopped"));
  options.relayLifetime.abort(new Error("outbox relay stopped"));
  let failure: unknown;
  failure = await captureFailure(
    () =>
      drainConsumer(
        options.consumerTask,
        options.handlerLifetime,
        options.config.SHUTDOWN_TIMEOUT_MS,
      ),
    failure,
  );
  failure = await captureFailure(() => options.relayTask, failure);
  failure = await captureFailure(
    () => options.relay.flush(AbortSignal.timeout(options.config.SHUTDOWN_TIMEOUT_MS)),
    failure,
  );
  failure = await captureFailure(() => options.health.close(), failure);
  failure = await captureFailure(() => options.broker.close(), failure);
  options.logger.flush();
  if (failure !== undefined) throw toError(failure);
}

async function drainConsumer(
  consumerTask: Promise<void>,
  handlers: AbortController,
  timeoutMs: number,
): Promise<void> {
  let timer: NodeJS.Timeout | undefined;
  const deadline = new Promise<never>((_resolve, reject) => {
    timer = setTimeout(() => {
      const error = new Error("drain deadline exceeded");
      handlers.abort(error);
      reject(error);
    }, timeoutMs);
  });
  try {
    await Promise.race([consumerTask, deadline]);
  } finally {
    if (timer !== undefined) clearTimeout(timer);
  }
}

async function captureFailure(
  operation: () => Promise<unknown>,
  current: unknown,
): Promise<unknown> {
  try {
    await operation();
    return current;
  } catch (error: unknown) {
    return current ?? error;
  }
}

async function supervise(
  signal: AbortSignal,
  consumerTask: Promise<void>,
  relayTask: Promise<void>,
): Promise<void> {
  await Promise.race([
    waitForAbort(signal),
    failOnUnexpectedCompletion("consumer", consumerTask),
    failOnUnexpectedCompletion("outbox relay", relayTask),
  ]);
}

async function failOnUnexpectedCompletion(name: string, task: Promise<void>): Promise<never> {
  await task;
  throw new Error(`${name} stopped unexpectedly`);
}

async function waitForAbort(signal: AbortSignal): Promise<void> {
  if (signal.aborted) return;
  await new Promise<void>((resolve) => {
    signal.addEventListener(
      "abort",
      () => {
        resolve();
      },
      { once: true },
    );
  });
}

function toError(error: unknown): Error {
  return error instanceof Error ? error : new Error("worker failed with a non-Error value");
}
