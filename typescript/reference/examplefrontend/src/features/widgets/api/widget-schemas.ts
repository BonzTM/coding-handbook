import { z } from "zod/v4";

export const widgetSchema = z.strictObject({
  id: z.uuid(),
  name: z.string().min(1).max(100),
  description: z.string().max(500).nullable(),
  created_at: z.iso.datetime({ offset: true }),
  updated_at: z.iso.datetime({ offset: true }),
  version: z.int().positive(),
});

export const widgetPageSchema = z.strictObject({
  items: z.array(widgetSchema).max(100),
  next_cursor: z.string(),
});

export const createWidgetFormSchema = z.object({
  name: z.string().trim().min(1, "Enter a widget name").max(100),
  description: z.string().trim().max(500, "Use 500 characters or fewer"),
});

export type Widget = z.output<typeof widgetSchema>;
export type WidgetPage = z.output<typeof widgetPageSchema>;
export type CreateWidgetForm = z.output<typeof createWidgetFormSchema>;
export type CreateWidgetInput = Readonly<{
  id: string;
  idempotencyKey: string;
  name: string;
  description: string | null;
}>;
