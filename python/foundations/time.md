# Time

Time handling rules that keep behavior deterministic, timezone-correct, and separate from elapsed-time measurement.

## Default Approach

Time is an injected input. Persist and transmit aware UTC instants, use monotonic time for durations, and represent civil-time rules with explicit IANA zones.

### Inject A Clock

Core code that makes a decision from “now” consumes a narrow Protocol:

```python
class Clock(Protocol):
    def now(self) -> datetime: ...

class SystemClock:
    def now(self) -> datetime:
        return datetime.now(timezone.utc)
```

Wire `SystemClock` in composition and a controllable fake in tests. Add sleep/timer behavior to a separate seam only when the use case schedules or waits; do not turn `Clock` into a general runtime abstraction.

### UTC Instants Only

`datetime.now(timezone.utc)` is the only direct wall-clock read in production clock code. `datetime.now()` without a zone, `datetime.utcnow()`, naive constructors at boundaries, and implicit local time are forbidden. Ruff `DTZ` rules enforce common violations.

Normalize persisted, logged, and serialized instants to UTC. An aware non-UTC input may be converted with `astimezone(timezone.utc)` at the boundary. Reject naive input instead of guessing its zone. Python distinguishes aware and naive objects as documented in the [datetime reference](https://docs.python.org/3/library/datetime.html#aware-and-naive-objects).

### Civil Time Uses zoneinfo

Use `zoneinfo.ZoneInfo` and an explicit IANA zone key when the domain means a human calendar rule such as “09:00 America/New_York.” Store the zone identifier with the local schedule; do not freeze the current UTC offset because daylight-saving and political rules change. The stdlib [zoneinfo documentation](https://docs.python.org/3/library/zoneinfo.html) defines its IANA database behavior and ambiguous-time `fold` handling.

Resolve nonexistent and ambiguous local times as a named domain policy. Tests cover both DST transitions. A host's local timezone is never a business default.

### Durations Use Monotonic Time

Use `time.monotonic()` or the event loop's monotonic `loop.time()` to measure elapsed duration and compute relative deadlines. Wall time can jump when synchronized or administratively changed; it is for timestamps, not timeout arithmetic. Do not subtract independent wall-clock reads to enforce a timeout.

Represent spans as `timedelta` inside domain/config code and document units on integer wire/storage fields. Raw numeric seconds or milliseconds do not cross an unlabelled boundary.

### Scheduling And Timeout Arithmetic

- Convert an absolute external deadline to one non-negative remaining duration at the boundary.
- Pass the remaining budget down or scope work with `asyncio.timeout()`; retries share that budget.
- Compute fixed-delay schedules from monotonic time so wall-clock correction does not burst or stall them.
- Compute calendar schedules from aware civil time and an explicit `ZoneInfo`, then convert the chosen instant to UTC.
- Cap every wait and loop; a schedule always has cancellation and an owner.

### Datetime Comparison And Serialization

Compare aware instants only. Normalize to UTC when crossing systems so equality and ordering are obvious. Never compare naive and aware datetimes or strip `tzinfo` to make comparison compile.

Serialize aware ISO 8601 UTC timestamps under [serialization](serialization.md). Database columns and drivers must preserve awareness/UTC semantics; round-trip tests use the real database.

### Tests Are Sleepless

The shared fake clock lives under `tests/testutil` or the documented test-support package and advances explicitly. Time-derived IDs, expiry, retry schedules, and golden timestamps use fixed fixtures. Coordinate async work with events/queues, never `time.sleep()` or a wall-clock delay. See [testing](../quality/testing.md).

## Common Mistakes And Forbidden Patterns

- `datetime.now()` or `datetime.utcnow()` in core/business logic.
- Naive datetime accepted, persisted, compared, or serialized; `tzinfo` attached without real conversion.
- The machine's local zone used for a business decision.
- A fixed UTC offset stored where an IANA civil-time rule is intended.
- Wall-clock subtraction used for elapsed time or timeout enforcement.
- Raw numeric duration with no unit.
- DST gaps/overlaps ignored in scheduling.
- Real sleeps in tests or ambient current time in golden output.

## Verification And Proof

```bash
uv run ruff check .
uv run pytest -k "time or clock or schedule or timeout"
make verify
```

Time handling is done when core has no direct wall-clock reads; all wire/storage values are aware UTC; monotonic time owns elapsed calculations; civil schedules cover DST gap/fold policy; fake-clock tests advance without sleeping; and real persistence/serialization round trips retain the intended instant.

Related: [concurrency](concurrency-and-asyncio.md), [configuration](configuration.md), and [resilience](../operations/resilience.md).
