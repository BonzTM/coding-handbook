import type { Widget } from "../core/widget.js";

export type WidgetResponse = Readonly<{
  id: string;
  name: string;
  description: string | null;
  created_at: string;
  updated_at: string;
  version: number;
}>;

export function encodeWidget(widget: Widget): WidgetResponse {
  return {
    id: widget.id,
    name: widget.name,
    description: widget.description,
    created_at: widget.createdAt.toISOString(),
    updated_at: widget.updatedAt.toISOString(),
    version: widget.version,
  };
}
