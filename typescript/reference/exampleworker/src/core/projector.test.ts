import { describe, expect, it } from "@jest/globals";

import { FakeClock } from "../testutil/fake-clock.js";
import { WIDGET_CREATED, WIDGET_DELETED } from "./events.js";
import { WidgetProjector } from "./projector.js";

const widgetId = "7d9d1e2e-7239-4b4f-aa48-769a1a886c5d";
const occurredAt = new Date("2026-08-09T11:00:00.000Z");

describe("WidgetProjector", () => {
  it("applies create and delete events using the injected clock", async () => {
    const clock = new FakeClock(new Date("2026-08-09T12:00:00.000Z"));
    const projector = new WidgetProjector(clock);

    await projector.process(
      { type: WIDGET_CREATED, widgetId, tenantId: "tenant-1", name: "Meter", occurredAt },
      new AbortController().signal,
    );
    clock.advance(1_000);
    await projector.process(
      { type: WIDGET_DELETED, widgetId, tenantId: "tenant-1", occurredAt },
      new AbortController().signal,
    );

    expect(projector.get("tenant-1", widgetId)).toEqual({
      id: widgetId,
      tenantId: "tenant-1",
      name: "Meter",
      occurredAt,
      appliedAt: new Date("2026-08-09T12:00:01.000Z"),
      deleted: true,
    });
  });

  it("does not mutate state when pre-aborted", async () => {
    const projector = new WidgetProjector(new FakeClock(occurredAt));
    const lifetime = new AbortController();
    lifetime.abort();

    await expect(
      projector.process(
        { type: WIDGET_CREATED, widgetId, tenantId: "tenant-1", name: "Meter", occurredAt },
        lifetime.signal,
      ),
    ).rejects.toThrow();
    expect(projector.get("tenant-1", widgetId)).toBeUndefined();
  });
});
