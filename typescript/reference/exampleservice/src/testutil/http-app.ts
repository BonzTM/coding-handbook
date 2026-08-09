import pino from "pino";
import { buildHttpApp } from "../api/app.js";
import { LocalDevAuthenticator } from "../api/auth.js";
import { WidgetService } from "../core/widget-service.js";
import { InMemoryWidgetRepository } from "../db/in-memory-widget-repository.js";
import type { AuditEvent, AuditSink } from "../telemetry/audit.js";
import { Readiness } from "../telemetry/readiness.js";
import { FakeClock } from "./fake-clock.js";

export class RecordingAuditSink implements AuditSink {
  readonly events: AuditEvent[] = [];

  emit(event: AuditEvent): void {
    this.events.push(event);
  }
}

export function buildTestHttpApp(token?: string) {
  const clock = new FakeClock(new Date("2026-08-08T12:00:00.000Z"));
  const repository = new InMemoryWidgetRepository();
  const widgets = new WidgetService(repository, clock, 60_000);
  const readiness = new Readiness();
  const audit = new RecordingAuditSink();
  readiness.markReady();
  return {
    app: buildHttpApp({
      logger: pino({ level: "silent" }),
      widgets,
      authenticator: new LocalDevAuthenticator(token),
      audit,
      clock,
      readiness,
      lifetimeSignal: new AbortController().signal,
      ready: () => Promise.resolve(true),
      readinessTimeoutMs: 100,
    }),
    audit,
    repository,
  };
}
