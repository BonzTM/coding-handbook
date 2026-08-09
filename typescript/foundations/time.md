# Time

Time, timer, timezone, and scheduling rules for deterministic TypeScript systems.

## Default Approach

Use UTC instants for storage and transport, inject a clock for decisions, and keep human calendar rules explicit.

### Clock Seam

Domain logic receives a narrow clock such as `now(): Date`; production composition supplies `new Date()`, tests supply a fixed or advancing fake. Do not read `Date.now()` throughout business code.

Create timestamps at the boundary that owns the event. Pass the value through one operation so related records do not disagree by milliseconds.

Use one small seam and return a fresh value so callers cannot mutate shared clock state:

```ts
export interface Clock {
  now(): Date;
}

export class SystemClock implements Clock {
  now(): Date {
    return new Date();
  }
}

export class FixedClock implements Clock {
  readonly #instant: Date;

  constructor(instant: Date) {
    this.#instant = new Date(instant.getTime());
  }

  now(): Date {
    return new Date(this.#instant.getTime());
  }
}
```

Inject the clock into the use case that owns the time decision:

```ts
type Token = Readonly<{ expiresAt: Date }>;

export class TokenPolicy {
  readonly #clock: Clock;

  constructor(clock: Clock) {
    this.#clock = clock;
  }

  isExpired(token: Token): boolean {
    return token.expiresAt.getTime() <= this.#clock.now().getTime();
  }
}
```

Composition creates `new TokenPolicy(new SystemClock())`; tests use `FixedClock`. The equality rule is visible and testable.

### Instants And Calendar Values

An instant is a point on the timeline. A local date, local time, timezone, and recurrence rule are different domain values. Do not store a user's appointment as an instant until the required zone and daylight-saving policy are known.

Persist instants in PostgreSQL `timestamptz` and serialize as RFC 3339 UTC strings. Preserve original zone identifiers separately when future calendar behavior depends on them. Never depend on host-local timezone.

### Parsing And Formatting

Parse with an explicit schema and reject ambiguous locale strings. Accept only documented formats. Formatting for humans belongs at the UI edge with an explicit locale and zone.

JavaScript `Date` is mutable and has legacy parsing traps. Copy retained dates, compare numeric epoch values, and never rely on implementation-defined parsing. Temporal may replace selected handling only when every supported runtime or an accepted polyfill strategy supports it.

### Durations And Deadlines

Represent configured durations in named units and validate bounds. Suffix numeric configuration and fields with the unit, such as `timeoutMs`. Do not mix epoch milliseconds, elapsed milliseconds, and seconds in the same primitive without naming.

Use monotonic elapsed-time facilities for measuring duration where the platform exposes them; wall-clock adjustments must not extend operational deadlines unexpectedly. Cancellation uses `AbortSignal` and a bounded timeout.

### Timers And Schedules

Every timer has an owner and cleanup path. Avoid long `setTimeout` values as durable scheduling; persist scheduled intent and use an owned scheduler or queue. Recurring work must prevent overlap or define overlap semantics explicitly.

Clock changes, missed runs, duplicate execution, restart, and daylight-saving transitions are part of scheduler behavior. Scheduled work must be idempotent when redelivery or overlap is possible.

### Tests

Use injected clocks for domain behavior and Jest fake timers for timer orchestration. Restore real timers after each test. Advance time deliberately and await scheduled promise work; do not use sleeps.

Test boundary instants, leap days, daylight-saving gaps and folds when local schedules matter, expiration equality, and clock movement where relevant.

Use modern Jest fake timers only when timer orchestration itself is the subject:

```ts
import { afterEach, describe, expect, it, jest } from "@jest/globals";

function scheduleOnce(delayMs: number, action: () => void): () => void {
  const handle = setTimeout(action, delayMs);
  return () => clearTimeout(handle);
}

afterEach(() => {
  jest.useRealTimers();
});

describe("scheduleOnce", () => {
  it("runs at the configured deadline", async () => {
    jest.useFakeTimers({
      now: new Date("2026-08-08T12:00:00Z"),
      timerLimit: 100,
    });
    const action = jest.fn();
    const cancel = scheduleOnce(1_000, action);

    await jest.advanceTimersByTimeAsync(999);
    expect(action).not.toHaveBeenCalled();
    await jest.advanceTimersByTimeAsync(1);
    expect(action).toHaveBeenCalledTimes(1);

    cancel();
  });
});
```

Do not fake timers when an injected `Clock` proves the behavior without touching global scheduler state.

## Common Mistakes And Forbidden Patterns

- Local time or timezone-free strings in storage and APIs.
- `new Date(string)` on an undocumented input format.
- Real sleeps in tests or polling loops.
- Timers created without cancellation or cleanup.
- Durable jobs represented only by an in-memory timer.
- Mixed units in unnamed numeric fields.
- Mocking the global clock when a narrow injected seam would prove the decision.

## Verification And Proof

- Domain tests use a fixed clock and contain no wall-clock dependency.
- Serialization tests pin RFC 3339 UTC representation and reject ambiguous input.
- Database integration tests prove `timestamptz` round trips as the same instant.
- Timer tests cover cancellation, cleanup, overlap, and missed-run policy.
- Local-calendar tests name locale, IANA zone, and daylight-saving expectations.
- Search finds no unexplained `Date.now()`, `new Date()`, or real test sleeps in core logic.

Related: [async-and-cancellation.md](async-and-cancellation.md), [serialization.md](serialization.md), and [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md). Test anchor: [Jest timer mocks](https://jestjs.io/docs/timer-mocks).
