import { describe, expect, it } from "@jest/globals";
import { ForbiddenError, WidgetNotFoundError } from "./errors.js";
import { WidgetService } from "./widget-service.js";
import { FakeClock } from "../testutil/fake-clock.js";
import { createFakeWidgetRepository } from "../testutil/fake-widget-repository.js";
import {
  buildPrincipal,
  SECOND_WIDGET_ID,
  WIDGET_ID,
} from "../testutil/builders.js";

const activeSignal = new AbortController().signal;

describe("WidgetService", () => {
  it("creates with one injected instant and enforces roles", async () => {
    const clock = new FakeClock(new Date("2026-08-08T12:00:00.000Z"));
    const service = new WidgetService(
      createFakeWidgetRepository(),
      clock,
      60_000,
    );

    const created = await service.create(
      buildPrincipal(),
      createCommand(WIDGET_ID, "Meter", "one"),
      { signal: activeSignal },
    );

    expect(created.widget.createdAt.toISOString()).toBe(
      "2026-08-08T12:00:00.000Z",
    );
    expect(() =>
      service.create(
        buildPrincipal(["widgets.reader"]),
        createCommand(SECOND_WIDGET_ID, "Gauge", "two"),
        { signal: activeSignal },
      ),
    ).toThrow(ForbiddenError);
  });

  it("paginates with a bounded opaque continuation", async () => {
    const clock = new FakeClock(new Date("2026-08-08T12:00:00.000Z"));
    const service = new WidgetService(
      createFakeWidgetRepository(),
      clock,
      60_000,
    );
    const principal = buildPrincipal();
    await service.create(principal, createCommand(WIDGET_ID, "Meter", "one"), {
      signal: activeSignal,
    });
    clock.advance(1);
    await service.create(
      principal,
      createCommand(SECOND_WIDGET_ID, "Gauge", "two"),
      { signal: activeSignal },
    );

    const first = await service.list(principal, undefined, 1, {
      signal: activeSignal,
    });
    const second = await service.list(principal, first.nextCursor, 1, {
      signal: activeSignal,
    });

    expect(first.items.map((widget) => widget.id)).toEqual([WIDGET_ID]);
    expect(second.items.map((widget) => widget.id)).toEqual([SECOND_WIDGET_ID]);
    expect(second.nextCursor).toBeUndefined();
  });

  it("maps an absent widget to the domain not-found error", async () => {
    const service = new WidgetService(
      createFakeWidgetRepository(),
      new FakeClock(new Date("2026-08-08T12:00:00.000Z")),
      60_000,
    );

    await expect(
      service.get(buildPrincipal(), WIDGET_ID, { signal: activeSignal }),
    ).rejects.toBeInstanceOf(WidgetNotFoundError);
  });
});

function createCommand(id: typeof WIDGET_ID, name: string, key: string) {
  return {
    id,
    name,
    description: null,
    idempotencyKey: key,
    fingerprint: `${key}-fingerprint`,
  };
}
