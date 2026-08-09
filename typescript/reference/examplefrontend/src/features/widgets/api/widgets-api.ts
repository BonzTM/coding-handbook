import type { ApiClient } from "../../../lib/api/http-client.js";
import {
  widgetPageSchema,
  widgetSchema,
  type CreateWidgetInput,
  type Widget,
  type WidgetPage,
} from "./widget-schemas.js";

const PAGE_SIZE = 20;

export class WidgetsApi {
  readonly #client: ApiClient;

  constructor(client: ApiClient) {
    this.#client = client;
  }

  list(cursor: string, signal: AbortSignal): Promise<WidgetPage> {
    const query = new URLSearchParams({ page_size: String(PAGE_SIZE) });
    if (cursor.length > 0) {
      query.set("cursor", cursor);
    }
    return this.#client.requestJson(
      `widgets?${query.toString()}`,
      widgetPageSchema,
      { signal },
    );
  }

  get(id: string, signal: AbortSignal): Promise<Widget> {
    return this.#client.requestJson(
      `widgets/${encodeURIComponent(id)}`,
      widgetSchema,
      { signal },
    );
  }

  create(input: CreateWidgetInput): Promise<Widget> {
    return this.#client.requestJson("widgets", widgetSchema, {
      method: "POST",
      headers: { "idempotency-key": input.idempotencyKey },
      body: {
        id: input.id,
        name: input.name,
        description: input.description,
      },
    });
  }

  delete(id: string): Promise<void> {
    return this.#client.requestEmpty(`widgets/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  }
}
