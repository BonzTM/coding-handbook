import { z } from "zod/v4";

const integerString = (minimum: number, maximum: number) =>
  z
    .string()
    .regex(/^\d+$/)
    .transform((value) => Number(value))
    .pipe(z.int().min(minimum).max(maximum));

const envSchema = z
  .strictObject({
    NODE_ENV: z.enum(["development", "test", "production"]),
    HOST: z.string().min(1).max(255).default("127.0.0.1"),
    PORT: integerString(1, 65_535).prefault("3000"),
    DATABASE_URL: z.url().refine((value) => value.startsWith("postgres"), {
      error: "DATABASE_URL must use a PostgreSQL scheme",
    }),
    LOG_LEVEL: z.enum(["debug", "info", "warn", "error"]).default("info"),
    AUTH_ENABLED: z
      .enum(["true", "false"])
      .prefault("false")
      .transform((value) => value === "true"),
    AUTH_TOKEN: z.string().min(16).max(512).optional(),
    SHUTDOWN_TIMEOUT_MS: integerString(100, 60_000).prefault("10000"),
    DATABASE_TIMEOUT_MS: integerString(100, 30_000).prefault("5000"),
    DATABASE_POOL_SIZE: integerString(1, 50).prefault("10"),
    IDEMPOTENCY_TTL_MS: integerString(1_000, 604_800_000).prefault("86400000"),
  })
  .superRefine((value, context) => {
    if (value.AUTH_ENABLED && value.AUTH_TOKEN === undefined) {
      context.addIssue({
        code: "custom",
        path: ["AUTH_TOKEN"],
        message: "AUTH_TOKEN is required when AUTH_ENABLED is true",
      });
    }
  });

export type Config = Readonly<{
  nodeEnv: "development" | "test" | "production";
  host: string;
  port: number;
  databaseUrl: string;
  logLevel: "debug" | "info" | "warn" | "error";
  authEnabled: boolean;
  authToken: string | undefined;
  shutdownTimeoutMs: number;
  databaseTimeoutMs: number;
  databasePoolSize: number;
  idempotencyTtlMs: number;
}>;

export function loadConfig(env: NodeJS.ProcessEnv): Config {
  const input = {
    NODE_ENV: env.NODE_ENV,
    HOST: env.HOST,
    PORT: env.PORT,
    DATABASE_URL: env.DATABASE_URL,
    LOG_LEVEL: env.LOG_LEVEL,
    AUTH_ENABLED: env.AUTH_ENABLED,
    AUTH_TOKEN: env.AUTH_TOKEN,
    SHUTDOWN_TIMEOUT_MS: env.SHUTDOWN_TIMEOUT_MS,
    DATABASE_TIMEOUT_MS: env.DATABASE_TIMEOUT_MS,
    DATABASE_POOL_SIZE: env.DATABASE_POOL_SIZE,
    IDEMPOTENCY_TTL_MS: env.IDEMPOTENCY_TTL_MS,
  };
  const parsed = envSchema.parse(input);
  return Object.freeze({
    nodeEnv: parsed.NODE_ENV,
    host: parsed.HOST,
    port: parsed.PORT,
    databaseUrl: parsed.DATABASE_URL,
    logLevel: parsed.LOG_LEVEL,
    authEnabled: parsed.AUTH_ENABLED,
    authToken: parsed.AUTH_TOKEN,
    shutdownTimeoutMs: parsed.SHUTDOWN_TIMEOUT_MS,
    databaseTimeoutMs: parsed.DATABASE_TIMEOUT_MS,
    databasePoolSize: parsed.DATABASE_POOL_SIZE,
    idempotencyTtlMs: parsed.IDEMPOTENCY_TTL_MS,
  });
}
