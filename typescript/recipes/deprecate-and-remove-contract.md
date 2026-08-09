# Recipe: Deprecate And Remove A Contract

Use this for an HTTP, event, library, configuration, or persisted contract.

## Files To Touch

- controlling schema, package exports, or migration
- producer and consumer adapters and contract fixtures
- usage telemetry, changelog, migration guide, and release notes
- deprecation owner/removal issue
- ADR when compatibility breaks or an invariant changes

## Steps

1. Classify the contract and all independently deployed producers and consumers.
2. Record owner, announcement date, earliest removal, telemetry, and completion criteria.
3. Introduce an additive replacement or parallel version; do not mutate meaning in place.
4. Support old producer/new consumer and new producer/old consumer during rollout.
5. Publish a migration guide and mark library exports or HTTP fields through their supported mechanism.
6. Measure remaining use without IDs or other high-cardinality metric attributes.
7. Migrate owned consumers and notify external owners through the approved channel.
8. Wait until compatibility window and measured removal criteria are satisfied.
9. Remove old code, schemas, tests, flags, telemetry, docs, and persisted shape in the safe order.

```bash
npm test -- --runInBand <contract-suite>
npm pack --dry-run
npm run verify
```

## Invariants To Preserve

- Compile-time shared types do not substitute for wire compatibility proof.
- Required fields, narrowed values, changed nullability, and status semantics are treated as breaking.
- Old and new versions coexist through rollout and rollback.
- Contract migration does not bypass authorization, validation, or retention policy.
- Removal is evidence-driven, not calendar-only.
- Accepted ADR history is superseded, never rewritten.

## Proof

- Golden fixtures cover every supported version and additive unknown fields.
- Compatibility tests cover both mixed-version directions.
- Usage telemetry shows the removal threshold for the required window.
- Packed-consumer proof covers library exports and declarations when applicable.
- `npm run verify` is green after all obsolete surfaces are removed.
