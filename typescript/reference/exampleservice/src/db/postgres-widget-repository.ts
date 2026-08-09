import { DatabaseError, type Pool, type PoolClient } from "pg";
import type { WidgetCursor } from "../core/cursor.js";
import {
  AppError,
  DependencyError,
  DuplicateWidgetNameError,
  IdempotencyConflictError,
  WidgetNotFoundError,
  WidgetVersionConflictError,
} from "../core/errors.js";
import type {
  CreateResult,
  IdempotencyClaim,
  OperationOptions,
  WidgetRepository,
} from "../core/widget-repository.js";
import type { CreateWidget, UpdateWidget, Widget } from "../core/widget.js";
import type { WidgetId } from "../core/widget-id.js";
import { parseWidgetRow } from "./rows.js";
import { inTransaction } from "./transaction.js";

type UnknownRow = Record<string, unknown>;

export class PostgresWidgetRepository implements WidgetRepository {
  readonly #pool: Pool;

  constructor(pool: Pool) {
    this.#pool = pool;
  }

  async create(
    tenantId: string,
    input: CreateWidget,
    claim: IdempotencyClaim,
    options: OperationOptions,
  ): Promise<CreateResult> {
    try {
      return await inTransaction(this.#pool, options.signal, async (client) => {
        const claimed = await claimIdempotency(client, tenantId, claim);
        if (!claimed) {
          return replayCreate(client, tenantId, claim);
        }
        const widget = await insertWidget(client, tenantId, input);
        await completeIdempotency(client, tenantId, claim.key, widget.id);
        return { widget, replayed: false };
      });
    } catch (error: unknown) {
      throw mapDatabaseError("create widget failed", error);
    }
  }

  async get(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<Widget | null> {
    options.signal.throwIfAborted();
    try {
      const result = await this.#pool.query<UnknownRow>({
        text: `${WIDGET_COLUMNS} WHERE tenant_id = $1 AND id = $2`,
        values: [tenantId, id],
      });
      options.signal.throwIfAborted();
      const row = result.rows[0];
      return row === undefined ? null : parseWidgetRow(row);
    } catch (error: unknown) {
      throw mapDatabaseError("get widget failed", error);
    }
  }

  async list(
    tenantId: string,
    cursor: WidgetCursor | undefined,
    limit: number,
    options: OperationOptions,
  ): Promise<readonly Widget[]> {
    options.signal.throwIfAborted();
    assertLimit(limit);
    try {
      const result =
        cursor === undefined
          ? await this.#pool.query<UnknownRow>({
              text: `${WIDGET_COLUMNS} WHERE tenant_id = $1 ORDER BY created_at, id LIMIT $2`,
              values: [tenantId, limit],
            })
          : await this.#pool.query<UnknownRow>({
              text: `${WIDGET_COLUMNS} WHERE tenant_id = $1 AND (created_at, id) > ($2, $3) ORDER BY created_at, id LIMIT $4`,
              values: [tenantId, cursor.createdAt, cursor.id, limit],
            });
      options.signal.throwIfAborted();
      return result.rows.map(parseWidgetRow);
    } catch (error: unknown) {
      throw mapDatabaseError("list widgets failed", error);
    }
  }

  async update(
    tenantId: string,
    id: WidgetId,
    input: UpdateWidget,
    options: OperationOptions,
  ): Promise<Widget> {
    options.signal.throwIfAborted();
    try {
      const result = await this.#pool.query<UnknownRow>({
        text: `UPDATE widgets SET name = $3, description = $4, updated_at = $5, version = version + 1 WHERE tenant_id = $1 AND id = $2 AND version = $6 RETURNING id, name, description, created_at, updated_at, version`,
        values: [
          tenantId,
          id,
          input.name,
          input.description,
          input.updatedAt,
          input.expectedVersion,
        ],
      });
      options.signal.throwIfAborted();
      const row = result.rows[0];
      if (row !== undefined) {
        return parseWidgetRow(row);
      }
      await this.#throwUpdateMiss(tenantId, id, options);
      throw new WidgetVersionConflictError();
    } catch (error: unknown) {
      throw mapDatabaseError("update widget failed", error);
    }
  }

  async delete(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<void> {
    options.signal.throwIfAborted();
    try {
      const result = await this.#pool.query({
        text: "DELETE FROM widgets WHERE tenant_id = $1 AND id = $2",
        values: [tenantId, id],
      });
      options.signal.throwIfAborted();
      if (result.rowCount !== 1) {
        throw new WidgetNotFoundError();
      }
    } catch (error: unknown) {
      throw mapDatabaseError("delete widget failed", error);
    }
  }

  async ready(options: OperationOptions): Promise<boolean> {
    options.signal.throwIfAborted();
    try {
      await this.#pool.query("SELECT 1");
      options.signal.throwIfAborted();
      return true;
    } catch {
      options.signal.throwIfAborted();
      return false;
    }
  }

  async #throwUpdateMiss(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<void> {
    const existing = await this.get(tenantId, id, options);
    if (existing === null) {
      throw new WidgetNotFoundError();
    }
  }
}

const WIDGET_COLUMNS =
  "SELECT id, name, description, created_at, updated_at, version FROM widgets";

async function claimIdempotency(
  client: PoolClient,
  tenantId: string,
  claim: IdempotencyClaim,
): Promise<boolean> {
  await client.query({
    text: "DELETE FROM idempotency_keys WHERE tenant_id = $1 AND operation = 'widget.create' AND idempotency_key = $2 AND expires_at <= CURRENT_TIMESTAMP",
    values: [tenantId, claim.key],
  });
  const result = await client.query({
    text: "INSERT INTO idempotency_keys (tenant_id, operation, idempotency_key, request_hash, widget_id, expires_at) VALUES ($1, 'widget.create', $2, $3, NULL, $4) ON CONFLICT DO NOTHING",
    values: [tenantId, claim.key, claim.fingerprint, claim.expiresAt],
  });
  return result.rowCount === 1;
}

async function replayCreate(
  client: PoolClient,
  tenantId: string,
  claim: IdempotencyClaim,
): Promise<CreateResult> {
  const result = await client.query<UnknownRow>({
    text: "SELECT request_hash, widget_id FROM idempotency_keys WHERE tenant_id = $1 AND operation = 'widget.create' AND idempotency_key = $2 FOR UPDATE",
    values: [tenantId, claim.key],
  });
  const row = result.rows[0];
  if (
    row?.request_hash !== claim.fingerprint ||
    typeof row.widget_id !== "string"
  ) {
    throw new IdempotencyConflictError();
  }
  const widgetResult = await client.query<UnknownRow>({
    text: `${WIDGET_COLUMNS} WHERE tenant_id = $1 AND id = $2`,
    values: [tenantId, row.widget_id],
  });
  const widgetRow = widgetResult.rows[0];
  if (widgetRow === undefined) {
    throw new WidgetNotFoundError();
  }
  return { widget: parseWidgetRow(widgetRow), replayed: true };
}

async function insertWidget(
  client: PoolClient,
  tenantId: string,
  input: CreateWidget,
): Promise<Widget> {
  const result = await client.query<UnknownRow>({
    text: "INSERT INTO widgets (tenant_id, id, name, description, created_at, updated_at, version) VALUES ($1, $2, $3, $4, $5, $5, 1) RETURNING id, name, description, created_at, updated_at, version",
    values: [
      tenantId,
      input.id,
      input.name,
      input.description,
      input.createdAt,
    ],
  });
  const row = result.rows[0];
  if (row === undefined) {
    throw new DependencyError("create widget returned no row", undefined);
  }
  return parseWidgetRow(row);
}

async function completeIdempotency(
  client: PoolClient,
  tenantId: string,
  key: string,
  widgetId: WidgetId,
): Promise<void> {
  const result = await client.query({
    text: "UPDATE idempotency_keys SET widget_id = $3 WHERE tenant_id = $1 AND operation = 'widget.create' AND idempotency_key = $2",
    values: [tenantId, key, widgetId],
  });
  if (result.rowCount !== 1) {
    throw new DependencyError("idempotency completion failed", undefined);
  }
}

function mapDatabaseError(message: string, error: unknown): unknown {
  if (error instanceof AppError || isAbortError(error)) {
    return error;
  }
  if (error instanceof DatabaseError && error.code === "23505") {
    return new DuplicateWidgetNameError();
  }
  return new DependencyError(message, error);
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

function assertLimit(limit: number): void {
  if (!Number.isInteger(limit) || limit < 1 || limit > 101) {
    throw new RangeError("repository list limit must be between 1 and 101");
  }
}
