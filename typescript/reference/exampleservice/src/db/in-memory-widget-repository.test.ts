import { describe, expect, it } from "@jest/globals";
import {
  DuplicateWidgetNameError,
  IdempotencyConflictError,
  WidgetVersionConflictError,
} from "../core/errors.js";
import {
  buildWidget,
  SECOND_WIDGET_ID,
  WIDGET_ID,
} from "../testutil/builders.js";
import { InMemoryWidgetRepository } from "./in-memory-widget-repository.js";

const options = { signal: new AbortController().signal };
const claim = {
  key: "request-one",
  fingerprint: "fingerprint-one",
  expiresAt: new Date("2026-08-09T12:00:00.000Z"),
};

describe("InMemoryWidgetRepository", () => {
  it("implements CRUD, optimistic versions, and tenant isolation", async () => {
    const repository = new InMemoryWidgetRepository();
    const widget = buildWidget();
    await repository.create("tenant-a", widget, claim, options);

    const updated = await repository.update(
      "tenant-a",
      WIDGET_ID,
      {
        name: "Updated Meter",
        description: "new",
        expectedVersion: 1,
        updatedAt: new Date("2026-08-08T12:00:01.000Z"),
      },
      options,
    );

    expect(updated.version).toBe(2);
    expect(await repository.get("tenant-b", WIDGET_ID, options)).toBeNull();
    expect(() =>
      repository.update(
        "tenant-a",
        WIDGET_ID,
        {
          name: "stale",
          description: null,
          expectedVersion: 1,
          updatedAt: new Date("2026-08-08T12:00:02.000Z"),
        },
        options,
      ),
    ).toThrow(WidgetVersionConflictError);
    await repository.delete("tenant-a", WIDGET_ID, options);
    expect(await repository.get("tenant-a", WIDGET_ID, options)).toBeNull();
  });

  it("replays one effect and rejects key reuse with another fingerprint", async () => {
    const repository = new InMemoryWidgetRepository();
    const first = await repository.create(
      "tenant-a",
      buildWidget(),
      claim,
      options,
    );
    const replay = await repository.create(
      "tenant-a",
      buildWidget(),
      claim,
      options,
    );

    expect(first.replayed).toBe(false);
    expect(replay).toEqual({ widget: first.widget, replayed: true });
    expect(() =>
      repository.create(
        "tenant-a",
        buildWidget(),
        { ...claim, fingerprint: "different" },
        options,
      ),
    ).toThrow(IdempotencyConflictError);
  });

  it("enforces unique names per tenant", async () => {
    const repository = new InMemoryWidgetRepository();
    await repository.create("tenant-a", buildWidget(), claim, options);

    expect(() =>
      repository.create(
        "tenant-a",
        buildWidget({ id: SECOND_WIDGET_ID }),
        { ...claim, key: "request-two" },
        options,
      ),
    ).toThrow(DuplicateWidgetNameError);
  });

  it("honors a pre-aborted operation", () => {
    const repository = new InMemoryWidgetRepository();
    const controller = new AbortController();
    controller.abort(new Error("test cancellation"));

    expect(() =>
      repository.get("tenant-a", WIDGET_ID, { signal: controller.signal }),
    ).toThrow("test cancellation");
  });
});
