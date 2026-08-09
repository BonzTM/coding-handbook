import { z } from "zod/v4";

export const widgetIdParamsSchema = z.strictObject({ id: z.uuid() });

export const createWidgetBodySchema = z.strictObject({
  id: z.uuid(),
  name: z.string().min(1).max(100),
  description: z.string().max(500).nullable().default(null),
});

export const updateWidgetBodySchema = z.strictObject({
  name: z.string().min(1).max(100),
  description: z.string().max(500).nullable(),
  expected_version: z.int().positive(),
});

export const listWidgetsQuerySchema = z.strictObject({
  cursor: z.string().max(512).optional(),
  page_size: z
    .string()
    .regex(/^\d+$/)
    .transform((value) => Number(value))
    .pipe(z.int().min(1))
    .optional(),
});

export const requestHeadersSchema = z.object({
  authorization: z.string().max(1024).optional(),
});

export const createRequestHeadersSchema = requestHeadersSchema.extend({
  "idempotency-key": z.string().min(1).max(128),
});

export const widgetResponseSchema = z.strictObject({
  id: z.uuid(),
  name: z.string(),
  description: z.string().nullable(),
  created_at: z.iso.datetime({ offset: true }),
  updated_at: z.iso.datetime({ offset: true }),
  version: z.int().positive(),
});

export const widgetPageResponseSchema = z.strictObject({
  items: z.array(widgetResponseSchema).max(100),
  next_cursor: z.string(),
});

export const probeResponseSchema = z.strictObject({
  status: z.enum(["ok", "ready", "not_ready"]),
});

export const noContentResponseSchema = z.undefined();
