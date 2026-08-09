import { loadConfig } from "./config/index.js";
import { buildWorker } from "./index.js";
import { buildLogger } from "./telemetry/logger.js";

async function main(environment: NodeJS.ProcessEnv): Promise<void> {
  const config = loadConfig(environment);
  const logger = buildLogger(config.LOG_LEVEL);
  const lifetime = new AbortController();
  let fatalReason: unknown;
  const stop = (): void => {
    lifetime.abort(new Error("termination signal received"));
  };
  const fail = (reason: unknown, event: string): void => {
    fatalReason = reason;
    logger.fatal({ err: reason, event }, "fatal process error");
    lifetime.abort(reason);
  };
  const reject = (reason: unknown): void => {
    fail(reason, "unhandled_rejection");
  };
  const crash = (error: Error): void => {
    fail(error, "uncaught_exception");
  };
  process.once("SIGINT", stop);
  process.once("SIGTERM", stop);
  process.once("unhandledRejection", reject);
  process.once("uncaughtException", crash);
  try {
    await buildWorker(config, logger).run(lifetime.signal);
    if (fatalReason !== undefined) throw toError(fatalReason);
  } finally {
    process.off("SIGINT", stop);
    process.off("SIGTERM", stop);
    process.off("unhandledRejection", reject);
    process.off("uncaughtException", crash);
  }
}

void main(process.env).catch((error: unknown) => {
  console.error("exampleworker failed", toError(error).message);
  process.exitCode = 1;
});

function toError(error: unknown): Error {
  return error instanceof Error ? error : new Error("process failed with a non-Error value");
}
