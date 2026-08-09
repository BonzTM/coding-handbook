import { http, HttpResponse } from "msw";
import { z } from "zod/v4";
import {
  widgetSchema,
  type Widget,
} from "../features/widgets/api/widget-schemas.js";

const createBodySchema = z.strictObject({
  id: z.uuid(),
  name: z.string().min(1).max(100),
  description: z.string().max(500).nullable(),
});

const initialWidgets: readonly Widget[] = [
  {
    id: "550e8400-e29b-41d4-a716-446655440000",
    name: "Meter",
    description: "Measures throughput",
    created_at: "2026-08-09T12:00:00.000Z",
    updated_at: "2026-08-09T12:00:00.000Z",
    version: 1,
  },
  {
    id: "6ba7b810-9dad-41d1-80b4-00c04fd430c8",
    name: "Gauge",
    description: null,
    created_at: "2026-08-09T12:01:00.000Z",
    updated_at: "2026-08-09T12:01:00.000Z",
    version: 1,
  },
];

let widgets = [...initialWidgets];

export function resetMockWidgets(): void {
  widgets = [...initialWidgets];
}

export const handlers = [
  http.get("*/widgets", ({ request }) => {
    const url = new URL(request.url);
    if (url.searchParams.get("page_size") !== "20") {
      return HttpResponse.json({ title: "Invalid page size" }, { status: 400 });
    }
    return HttpResponse.json({ items: widgets, next_cursor: "" });
  }),
  http.get("*/widgets/:id", ({ params }) => {
    const widget = widgets.find((candidate) => candidate.id === params.id);
    return widget === undefined
      ? HttpResponse.json({ title: "Widget not found" }, { status: 404 })
      : HttpResponse.json(widget);
  }),
  http.post("*/widgets", async ({ request }) => {
    if (request.headers.get("idempotency-key") === null) {
      return HttpResponse.json(
        { title: "Missing idempotency key" },
        { status: 400 },
      );
    }
    const parsed = createBodySchema.safeParse(await request.json());
    if (!parsed.success) {
      return HttpResponse.json({ title: "Invalid widget" }, { status: 400 });
    }
    const widget = widgetSchema.parse({
      ...parsed.data,
      created_at: "2026-08-09T12:02:00.000Z",
      updated_at: "2026-08-09T12:02:00.000Z",
      version: 1,
    });
    widgets = [widget, ...widgets];
    return HttpResponse.json(widget, { status: 201 });
  }),
  http.delete("*/widgets/:id", ({ params }) => {
    widgets = widgets.filter((widget) => widget.id !== params.id);
    return new HttpResponse(null, { status: 204 });
  }),
];
