import type { FastifyReply, FastifyRequest } from "fastify";
import { z } from "zod/v4";
import { AppError, type ErrorCode } from "../core/errors.js";
import type { Clock } from "../core/clock.js";
import type { AuditSink } from "../telemetry/audit.js";

type ProblemDefinition = Readonly<{
  type: string;
  title: string;
  status: number;
}>;

const problems = {
  unauthenticated: {
    type: "/problems/unauthenticated",
    title: "Authentication required",
    status: 401,
  },
  forbidden: {
    type: "/problems/forbidden",
    title: "Forbidden",
    status: 403,
  },
  widget_invalid: {
    type: "/problems/validation",
    title: "Invalid widget",
    status: 422,
  },
  cursor_invalid: {
    type: "/problems/validation",
    title: "Invalid cursor",
    status: 400,
  },
  widget_not_found: {
    type: "/problems/not-found",
    title: "Widget not found",
    status: 404,
  },
  widget_name_conflict: {
    type: "/problems/conflict",
    title: "Widget name conflict",
    status: 409,
  },
  widget_version_conflict: {
    type: "/problems/conflict",
    title: "Widget version conflict",
    status: 409,
  },
  idempotency_conflict: {
    type: "/problems/idempotency-key-reuse",
    title: "Idempotency key reuse",
    status: 422,
  },
  dependency_failed: {
    type: "/problems/unavailable",
    title: "Service unavailable",
    status: 503,
  },
} as const satisfies Readonly<Record<ErrorCode, ProblemDefinition>>;

const validationErrorSchema = z.object({
  validation: z.array(
    z.object({
      instancePath: z.string(),
      keyword: z.string(),
    }),
  ),
});

const fastifyErrorSchema = z.object({ code: z.string() });

export function createErrorHandler(
  audit: AuditSink,
  clock: Clock,
): (
  error: unknown,
  request: FastifyRequest,
  reply: FastifyReply,
) => Promise<void> {
  return async (error, request, reply) => {
    const validation = validationErrorSchema.safeParse(error);
    if (validation.success) {
      await reply
        .type("application/problem+json")
        .code(400)
        .send(validationProblem(request, validation.data.validation));
      return;
    }
    const transportProblem = parseTransportProblem(error);
    if (transportProblem !== undefined) {
      await reply
        .type("application/problem+json")
        .code(transportProblem.status)
        .send({
          ...transportProblem,
          instance: `urn:request:${request.id}`,
        });
      return;
    }
    if (error instanceof AppError) {
      auditAuthorizationDenial(error, request, audit, clock);
      const definition = problems[error.code];
      await reply
        .type("application/problem+json")
        .code(definition.status)
        .send({
          ...definition,
          detail: error.message,
          instance: `urn:request:${request.id}`,
          code: error.code,
        });
      return;
    }
    request.log.error(
      { err: error, event: "request_failed" },
      "request failed",
    );
    await reply
      .type("application/problem+json")
      .code(500)
      .send({
        type: "/problems/internal",
        title: "Internal server error",
        status: 500,
        detail: "an unexpected error occurred",
        instance: `urn:request:${request.id}`,
        code: "internal",
      });
  };
}

function parseTransportProblem(error: unknown) {
  const parsed = fastifyErrorSchema.safeParse(error);
  if (!parsed.success) {
    return undefined;
  }
  if (parsed.data.code === "FST_ERR_CTP_BODY_TOO_LARGE") {
    return {
      type: "/problems/payload-too-large",
      title: "Payload too large",
      status: 413,
      detail: "the request body exceeds the service limit",
      code: "payload_too_large",
    } as const;
  }
  if (parsed.data.code === "FST_ERR_CTP_INVALID_MEDIA_TYPE") {
    return {
      type: "/problems/unsupported-media-type",
      title: "Unsupported media type",
      status: 415,
      detail: "the request content type is not supported",
      code: "unsupported_media_type",
    } as const;
  }
  return undefined;
}

function validationProblem(
  request: FastifyRequest,
  issues: readonly {
    readonly instancePath: string;
    readonly keyword: string;
  }[],
) {
  return {
    type: "/problems/validation",
    title: "Request validation failed",
    status: 400,
    detail: "one or more request values are invalid",
    instance: `urn:request:${request.id}`,
    code: "validation_failed",
    errors: issues.map((issue) => ({
      field: issue.instancePath.length === 0 ? "request" : issue.instancePath,
      code: issue.keyword,
    })),
  };
}

function auditAuthorizationDenial(
  error: AppError,
  request: FastifyRequest,
  audit: AuditSink,
  clock: Clock,
): void {
  if (error.code !== "forbidden" || request.principal === undefined) {
    return;
  }
  audit.emit({
    actor: request.principal.subject,
    tenantId: request.principal.tenantId,
    action: "authorize",
    targetId: request.routeOptions.url ?? "unknown",
    result: "denied",
    occurredAt: clock.now().toISOString(),
    requestId: request.id,
  });
}
