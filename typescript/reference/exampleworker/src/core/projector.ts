import type { Clock } from "./clock.js";
import { WIDGET_CREATED, WIDGET_DELETED, type EventProcessor, type WidgetEvent } from "./events.js";

export type ProjectedWidget = Readonly<{
  id: string;
  tenantId: string;
  name: string;
  occurredAt: Date;
  appliedAt: Date;
  deleted: boolean;
}>;

export class WidgetProjector implements EventProcessor {
  readonly #clock: Clock;
  readonly #widgets = new Map<string, ProjectedWidget>();

  constructor(clock: Clock) {
    this.#clock = clock;
  }

  process(event: WidgetEvent, signal: AbortSignal): Promise<void> {
    if (signal.aborted) {
      const error =
        signal.reason instanceof Error ? signal.reason : new Error("projection was aborted");
      return Promise.reject(error);
    }
    const key = projectionKey(event.tenantId, event.widgetId);
    const current = this.#widgets.get(key);
    const appliedAt = this.#clock.now();

    switch (event.type) {
      case WIDGET_CREATED:
        this.#widgets.set(key, createProjection(event, appliedAt));
        break;
      case WIDGET_DELETED:
        this.#widgets.set(key, deleteProjection(event, current, appliedAt));
        break;
    }
    return Promise.resolve();
  }

  get(tenantId: string, widgetId: string): ProjectedWidget | undefined {
    return this.#widgets.get(projectionKey(tenantId, widgetId));
  }
}

function createProjection(
  event: Extract<WidgetEvent, { type: typeof WIDGET_CREATED }>,
  appliedAt: Date,
): ProjectedWidget {
  return freezeDates({
    id: event.widgetId,
    tenantId: event.tenantId,
    name: event.name,
    occurredAt: event.occurredAt,
    appliedAt,
    deleted: false,
  });
}

function deleteProjection(
  event: Extract<WidgetEvent, { type: typeof WIDGET_DELETED }>,
  current: ProjectedWidget | undefined,
  appliedAt: Date,
): ProjectedWidget {
  return freezeDates({
    id: event.widgetId,
    tenantId: event.tenantId,
    name: current?.name ?? "",
    occurredAt: event.occurredAt,
    appliedAt,
    deleted: true,
  });
}

function projectionKey(tenantId: string, widgetId: string): string {
  return `${tenantId}:${widgetId}`;
}

function freezeDates(widget: ProjectedWidget): ProjectedWidget {
  return Object.freeze({
    ...widget,
    occurredAt: new Date(widget.occurredAt.getTime()),
    appliedAt: new Date(widget.appliedAt.getTime()),
  });
}
