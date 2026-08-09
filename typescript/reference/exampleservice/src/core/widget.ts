import type { WidgetId } from "./widget-id.js";

export type Widget = Readonly<{
  id: WidgetId;
  name: string;
  description: string | null;
  createdAt: Date;
  updatedAt: Date;
  version: number;
}>;

export type CreateWidget = Readonly<{
  id: WidgetId;
  name: string;
  description: string | null;
  createdAt: Date;
}>;

export type UpdateWidget = Readonly<{
  name: string;
  description: string | null;
  expectedVersion: number;
  updatedAt: Date;
}>;

export function copyWidget(widget: Widget): Widget {
  return {
    ...widget,
    createdAt: new Date(widget.createdAt.getTime()),
    updatedAt: new Date(widget.updatedAt.getTime()),
  };
}
