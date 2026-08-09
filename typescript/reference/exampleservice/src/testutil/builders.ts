import type { Principal, Role } from "../core/auth.js";
import { parseWidgetId } from "../core/widget-id.js";
import type { Widget } from "../core/widget.js";

export const WIDGET_ID = parseWidgetId("550e8400-e29b-41d4-a716-446655440000");
export const SECOND_WIDGET_ID = parseWidgetId(
  "550e8400-e29b-41d4-a716-446655440001",
);

export function buildPrincipal(
  roles: readonly Role[] = ["widgets.reader", "widgets.writer"],
): Principal {
  return {
    subject: "test-user",
    tenantId: "test-tenant",
    roles: new Set(roles),
  };
}

export function buildWidget(overrides: Partial<Widget> = {}): Widget {
  return {
    id: WIDGET_ID,
    name: "Meter",
    description: null,
    createdAt: new Date("2026-08-08T12:00:00.000Z"),
    updatedAt: new Date("2026-08-08T12:00:00.000Z"),
    version: 1,
    ...overrides,
  };
}
