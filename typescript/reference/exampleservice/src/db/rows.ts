import { z } from "zod/v4";
import { parseWidgetId } from "../core/widget-id.js";
import type { Widget } from "../core/widget.js";

export const widgetRowSchema = z.strictObject({
  id: z.uuid(),
  name: z.string().min(1).max(100),
  description: z.string().max(500).nullable(),
  created_at: z.date(),
  updated_at: z.date(),
  version: z.number().int().positive(),
});

export function parseWidgetRow(input: unknown): Widget {
  const row = widgetRowSchema.parse(input);
  return {
    id: parseWidgetId(row.id),
    name: row.name,
    description: row.description,
    createdAt: new Date(row.created_at.getTime()),
    updatedAt: new Date(row.updated_at.getTime()),
    version: row.version,
  };
}
