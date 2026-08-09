import type { WidgetCursor } from "./cursor.js";
import type { CreateWidget, UpdateWidget, Widget } from "./widget.js";
import type { WidgetId } from "./widget-id.js";

export type OperationOptions = Readonly<{ signal: AbortSignal }>;

export type IdempotencyClaim = Readonly<{
  key: string;
  fingerprint: string;
  expiresAt: Date;
}>;

export type CreateResult = Readonly<{
  widget: Widget;
  replayed: boolean;
}>;

export interface WidgetRepository {
  create(
    tenantId: string,
    input: CreateWidget,
    claim: IdempotencyClaim,
    options: OperationOptions,
  ): Promise<CreateResult>;
  get(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<Widget | null>;
  list(
    tenantId: string,
    cursor: WidgetCursor | undefined,
    limit: number,
    options: OperationOptions,
  ): Promise<readonly Widget[]>;
  update(
    tenantId: string,
    id: WidgetId,
    input: UpdateWidget,
    options: OperationOptions,
  ): Promise<Widget>;
  delete(
    tenantId: string,
    id: WidgetId,
    options: OperationOptions,
  ): Promise<void>;
  ready(options: OperationOptions): Promise<boolean>;
}
