import { createHash } from "node:crypto";
import type {
  FastifyInstance,
  FastifyPluginAsync,
  FastifyRequest,
} from "fastify";
import type { ZodTypeProvider } from "fastify-type-provider-zod";
import type { Clock } from "../core/clock.js";
import { decodeCursor, encodeCursor } from "../core/cursor.js";
import type { WidgetService } from "../core/widget-service.js";
import { parseWidgetId } from "../core/widget-id.js";
import type { AuditSink } from "../telemetry/audit.js";
import type { Authenticator } from "./auth.js";
import { withRequestSignal } from "./request-signal.js";
import {
  createRequestHeadersSchema,
  createWidgetBodySchema,
  listWidgetsQuerySchema,
  noContentResponseSchema,
  requestHeadersSchema,
  updateWidgetBodySchema,
  widgetIdParamsSchema,
  widgetPageResponseSchema,
  widgetResponseSchema,
} from "./schemas.js";
import { encodeWidget } from "./serialization.js";

export type WidgetRoutesOptions = Readonly<{
  widgets: WidgetService;
  authenticator: Authenticator;
  audit: AuditSink;
  clock: Clock;
  lifetimeSignal: AbortSignal;
}>;

export const widgetRoutes: FastifyPluginAsync<WidgetRoutesOptions> = (
  app,
  options,
) => {
  app.addHook("onRequest", (request) => authenticate(request, options));
  registerCreateRoute(app, options);
  registerGetRoute(app, options);
  registerListRoute(app, options);
  registerUpdateRoute(app, options);
  registerDeleteRoute(app, options);
  return Promise.resolve();
};

function authenticate(
  request: FastifyRequest,
  options: WidgetRoutesOptions,
): Promise<void> {
  try {
    request.principal = options.authenticator.authenticate(
      request.headers.authorization,
    );
    return Promise.resolve();
  } catch (error: unknown) {
    options.audit.emit({
      actor: "unknown",
      tenantId: "unknown",
      action: "authenticate",
      targetId: request.routeOptions.url ?? "unknown",
      result: "failure",
      occurredAt: options.clock.now().toISOString(),
      requestId: request.id,
    });
    const rejection =
      error instanceof Error
        ? error
        : new Error("authentication failed", { cause: error });
    return Promise.reject(rejection);
  }
}

function registerCreateRoute(
  app: FastifyInstance,
  options: WidgetRoutesOptions,
): void {
  app.withTypeProvider<ZodTypeProvider>().post(
    "/widgets",
    {
      schema: {
        headers: createRequestHeadersSchema,
        body: createWidgetBodySchema,
        response: { 201: widgetResponseSchema },
      },
    },
    async (request, reply) => {
      const result = await withRequestSignal(
        request,
        options.lifetimeSignal,
        (signal) =>
          options.widgets.create(
            request.principal,
            {
              id: parseWidgetId(request.body.id),
              name: request.body.name,
              description: request.body.description,
              idempotencyKey: request.headers["idempotency-key"],
              fingerprint: fingerprint(request.body),
            },
            { signal },
          ),
      );
      if (!result.replayed) {
        emitMutationAudit(options, request, "widget.create", result.widget.id);
      }
      return reply.code(201).send(encodeWidget(result.widget));
    },
  );
}

function registerGetRoute(
  app: FastifyInstance,
  options: WidgetRoutesOptions,
): void {
  app.withTypeProvider<ZodTypeProvider>().get(
    "/widgets/:id",
    {
      schema: {
        headers: requestHeadersSchema,
        params: widgetIdParamsSchema,
        response: { 200: widgetResponseSchema },
      },
    },
    async (request, reply) => {
      const widget = await withRequestSignal(
        request,
        options.lifetimeSignal,
        (signal) =>
          options.widgets.get(
            request.principal,
            parseWidgetId(request.params.id),
            { signal },
          ),
      );
      return reply.code(200).send(encodeWidget(widget));
    },
  );
}

function registerListRoute(
  app: FastifyInstance,
  options: WidgetRoutesOptions,
): void {
  app.withTypeProvider<ZodTypeProvider>().get(
    "/widgets",
    {
      schema: {
        headers: requestHeadersSchema,
        querystring: listWidgetsQuerySchema,
        response: { 200: widgetPageResponseSchema },
      },
    },
    async (request, reply) => {
      const page = await withRequestSignal(
        request,
        options.lifetimeSignal,
        (signal) =>
          options.widgets.list(
            request.principal,
            decodeCursor(request.query.cursor ?? ""),
            request.query.page_size,
            { signal },
          ),
      );
      return reply.code(200).send({
        items: page.items.map(encodeWidget),
        next_cursor: encodeCursor(page.nextCursor),
      });
    },
  );
}

function registerUpdateRoute(
  app: FastifyInstance,
  options: WidgetRoutesOptions,
): void {
  app.withTypeProvider<ZodTypeProvider>().put(
    "/widgets/:id",
    {
      schema: {
        headers: requestHeadersSchema,
        params: widgetIdParamsSchema,
        body: updateWidgetBodySchema,
        response: { 200: widgetResponseSchema },
      },
    },
    async (request, reply) => {
      const widget = await withRequestSignal(
        request,
        options.lifetimeSignal,
        (signal) =>
          options.widgets.update(
            request.principal,
            parseWidgetId(request.params.id),
            {
              name: request.body.name,
              description: request.body.description,
              expectedVersion: request.body.expected_version,
            },
            { signal },
          ),
      );
      emitMutationAudit(options, request, "widget.update", widget.id);
      return reply.code(200).send(encodeWidget(widget));
    },
  );
}

function registerDeleteRoute(
  app: FastifyInstance,
  options: WidgetRoutesOptions,
): void {
  app.withTypeProvider<ZodTypeProvider>().delete(
    "/widgets/:id",
    {
      schema: {
        headers: requestHeadersSchema,
        params: widgetIdParamsSchema,
        response: { 204: noContentResponseSchema },
      },
    },
    async (request, reply) => {
      const id = parseWidgetId(request.params.id);
      await withRequestSignal(request, options.lifetimeSignal, (signal) =>
        options.widgets.delete(request.principal, id, { signal }),
      );
      emitMutationAudit(options, request, "widget.delete", id);
      return reply.code(204).send();
    },
  );
}

function fingerprint(value: unknown): string {
  return createHash("sha256").update(JSON.stringify(value)).digest("hex");
}

function emitMutationAudit(
  options: WidgetRoutesOptions,
  request: FastifyRequest,
  action: "widget.create" | "widget.update" | "widget.delete",
  targetId: string,
): void {
  const principal = request.principal;
  if (principal === undefined) {
    return;
  }
  options.audit.emit({
    actor: principal.subject,
    tenantId: principal.tenantId,
    action,
    targetId,
    result: "success",
    occurredAt: options.clock.now().toISOString(),
    requestId: request.id,
  });
}
