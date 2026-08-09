import { z } from "zod";

const integerString = (minimum: number, maximum: number) =>
  z.string().regex(/^\d+$/).transform(Number).pipe(z.int().min(minimum).max(maximum));

const envSchema = z
  .strictObject({
    NODE_ENV: z.enum(["development", "test", "production"]).default("development"),
    HOST: z.string().min(1).default("127.0.0.1"),
    HEALTH_PORT: integerString(1, 65_535).prefault("3001"),
    LOG_LEVEL: z.enum(["debug", "info", "warn", "error"]).default("info"),
    TOPIC: z.string().min(1).max(200).default("widget.events"),
    CONSUMER_NAME: z.string().min(1).max(100).default("widget-projector"),
    CONSUMER_CONCURRENCY: integerString(1, 32).prefault("4"),
    CONSUMER_MAX_ATTEMPTS: integerString(1, 20).prefault("5"),
    CONSUMER_BASE_BACKOFF_MS: integerString(1, 60_000).prefault("100"),
    CONSUMER_MAX_BACKOFF_MS: integerString(1, 300_000).prefault("30000"),
    OUTBOX_POLL_INTERVAL_MS: integerString(1, 60_000).prefault("1000"),
    OUTBOX_BATCH_SIZE: integerString(1, 1_000).prefault("100"),
    SHUTDOWN_TIMEOUT_MS: integerString(1, 120_000).prefault("15000"),
  })
  .superRefine((value, context) => {
    if (value.CONSUMER_MAX_BACKOFF_MS < value.CONSUMER_BASE_BACKOFF_MS) {
      context.addIssue({
        code: "custom",
        path: ["CONSUMER_MAX_BACKOFF_MS"],
        message: "must be greater than or equal to CONSUMER_BASE_BACKOFF_MS",
      });
    }
  });

export type Config = Readonly<z.infer<typeof envSchema>>;

export function loadConfig(environment: NodeJS.ProcessEnv): Config {
  return Object.freeze(
    envSchema.parse({
      NODE_ENV: environment.NODE_ENV,
      HOST: environment.HOST,
      HEALTH_PORT: environment.HEALTH_PORT,
      LOG_LEVEL: environment.LOG_LEVEL,
      TOPIC: environment.TOPIC,
      CONSUMER_NAME: environment.CONSUMER_NAME,
      CONSUMER_CONCURRENCY: environment.CONSUMER_CONCURRENCY,
      CONSUMER_MAX_ATTEMPTS: environment.CONSUMER_MAX_ATTEMPTS,
      CONSUMER_BASE_BACKOFF_MS: environment.CONSUMER_BASE_BACKOFF_MS,
      CONSUMER_MAX_BACKOFF_MS: environment.CONSUMER_MAX_BACKOFF_MS,
      OUTBOX_POLL_INTERVAL_MS: environment.OUTBOX_POLL_INTERVAL_MS,
      OUTBOX_BATCH_SIZE: environment.OUTBOX_BATCH_SIZE,
      SHUTDOWN_TIMEOUT_MS: environment.SHUTDOWN_TIMEOUT_MS,
    }),
  );
}
