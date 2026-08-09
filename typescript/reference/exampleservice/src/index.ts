import type { Logger } from "pino";
import { LocalDevAuthenticator } from "./api/auth.js";
import { buildHttpApp } from "./api/app.js";
import type { Config } from "./config/index.js";
import { SystemClock } from "./core/clock.js";
import { WidgetService } from "./core/widget-service.js";
import { createPool } from "./db/pool.js";
import { PostgresWidgetRepository } from "./db/postgres-widget-repository.js";
import { PinoAuditSink } from "./telemetry/audit.js";
import { createAuditLogger } from "./telemetry/logger.js";
import { Readiness } from "./telemetry/readiness.js";
import { NoopTelemetry, type TelemetryLifecycle } from "./telemetry/tracing.js";

export type Application = Readonly<{
  start(): Promise<void>;
  stop(timeoutMs: number): Promise<void>;
}>;

export type ApplicationOptions = Readonly<{
  config: Config;
  logger: Logger;
  signal: AbortSignal;
  telemetry?: TelemetryLifecycle;
}>;

export function buildApplication(options: ApplicationOptions): Application {
  const pool = createPool({
    databaseUrl: options.config.databaseUrl,
    maxConnections: options.config.databasePoolSize,
    timeoutMs: options.config.databaseTimeoutMs,
  });
  const repository = new PostgresWidgetRepository(pool);
  const clock = new SystemClock();
  const readiness = new Readiness();
  const widgets = new WidgetService(
    repository,
    clock,
    options.config.idempotencyTtlMs,
  );
  const authenticator = new LocalDevAuthenticator(
    options.config.authEnabled ? options.config.authToken : undefined,
  );
  const audit = new PinoAuditSink(createAuditLogger(options.config.logLevel));
  const telemetry = options.telemetry ?? new NoopTelemetry();
  const app = buildHttpApp({
    logger: options.logger,
    widgets,
    authenticator,
    audit,
    clock,
    readiness,
    lifetimeSignal: options.signal,
    ready: (signal) => repository.ready({ signal }),
    readinessTimeoutMs: options.config.databaseTimeoutMs,
  });
  return createApplication(
    app,
    readiness,
    telemetry,
    () => pool.end(),
    options.config,
  );
}

function createApplication(
  app: HttpLifecycle,
  readiness: Readiness,
  telemetry: TelemetryLifecycle,
  closePool: () => Promise<void>,
  config: Config,
): Application {
  let stopPromise: Promise<void> | undefined;
  return {
    start: async () => {
      await app.listen({ port: config.port, host: config.host });
      readiness.markReady();
    },
    stop: (timeoutMs) => {
      if (stopPromise !== undefined) {
        return stopPromise;
      }
      readiness.markDraining();
      stopPromise = stopWithinDeadline(app, closePool, telemetry, timeoutMs);
      return stopPromise;
    },
  };
}

async function stopWithinDeadline(
  app: HttpLifecycle,
  closePool: () => Promise<void>,
  telemetry: TelemetryLifecycle,
  timeoutMs: number,
): Promise<void> {
  if (!Number.isInteger(timeoutMs) || timeoutMs < 1) {
    throw new RangeError("shutdown timeout must be a positive integer");
  }
  const controller = new AbortController();
  let timer: NodeJS.Timeout | undefined;
  const deadline = new Promise<never>((_resolve, reject) => {
    timer = setTimeout(() => {
      const error = new Error("shutdown deadline exceeded");
      controller.abort(error);
      reject(error);
    }, timeoutMs);
  });
  const shutdown = closeResources(app, closePool, telemetry, controller.signal);
  try {
    await Promise.race([shutdown, deadline]);
  } finally {
    if (timer !== undefined) {
      clearTimeout(timer);
    }
  }
}

async function closeResources(
  app: HttpLifecycle,
  closePool: () => Promise<void>,
  telemetry: TelemetryLifecycle,
  signal: AbortSignal,
): Promise<void> {
  await app.close();
  await closePool();
  await telemetry.shutdown(signal);
}

type HttpLifecycle = Readonly<{
  listen(options: {
    readonly port: number;
    readonly host: string;
  }): Promise<string>;
  close(): Promise<void>;
}>;
