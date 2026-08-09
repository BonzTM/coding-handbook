import { loadConfig } from "./config/index.js";
import { buildApplication } from "./index.js";
import { createLogger } from "./telemetry/logger.js";

async function run(): Promise<void> {
  const config = loadConfig(process.env);
  const logger = createLogger(config.logLevel);
  const lifetime = new AbortController();
  const application = buildApplication({
    config,
    logger,
    signal: lifetime.signal,
  });
  let shutdownPromise: Promise<void> | undefined;

  const shutdown = (exitCode: number, reason: unknown): Promise<void> => {
    if (shutdownPromise !== undefined) {
      return shutdownPromise;
    }
    lifetime.abort(reason);
    shutdownPromise = application.stop(config.shutdownTimeoutMs).then(
      () => process.exit(exitCode),
      (error: unknown) => {
        logger.fatal(
          { err: error, event: "shutdown_failed" },
          "shutdown failed",
        );
        process.exit(1);
      },
    );
    return shutdownPromise;
  };

  for (const signal of ["SIGTERM", "SIGINT"] as const) {
    process.once(signal, () => {
      void shutdown(0, new Error(`received ${signal}`));
    });
  }
  process.once("unhandledRejection", (reason: unknown) => {
    logger.fatal(
      { err: reason, event: "unhandled_rejection" },
      "fatal process error",
    );
    void shutdown(1, reason);
  });
  process.once("uncaughtException", (error: Error) => {
    logger.fatal(
      { err: error, event: "uncaught_exception" },
      "fatal process error",
    );
    void shutdown(1, error);
  });

  await application.start();
  logger.info(
    { event: "service_started", port: config.port },
    "service started",
  );
}

void run().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : "unknown error";
  console.error("startup failed", message);
  process.exitCode = 1;
});
