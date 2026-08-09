import type { Pool, PoolClient } from "pg";

export async function inTransaction<T>(
  pool: Pool,
  signal: AbortSignal,
  operation: (client: PoolClient) => Promise<T>,
): Promise<T> {
  signal.throwIfAborted();
  const client = await pool.connect();
  try {
    signal.throwIfAborted();
    await client.query("BEGIN");
    const value = await operation(client);
    signal.throwIfAborted();
    await client.query("COMMIT");
    return value;
  } catch (error: unknown) {
    try {
      await client.query("ROLLBACK");
    } catch (rollbackError: unknown) {
      throw new AggregateError(
        [error, rollbackError],
        "transaction and rollback failed",
        { cause: rollbackError },
      );
    }
    throw error;
  } finally {
    client.release();
  }
}
