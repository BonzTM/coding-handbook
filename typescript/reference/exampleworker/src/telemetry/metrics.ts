export type ConsumeOutcome = "ack" | "retry" | "duplicate" | "dead_lettered" | "nack";

export interface Metrics {
  consumed(eventType: string, outcome: ConsumeOutcome): void;
  published(eventType: string): void;
  snapshot(): MetricsSnapshot;
}

export type MetricsSnapshot = Readonly<{
  consumed: Readonly<Record<string, number>>;
  published: Readonly<Record<string, number>>;
}>;

export class CounterMetrics implements Metrics {
  readonly #consumed = new Map<string, number>();
  readonly #published = new Map<string, number>();

  consumed(eventType: string, outcome: ConsumeOutcome): void {
    increment(this.#consumed, `${eventType}:${outcome}`);
  }

  published(eventType: string): void {
    increment(this.#published, eventType);
  }

  snapshot(): MetricsSnapshot {
    return Object.freeze({
      consumed: Object.freeze(Object.fromEntries(this.#consumed)),
      published: Object.freeze(Object.fromEntries(this.#published)),
    });
  }
}

export function renderMetrics(metrics: Metrics): string {
  const snapshot = metrics.snapshot();
  const lines = ["# TYPE exampleworker_messages_consumed_total counter"];
  for (const [key, value] of Object.entries(snapshot.consumed)) {
    const separator = key.lastIndexOf(":");
    const eventType = key.slice(0, separator);
    const outcome = key.slice(separator + 1);
    lines.push(
      `exampleworker_messages_consumed_total{event_type="${eventType}",outcome="${outcome}"} ${String(value)}`,
    );
  }
  lines.push("# TYPE exampleworker_messages_published_total counter");
  for (const [eventType, value] of Object.entries(snapshot.published)) {
    lines.push(
      `exampleworker_messages_published_total{event_type="${eventType}"} ${String(value)}`,
    );
  }
  return `${lines.join("\n")}\n`;
}

function increment(values: Map<string, number>, key: string): void {
  values.set(key, (values.get(key) ?? 0) + 1);
}
