import { afterEach, describe, expect, it } from "@jest/globals";
import { buildTestHttpApp } from "../testutil/http-app.js";
import { SECOND_WIDGET_ID, WIDGET_ID } from "../testutil/builders.js";

type TestApp = ReturnType<typeof buildTestHttpApp>["app"];

const openApps: TestApp[] = [];

afterEach(async () => {
  await Promise.all(openApps.splice(0).map((app) => app.close()));
});

describe("widget HTTP API", () => {
  it("creates, reads, updates, lists, and deletes a widget", async () => {
    const app = testApp();
    const created = await createWidget(app, "request-one");
    expect(created.statusCode).toBe(201);
    expect(created.headers["x-request-id"]).toBeDefined();

    const fetched = await app.inject({
      method: "GET",
      url: `/widgets/${WIDGET_ID}`,
    });
    expect(fetched.statusCode).toBe(200);
    expect(fetched.json()).toMatchObject({ id: WIDGET_ID, version: 1 });

    const updated = await app.inject({
      method: "PUT",
      url: `/widgets/${WIDGET_ID}`,
      payload: { name: "Gauge", description: null, expected_version: 1 },
    });
    expect(updated.json()).toMatchObject({ name: "Gauge", version: 2 });

    const listed = await app.inject({
      method: "GET",
      url: "/widgets?page_size=1",
    });
    expect(listed.json()).toMatchObject({ next_cursor: "" });

    const deleted = await app.inject({
      method: "DELETE",
      url: `/widgets/${WIDGET_ID}`,
    });
    expect(deleted.statusCode).toBe(204);
  });

  it("rejects authentication failures without leaking the token", async () => {
    const context = buildTestHttpApp("sixteen-byte-token");
    openApps.push(context.app);

    const response = await context.app.inject({
      method: "GET",
      url: "/widgets",
    });

    expect(response.statusCode).toBe(401);
    expect(response.body).not.toContain("sixteen-byte-token");
    expect(context.audit.events).toHaveLength(1);
    expect(context.audit.events[0]).toMatchObject({
      action: "authenticate",
      result: "failure",
    });
  });

  it("returns RFC 9457 validation details for malformed input", async () => {
    const app = testApp();
    const response = await app.inject({
      method: "POST",
      url: "/widgets",
      headers: { "idempotency-key": "request-one" },
      payload: { name: "Meter", unexpected: true },
    });
    const problem: unknown = response.json();

    expect(response.statusCode).toBe(400);
    expect(response.headers["content-type"]).toContain(
      "application/problem+json",
    );
    expect(problem).toMatchObject({
      type: "/problems/validation",
      title: "Request validation failed",
      status: 400,
      code: "validation_failed",
      instance: expect.stringMatching(/^urn:request:/),
    });
  });

  it("replays an idempotent create byte-for-byte with one effect", async () => {
    const app = testApp();
    const first = await createWidget(app, "request-one");
    const replay = await createWidget(app, "request-one");
    const list = await app.inject({ method: "GET", url: "/widgets" });

    expect(replay.statusCode).toBe(201);
    expect(replay.body).toBe(first.body);
    expect(list.json()).toMatchObject({ items: [first.json()] });
  });

  it("rejects an oversized body with safe problem details", async () => {
    const app = testApp();
    const response = await app.inject({
      method: "POST",
      url: "/widgets",
      headers: {
        "content-type": "application/json",
        "idempotency-key": "request-large",
      },
      payload: JSON.stringify({
        id: WIDGET_ID,
        name: "Meter",
        description: "x".repeat(70_000),
      }),
    });

    expect(response.statusCode).toBe(413);
    expect(response.json()).toMatchObject({
      type: "/problems/payload-too-large",
      status: 413,
      code: "payload_too_large",
    });
  });

  it("returns a safe problem-details shape for not found", async () => {
    const app = testApp();
    const response = await app.inject({
      method: "GET",
      url: `/widgets/${SECOND_WIDGET_ID}`,
    });

    expect(response.statusCode).toBe(404);
    expect(response.json()).toEqual({
      type: "/problems/not-found",
      title: "Widget not found",
      status: 404,
      detail: "the widget was not found",
      instance: expect.stringMatching(/^urn:request:/),
      code: "widget_not_found",
    });
  });

  it("separates liveness from readiness", async () => {
    const app = testApp();

    expect(
      (await app.inject({ method: "GET", url: "/livez" })).statusCode,
    ).toBe(200);
    expect(
      (await app.inject({ method: "GET", url: "/readyz" })).statusCode,
    ).toBe(200);
  });
});

function testApp(): TestApp {
  const context = buildTestHttpApp();
  openApps.push(context.app);
  return context.app;
}

function createWidget(app: TestApp, key: string) {
  return app.inject({
    method: "POST",
    url: "/widgets",
    headers: { "idempotency-key": key },
    payload: { id: WIDGET_ID, name: "Meter", description: null },
  });
}
