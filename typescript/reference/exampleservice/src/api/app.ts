import { randomUUID } from "node:crypto";
import Fastify, { LogController } from "fastify";
import {
  serializerCompiler,
  validatorCompiler,
} from "fastify-type-provider-zod";
import type { Logger } from "pino";
import type { Clock } from "../core/clock.js";
import type { WidgetService } from "../core/widget-service.js";
import type { AuditSink } from "../telemetry/audit.js";
import type { Readiness } from "../telemetry/readiness.js";
import type { Authenticator } from "./auth.js";
import { createErrorHandler } from "./errors.js";
import { healthRoutes } from "./health-routes.js";
import { widgetRoutes } from "./widgets-routes.js";

const REQUEST_ID_PATTERN = /^[A-Za-z0-9._-]{1,64}$/;

export type HttpAppOptions = Readonly<{
  logger: Logger;
  widgets: WidgetService;
  authenticator: Authenticator;
  audit: AuditSink;
  clock: Clock;
  readiness: Readiness;
  lifetimeSignal: AbortSignal;
  ready: (signal: AbortSignal) => Promise<boolean>;
  readinessTimeoutMs: number;
}>;

export function buildHttpApp(options: HttpAppOptions) {
  const app = Fastify({
    loggerInstance: options.logger,
    logController: new LogController({ disableRequestLogging: true }),
    bodyLimit: 65_536,
    requestIdHeader: false,
    genReqId: (request) => validatedRequestId(request.headers["x-request-id"]),
  });
  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);
  app.setErrorHandler(createErrorHandler(options.audit, options.clock));
  app.addHook("onRequest", (request, reply) => {
    reply.header("x-request-id", request.id);
    return Promise.resolve();
  });
  app.addHook("onResponse", (request, reply) => {
    request.log.info(
      {
        event: "request_completed",
        method: request.method,
        route: request.routeOptions.url ?? "unmatched",
        statusCode: reply.statusCode,
        requestId: request.id,
      },
      "request completed",
    );
    return Promise.resolve();
  });
  app.register(healthRoutes, {
    readiness: options.readiness,
    ready: options.ready,
    timeoutMs: options.readinessTimeoutMs,
  });
  app.register(widgetRoutes, {
    widgets: options.widgets,
    authenticator: options.authenticator,
    audit: options.audit,
    clock: options.clock,
    lifetimeSignal: options.lifetimeSignal,
  });
  return app;
}

function validatedRequestId(value: string | string[] | undefined): string {
  if (typeof value === "string" && REQUEST_ID_PATTERN.test(value)) {
    return value;
  }
  return randomUUID();
}
