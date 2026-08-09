import {
  afterAll,
  afterEach,
  beforeAll,
  describe,
  expect,
  it,
} from "@jest/globals";
import {
  PostgreSqlContainer,
  type StartedPostgreSqlContainer,
} from "@testcontainers/postgresql";
import type { Pool } from "pg";
import { IdempotencyConflictError } from "../core/errors.js";
import { buildWidget, WIDGET_ID } from "../testutil/builders.js";
import { createPool } from "../db/pool.js";
import { PostgresWidgetRepository } from "../db/postgres-widget-repository.js";

const options = { signal: new AbortController().signal };
const execFileAsync = promisify(execFile);
let container: StartedPostgreSqlContainer | undefined;
let pool: Pool | undefined;

beforeAll(async () => {
  container = await new PostgreSqlContainer("postgres:17.6").start();
  const databaseUrl = container.getConnectionUri();
  await runMigrations(databaseUrl);
  pool = createPool({ databaseUrl, maxConnections: 2, timeoutMs: 5_000 });
});

afterEach(async () => {
  await requiredPool().query("TRUNCATE idempotency_keys, widgets");
});

afterAll(async () => {
  if (pool !== undefined) {
    await pool.end();
  }
  if (container !== undefined) {
    await container.stop();
  }
});

describe("PostgresWidgetRepository", () => {
  it("applies migrations and round-trips CRUD with timestamptz", async () => {
    const repository = new PostgresWidgetRepository(requiredPool());
    const widget = buildWidget();
    const created = await repository.create(
      "tenant-a",
      widget,
      claim("one"),
      options,
    );
    const fetched = await repository.get("tenant-a", WIDGET_ID, options);
    const updated = await repository.update(
      "tenant-a",
      WIDGET_ID,
      {
        name: "Gauge",
        description: "updated",
        expectedVersion: 1,
        updatedAt: new Date("2026-08-08T12:00:01.000Z"),
      },
      options,
    );

    expect(created.widget.createdAt.toISOString()).toBe(
      "2026-08-08T12:00:00.000Z",
    );
    expect(fetched).toEqual(created.widget);
    expect(updated.version).toBe(2);
    await repository.delete("tenant-a", WIDGET_ID, options);
    expect(await repository.get("tenant-a", WIDGET_ID, options)).toBeNull();
  });

  it("atomically replays one create and rejects a mismatched fingerprint", async () => {
    const repository = new PostgresWidgetRepository(requiredPool());
    const first = await repository.create(
      "tenant-a",
      buildWidget(),
      claim("one"),
      options,
    );
    const replay = await repository.create(
      "tenant-a",
      buildWidget(),
      claim("one"),
      options,
    );

    expect(first.replayed).toBe(false);
    expect(replay).toEqual({ widget: first.widget, replayed: true });
    await expect(
      repository.create(
        "tenant-a",
        buildWidget(),
        { ...claim("one"), fingerprint: "different" },
        options,
      ),
    ).rejects.toBeInstanceOf(IdempotencyConflictError);
  });
});

function requiredPool(): Pool {
  if (pool === undefined) {
    throw new Error("PostgreSQL pool was not initialized");
  }
  return pool;
}

function claim(key: string) {
  return {
    key,
    fingerprint: `${key}-fingerprint`,
    expiresAt: new Date("2026-08-09T12:00:00.000Z"),
  };
}

async function runMigrations(databaseUrl: string): Promise<void> {
  await execFileAsync(
    process.execPath,
    [
      "node_modules/node-pg-migrate/bin/node-pg-migrate.js",
      "up",
      "--migrations-dir",
      "src/db/migrations",
      "--migration-file-language",
      "sql",
      "--check-order",
    ],
    {
      env: { DATABASE_URL: databaseUrl },
      timeout: 30_000,
      maxBuffer: 1_048_576,
    },
  );
}
import { execFile } from "node:child_process";
import { promisify } from "node:util";
