import { z } from "zod/v4";

import { WIDGET_CREATED, WIDGET_DELETED, type WidgetEvent } from "../core/events.js";

const envelopeBase = z.strictObject({
  id: z.uuid(),
  version: z.literal(1),
  occurredAt: z.iso.datetime({ offset: true }),
  producer: z.string().min(1).max(100),
  traceparent: z
    .string()
    .regex(/^00-[0-9a-f]{32}-[0-9a-f]{16}-[0-9a-f]{2}$/)
    .optional(),
});

const widgetCreatedEnvelope = envelopeBase.extend({
  type: z.literal(WIDGET_CREATED),
  payload: z.strictObject({
    widgetId: z.uuid(),
    tenantId: z.string().min(1).max(100),
    name: z.string().min(1).max(100),
  }),
});

const widgetDeletedEnvelope = envelopeBase.extend({
  type: z.literal(WIDGET_DELETED),
  payload: z.strictObject({
    widgetId: z.uuid(),
    tenantId: z.string().min(1).max(100),
  }),
});

const envelopeSchema = z.discriminatedUnion("type", [widgetCreatedEnvelope, widgetDeletedEnvelope]);

export type EventEnvelope = Readonly<z.infer<typeof envelopeSchema>>;

export const MAX_MESSAGE_BYTES = 256 * 1_024;

export function parseEnvelope(body: string): EventEnvelope {
  if (Buffer.byteLength(body, "utf8") > MAX_MESSAGE_BYTES) {
    throw new RangeError(`message exceeds ${String(MAX_MESSAGE_BYTES)} bytes`);
  }
  const input: unknown = JSON.parse(body);
  return envelopeSchema.parse(input);
}

export function toWidgetEvent(envelope: EventEnvelope): WidgetEvent {
  const occurredAt = new Date(envelope.occurredAt);
  switch (envelope.type) {
    case WIDGET_CREATED:
      return Object.freeze({ type: envelope.type, ...envelope.payload, occurredAt });
    case WIDGET_DELETED:
      return Object.freeze({ type: envelope.type, ...envelope.payload, occurredAt });
  }
}
