import type { Clock } from "./clock.js";
import type { Principal } from "./auth.js";
import { requireRole } from "./auth.js";
import type { WidgetCursor } from "./cursor.js";
import { InvalidWidgetError, WidgetNotFoundError } from "./errors.js";
import type {
  CreateResult,
  OperationOptions,
  WidgetRepository,
} from "./widget-repository.js";
import type { Widget } from "./widget.js";
import type { WidgetId } from "./widget-id.js";

export const DEFAULT_PAGE_SIZE = 20;
export const MAX_PAGE_SIZE = 100;

export type CreateWidgetCommand = Readonly<{
  id: WidgetId;
  name: string;
  description: string | null;
  idempotencyKey: string;
  fingerprint: string;
}>;

export type UpdateWidgetCommand = Readonly<{
  name: string;
  description: string | null;
  expectedVersion: number;
}>;

export type WidgetPage = Readonly<{
  items: readonly Widget[];
  nextCursor: WidgetCursor | undefined;
}>;

export class WidgetService {
  readonly #repository: WidgetRepository;
  readonly #clock: Clock;
  readonly #idempotencyTtlMs: number;

  constructor(
    repository: WidgetRepository,
    clock: Clock,
    idempotencyTtlMs: number,
  ) {
    if (!Number.isInteger(idempotencyTtlMs) || idempotencyTtlMs < 1) {
      throw new RangeError("idempotency TTL must be a positive integer");
    }
    this.#repository = repository;
    this.#clock = clock;
    this.#idempotencyTtlMs = idempotencyTtlMs;
  }

  create(
    principal: Principal | undefined,
    command: CreateWidgetCommand,
    options: OperationOptions,
  ): Promise<CreateResult> {
    const actor = requireRole(principal, "widgets.writer");
    validateName(command.name);
    options.signal.throwIfAborted();
    const now = this.#clock.now();
    return this.#repository.create(
      actor.tenantId,
      { ...command, createdAt: now },
      {
        key: command.idempotencyKey,
        fingerprint: command.fingerprint,
        expiresAt: new Date(now.getTime() + this.#idempotencyTtlMs),
      },
      options,
    );
  }

  async get(
    principal: Principal | undefined,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<Widget> {
    const actor = requireRole(principal, "widgets.reader");
    const widget = await this.#repository.get(actor.tenantId, id, options);
    if (widget === null) {
      throw new WidgetNotFoundError();
    }
    return widget;
  }

  async list(
    principal: Principal | undefined,
    cursor: WidgetCursor | undefined,
    requestedPageSize: number | undefined,
    options: OperationOptions,
  ): Promise<WidgetPage> {
    const actor = requireRole(principal, "widgets.reader");
    const pageSize = clampPageSize(requestedPageSize);
    const rows = await this.#repository.list(
      actor.tenantId,
      cursor,
      pageSize + 1,
      options,
    );
    const items = rows.slice(0, pageSize);
    const last = items.at(-1);
    const nextCursor =
      rows.length > pageSize && last !== undefined
        ? { createdAt: last.createdAt, id: last.id }
        : undefined;
    return { items, nextCursor };
  }

  update(
    principal: Principal | undefined,
    id: WidgetId,
    command: UpdateWidgetCommand,
    options: OperationOptions,
  ): Promise<Widget> {
    const actor = requireRole(principal, "widgets.writer");
    validateName(command.name);
    return this.#repository.update(
      actor.tenantId,
      id,
      { ...command, updatedAt: this.#clock.now() },
      options,
    );
  }

  delete(
    principal: Principal | undefined,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<void> {
    const actor = requireRole(principal, "widgets.writer");
    return this.#repository.delete(actor.tenantId, id, options);
  }
}

function validateName(name: string): void {
  if (name.length < 1 || name.length > 100) {
    throw new InvalidWidgetError(
      "widget name must contain 1 to 100 characters",
    );
  }
}

function clampPageSize(requested: number | undefined): number {
  if (requested === undefined || requested < 1) {
    return DEFAULT_PAGE_SIZE;
  }
  return Math.min(requested, MAX_PAGE_SIZE);
}
