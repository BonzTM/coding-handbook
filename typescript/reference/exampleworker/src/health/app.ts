import Fastify, { type FastifyInstance } from "fastify";

import type { Broker } from "../messaging/broker.js";
import type { Metrics } from "../telemetry/metrics.js";
import { renderMetrics } from "../telemetry/metrics.js";
import type { Readiness } from "../telemetry/readiness.js";

export type HealthAppOptions = Readonly<{
  readiness: Readiness;
  broker: Pick<Broker, "isHealthy">;
  metrics: Metrics;
}>;

export function buildHealthApp(options: HealthAppOptions): FastifyInstance {
  const app = Fastify({ logger: false, bodyLimit: 1_024 });

  app.get("/livez", async (_request, reply) => reply.type("text/plain").send("ok"));
  app.get("/readyz", async (_request, reply) => {
    if (!options.readiness.isReady() || !options.broker.isHealthy()) {
      return reply.code(503).type("text/plain").send("not ready");
    }
    return reply.type("text/plain").send("ready");
  });
  app.get("/metrics", async (_request, reply) =>
    reply.type("text/plain; version=0.0.4").send(renderMetrics(options.metrics)),
  );

  return app;
}
