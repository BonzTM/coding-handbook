import { Pool } from "pg";

export type PoolSettings = Readonly<{
  databaseUrl: string;
  maxConnections: number;
  timeoutMs: number;
}>;

export function createPool(settings: PoolSettings): Pool {
  if (settings.maxConnections < 1 || settings.timeoutMs < 1) {
    throw new RangeError("pool settings must be positive");
  }
  return new Pool({
    connectionString: settings.databaseUrl,
    application_name: "typescript-exampleservice",
    max: settings.maxConnections,
    connectionTimeoutMillis: settings.timeoutMs,
    idleTimeoutMillis: 30_000,
    statement_timeout: settings.timeoutMs,
    query_timeout: settings.timeoutMs + 1_000,
    lock_timeout: Math.min(settings.timeoutMs, 1_000),
    idle_in_transaction_session_timeout: settings.timeoutMs,
  });
}
