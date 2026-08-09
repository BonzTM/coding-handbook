import { afterEach, describe, expect, it } from "@jest/globals";

import { CounterMetrics } from "../telemetry/metrics.js";
import { Readiness } from "../telemetry/readiness.js";
import { buildHealthApp } from "./app.js";

const apps: ReturnType<typeof buildHealthApp>[] = [];

afterEach(async () => {
  await Promise.all(apps.splice(0).map(async (app) => app.close()));
});

describe("health sidecar", () => {
  it("keeps liveness independent and gates readiness on drain and broker health", async () => {
    const readiness = new Readiness();
    let healthy = true;
    const broker = { isHealthy: () => healthy };
    const app = buildHealthApp({
      readiness,
      broker,
      metrics: new CounterMetrics(),
    });
    apps.push(app);

    expect((await app.inject({ method: "GET", url: "/livez" })).statusCode).toBe(200);
    expect((await app.inject({ method: "GET", url: "/readyz" })).statusCode).toBe(503);
    readiness.set(true);
    expect((await app.inject({ method: "GET", url: "/readyz" })).statusCode).toBe(200);
    healthy = false;
    expect((await app.inject({ method: "GET", url: "/readyz" })).statusCode).toBe(503);
    healthy = true;
    readiness.set(false);
    expect((await app.inject({ method: "GET", url: "/livez" })).statusCode).toBe(200);
  });

  it("exposes low-cardinality counters", async () => {
    const metrics = new CounterMetrics();
    metrics.consumed("widget.created", "ack");
    const app = buildHealthApp({
      readiness: new Readiness(),
      broker: { isHealthy: () => true },
      metrics,
    });
    apps.push(app);

    const response = await app.inject({ method: "GET", url: "/metrics" });
    expect(response.statusCode).toBe(200);
    expect(response.body).toContain(
      'exampleworker_messages_consumed_total{event_type="widget.created",outcome="ack"} 1',
    );
  });
});
