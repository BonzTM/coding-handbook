import type { WidgetRepository } from "../core/widget-repository.js";
import { InMemoryWidgetRepository } from "../db/in-memory-widget-repository.js";

export function createFakeWidgetRepository(): WidgetRepository {
  return new InMemoryWidgetRepository();
}
