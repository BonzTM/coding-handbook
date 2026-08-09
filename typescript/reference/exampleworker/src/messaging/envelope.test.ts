import { describe, expect, it } from "@jest/globals";

import { WIDGET_CREATED } from "../core/events.js";
import { MAX_MESSAGE_BYTES, parseEnvelope, toWidgetEvent } from "./envelope.js";

const validEnvelope = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  type: WIDGET_CREATED,
  version: 1,
  occurredAt: "2026-08-09T12:00:00.000Z",
  producer: "exampleservice",
  payload: {
    widgetId: "7d9d1e2e-7239-4b4f-aa48-769a1a886c5d",
    tenantId: "tenant-1",
    name: "Meter",
  },
};

describe("event envelope", () => {
  it("parses a stable envelope and Zod-validated payload", () => {
    const envelope = parseEnvelope(JSON.stringify(validEnvelope));
    expect(envelope.id).toBe(validEnvelope.id);
    expect(toWidgetEvent(envelope)).toEqual({
      type: WIDGET_CREATED,
      ...validEnvelope.payload,
      occurredAt: new Date(validEnvelope.occurredAt),
    });
  });

  it("rejects malformed JSON, unknown versions, and invalid payloads", () => {
    expect(() => parseEnvelope("{broken")).toThrow();
    expect(() => parseEnvelope(JSON.stringify({ ...validEnvelope, version: 2 }))).toThrow();
    expect(() =>
      parseEnvelope(
        JSON.stringify({ ...validEnvelope, payload: { ...validEnvelope.payload, name: "" } }),
      ),
    ).toThrow();
    expect(() => parseEnvelope("x".repeat(MAX_MESSAGE_BYTES + 1))).toThrow(/exceeds/);
  });
});
