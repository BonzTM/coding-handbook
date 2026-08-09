import type { WidgetCursor } from "../core/cursor.js";
import {
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
import {
  copyWidget,
  type CreateWidget,
  type UpdateWidget,
  type Widget,
} from "../core/widget.js";
import type { WidgetId } from "../core/widget-id.js";

type IdempotencyRecord = Readonly<{
  fingerprint: string;
  widgetKey: string;
  expiresAt: Date;
}>;

export class InMemoryWidgetRepository implements WidgetRepository {
  readonly #widgets = new Map<string, Widget>();
  readonly #idempotency = new Map<string, IdempotencyRecord>();

  create(
    tenantId: string,
    input: CreateWidget,
    claim: IdempotencyClaim,
    options: OperationOptions,
  ): Promise<CreateResult> {
    options.signal.throwIfAborted();
    const idempotencyKey = compositeKey(tenantId, claim.key);
    const prior = this.#idempotency.get(idempotencyKey);
    if (
      prior !== undefined &&
      prior.expiresAt.getTime() > input.createdAt.getTime()
    ) {
      return Promise.resolve(this.#replay(prior, claim.fingerprint));
    }
    this.#idempotency.delete(idempotencyKey);
    this.#assertNameAvailable(tenantId, input.name);
    const widgetKey = compositeKey(tenantId, input.id);
    if (this.#widgets.has(widgetKey)) {
      throw new DuplicateWidgetNameError();
    }
    const widget: Widget = {
      ...input,
      updatedAt: new Date(input.createdAt.getTime()),
      version: 1,
    };
    this.#widgets.set(widgetKey, copyWidget(widget));
    this.#idempotency.set(idempotencyKey, {
      fingerprint: claim.fingerprint,
      widgetKey,
      expiresAt: new Date(claim.expiresAt.getTime()),
    });
    return Promise.resolve({ widget: copyWidget(widget), replayed: false });
  }

  get(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<Widget | null> {
    options.signal.throwIfAborted();
    const widget = this.#widgets.get(compositeKey(tenantId, id));
    return Promise.resolve(widget === undefined ? null : copyWidget(widget));
  }

  list(
    tenantId: string,
    cursor: WidgetCursor | undefined,
    limit: number,
    options: OperationOptions,
  ): Promise<readonly Widget[]> {
    options.signal.throwIfAborted();
    assertLimit(limit);
    const prefix = `${tenantId}\u0000`;
    const widgets = [...this.#widgets.entries()]
      .filter(([key]) => key.startsWith(prefix))
      .map(([, widget]) => widget)
      .filter((widget) => isAfter(widget, cursor))
      .sort(compareWidgets)
      .slice(0, limit)
      .map(copyWidget);
    return Promise.resolve(widgets);
  }

  update(
    tenantId: string,
    id: WidgetId,
    input: UpdateWidget,
    options: OperationOptions,
  ): Promise<Widget> {
    options.signal.throwIfAborted();
    const key = compositeKey(tenantId, id);
    const widget = this.#widgets.get(key);
    if (widget === undefined) {
      throw new WidgetNotFoundError();
    }
    if (widget.version !== input.expectedVersion) {
      throw new WidgetVersionConflictError();
    }
    this.#assertNameAvailable(tenantId, input.name, key);
    const updated: Widget = {
      ...widget,
      name: input.name,
      description: input.description,
      updatedAt: new Date(input.updatedAt.getTime()),
      version: widget.version + 1,
    };
    this.#widgets.set(key, copyWidget(updated));
    return Promise.resolve(copyWidget(updated));
  }

  delete(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<void> {
    options.signal.throwIfAborted();
    if (!this.#widgets.delete(compositeKey(tenantId, id))) {
      throw new WidgetNotFoundError();
    }
    return Promise.resolve();
  }

  ready(options: OperationOptions): Promise<boolean> {
    options.signal.throwIfAborted();
    return Promise.resolve(true);
  }

  #replay(record: IdempotencyRecord, fingerprint: string): CreateResult {
    if (record.fingerprint !== fingerprint) {
      throw new IdempotencyConflictError();
    }
    const widget = this.#widgets.get(record.widgetKey);
    if (widget === undefined) {
      throw new WidgetNotFoundError();
    }
    return { widget: copyWidget(widget), replayed: true };
  }

  #assertNameAvailable(
    tenantId: string,
    name: string,
    ignoredKey?: string,
  ): void {
    const prefix = `${tenantId}\u0000`;
    const duplicate = [...this.#widgets.entries()].find(
      ([key, widget]) =>
        key !== ignoredKey && key.startsWith(prefix) && widget.name === name,
    );
    if (duplicate !== undefined) {
      throw new DuplicateWidgetNameError();
    }
  }
}

function compositeKey(first: string, second: string): string {
  return `${first}\u0000${second}`;
}

function compareWidgets(left: Widget, right: Widget): number {
  const timeOrder = left.createdAt.getTime() - right.createdAt.getTime();
  return timeOrder === 0 ? left.id.localeCompare(right.id) : timeOrder;
}

function isAfter(widget: Widget, cursor: WidgetCursor | undefined): boolean {
  if (cursor === undefined) {
    return true;
  }
  const timeOrder = widget.createdAt.getTime() - cursor.createdAt.getTime();
  return timeOrder > 0 || (timeOrder === 0 && widget.id > cursor.id);
}

function assertLimit(limit: number): void {
  if (!Number.isInteger(limit) || limit < 1 || limit > 101) {
    throw new RangeError("repository list limit must be between 1 and 101");
  }
}
