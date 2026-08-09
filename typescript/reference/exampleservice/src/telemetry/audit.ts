import type { Logger } from "pino";

export type AuditEvent = Readonly<{
  actor: string;
  tenantId: string;
  action:
    | "authenticate"
    | "authorize"
    | "widget.create"
    | "widget.update"
    | "widget.delete";
  targetId: string;
  result: "success" | "failure" | "denied";
  occurredAt: string;
  requestId: string;
}>;

export interface AuditSink {
  emit(event: AuditEvent): void;
}

export class PinoAuditSink implements AuditSink {
  readonly #logger: Logger;

  constructor(logger: Logger) {
    this.#logger = logger;
  }

  emit(event: AuditEvent): void {
    this.#logger.info({ event }, "security audit event");
  }
}
