# Recipe: Add Metric

Use this when a product or operational question needs a metric.

## Files To Touch

- `src/telemetry/metrics.ts` or the owning instrumentation module
- the acting boundary that records the measurement
- unit and exporter smoke tests
- dashboard, alert, SLO, and runbook when the metric supports them
- data-classification notes when dimensions could be sensitive

## Steps

1. Write the question, owner, unit, instrument kind, and expected action first.
2. Choose a counter for events, histogram for distributions, or observable gauge for current state.
3. Name the metric and unit consistently with existing telemetry conventions.
4. Define a closed attribute vocabulary such as operation, route template, result, or status class.
5. Reject raw paths, IDs, URLs, SQL, cache keys, messages, tenants, and error strings as attributes.
6. Record at the boundary that owns the completed outcome.
7. Cover success, failure, cancellation, and timeout without double counting.
8. Add dashboard or alert use; remove the metric if no consumer exists.
9. Confirm exporter queue, aggregation, and shutdown remain bounded.

```bash
npm test -- --runInBand src/telemetry
npm run typecheck
npm run verify
```

## Invariants To Preserve

- Attribute cardinality has a finite, reviewable upper bound.
- Units and histogram boundaries match the measured quantity.
- Metrics contain no PII, secrets, resource IDs, or unbounded error text.
- Retries do not accidentally count one logical operation as multiple successes.
- Telemetry failure cannot block service behavior indefinitely.
- A metric has a named operational consumer.

## Proof

- A recording exporter test asserts name, kind, unit, value, and attributes.
- Cardinality review enumerates every possible attribute value or bound.
- Failure-path tests prove one final outcome per logical operation.
- OTLP smoke exports the metric and shutdown flushes within deadline.
- `npm run verify` is green.
