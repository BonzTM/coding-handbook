import type { FastifyPluginAsync } from "fastify";
import type { ZodTypeProvider } from "fastify-type-provider-zod";
import type { Readiness } from "../telemetry/readiness.js";
import { probeResponseSchema } from "./schemas.js";

export type HealthRoutesOptions = Readonly<{
  readiness: Readiness;
  ready: (signal: AbortSignal) => Promise<boolean>;
  timeoutMs: number;
}>;

export const healthRoutes: FastifyPluginAsync<HealthRoutesOptions> = (
  app,
  options,
) => {
  const typed = app.withTypeProvider<ZodTypeProvider>();
  typed.get(
    "/livez",
    { schema: { response: { 200: probeResponseSchema } } },
    async (_request, reply) => reply.code(200).send({ status: "ok" }),
  );
  typed.get(
    "/readyz",
    {
      schema: {
        response: {
          200: probeResponseSchema,
          503: probeResponseSchema,
        },
      },
    },
    async (_request, reply) => {
      const signal = AbortSignal.timeout(options.timeoutMs);
      const ready =
        options.readiness.isReady() && (await options.ready(signal));
      return ready
        ? reply.code(200).send({ status: "ready" })
        : reply.code(503).send({ status: "not_ready" });
    },
  );
  return Promise.resolve();
};
