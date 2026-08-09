export const WIDGET_CREATED = "widget.created";
export const WIDGET_DELETED = "widget.deleted";

export type WidgetCreated = Readonly<{
  type: typeof WIDGET_CREATED;
  widgetId: string;
  tenantId: string;
  name: string;
  occurredAt: Date;
}>;

export type WidgetDeleted = Readonly<{
  type: typeof WIDGET_DELETED;
  widgetId: string;
  tenantId: string;
  occurredAt: Date;
}>;

export type WidgetEvent = WidgetCreated | WidgetDeleted;

export interface EventProcessor {
  process(event: WidgetEvent, signal: AbortSignal): Promise<void>;
}
