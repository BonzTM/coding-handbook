# Live Tuning And Telemetry

Defaults for games that measure and adjust after release: what to instrument, how values change in production, and how quantitative data pairs with qualitative evidence before a design change counts as an improvement.

## Default Approach

| Concern | Default | Notes |
|---|---|---|
| instrumentation scope | events that answer a written design question | never "collect everything, ask later" |
| tunable values | remote parameters with committed client defaults | the game must run with fetch failure |
| change rollout | hypothesis first, segment first, then global | one experiment per value at a time |
| interpretation | telemetry paired with a qualitative source | numbers say what; sessions say why |
| data posture | minimum needed for the question, no PII in events | aggregate where the question allows |

Game analytics is the practice of discovering and communicating patterns in player data to drive design and production decisions (Seif El-Nasr, Drachen & Canossa, *Game Analytics: Maximizing the Value of Player Data*, `link.springer.com/book/10.1007/978-1-4471-4769-5`). This doc owns how that practice is applied; the values being tuned are owned by [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) and [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md).

### Instrument The Design Question

- Every instrumented event starts as a written question tied to a pillar or a known risk: name the metric, the decision it informs, and the threshold that would trigger a change. An event no question owns is telemetry debt — do not ship it.
- Baseline event set, each mapped to its owning doc: session start and end; progression checkpoints — start, complete, fail, quit ([../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md)); economy transactions with source, sink, and resulting balance ([../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md)); onboarding funnel steps ([../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)).
- Events carry a stable schema: event name, schema version, session id, build id, active config version, parameters. Never redefine an event's meaning in place — bump the version so historical comparisons stay valid.
- Keep raw player identifiers, chat, and free-text input out of event parameters; the design question is about populations, not persons.

### Remote Tuning Values

- Ship tunable parameters via remote config so values change without a client update, and test changes on a segment before global rollout (`gameanalytics.com/blog/live-ops-remote-configs-ab-testing-games`). The tunable surface is the numeric layer already specified in [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md): costs, rewards, drop rates, timers, difficulty scalars — not logic, not content.
- Every remote value has a committed default in the client, and the client plays correctly on fetch failure using those defaults. A game that needs the config service to boot is a forbidden pattern, not a live-ops feature.
- Clamp server-supplied values to a sane committed range in the client. Bad config must degrade to a badly tuned game, never a broken one.
- Every event logs the active config version, so any cohort can be reconstructed and any tuning change attributed.
- A live tuning change is a design change: it gets the same review as a code change, updates the balance sheet source of truth in the same change, and a change that moves a pillar-level promise is recorded via [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

### A/B Discipline

- One written hypothesis per experiment, before launch: the variant, the decision metric, the expected direction, cohort size, minimum run length, and the ship/kill rule. Telemetry plus large-scale A/B and multivariate experimentation is the established mechanism for tuning retention on live games (`gdcvault.com/play/1016149/Optimization-of-Online-Games-through`; practitioner treatment: `gamedeveloper.com/business/a-b-tests-for-analysing-liveops-part-1`).
- Predeclare the decision point and hold to it. Peeking mid-run and shipping the leading arm turns noise into policy; retention questions need their full day-7 or day-30 window.
- Never run overlapping experiments on the same value or the same funnel step; attribution dies.
- Close experiments completely: ship the winning arm as the committed default, delete the losing arm and its flag in the same change.

### Quantitative-Qualitative Pairing

- The pairing rule: telemetry tells you *what* players do, not *why* — so no telemetry-driven design change is accepted as an improvement until the metric movement is paired with at least one qualitative source that explains it: a playtest session, a session recording review, or structured player feedback. This is the canonical practitioner caveat on game analytics (see the digest of Seif El-Nasr et al. and the practitioner literature above).
- A metric anomaly is a hypothesis, not a finding. Route it to [../quality/playtesting.md](../quality/playtesting.md) — run the session, capture the why in a [../templates/playtest-report.md](../templates/playtest-report.md), then tune.
- The pairing runs both directions: a playtest observation that motivates a tuning change also names the telemetry that will confirm the fix landed at population scale.

### Ethical Guardrails

- Instrument to deliver the experience the pillars promise ([../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md)). Metrics-first optimization of spend and session time against the player's interest is ethically contested in the field and forbidden here: engagement metrics serve the design, not the reverse. The psychology being leveraged is owned by [../foundations/player-psychology.md](../foundations/player-psychology.md), including its dark-pattern boundaries.
- Every A/B arm must be a variant you would ship to every player. Experiments on live players are experiments on people; a "just to see" arm you would not defend publicly does not run.
- Pair every engagement win metric with a counter-metric — churn, refund or complaint rate, session-length distribution tails — so a change that harms a minority of players cannot register as a pure win. This is a convention this handbook adopts.
- Collect the minimum data the question needs, prefer aggregates, honor platform consent requirements, and never degrade gameplay when a player opts out of telemetry.
- Never read low usage of an assistive or accessibility option as license to cut it: option usage measures reach, not necessity ([../disciplines/accessibility.md](../disciplines/accessibility.md)).

## Common Mistakes And Forbidden Patterns

- Collecting everything and hoping questions emerge later — storage-heavy, decision-light, and a privacy liability.
- Shipping a remote-tunable value with no committed client default or no clamp range.
- Redefining an event's meaning without a schema version bump, silently corrupting every historical comparison.
- Declaring a metric movement a win with no qualitative explanation attached.
- Peeking mid-experiment and shipping the leading arm before the predeclared window closes.
- Overlapping experiments touching the same value or funnel step.
- Tuning live values without updating the balance spreadsheet source of truth in the same change.
- Dead experiment flags and orphaned losing arms left in the client.
- A/B arms designed to probe how much friction or spend players will tolerate.
- Player identifiers, free text, or chat content in event parameters.

## Verification And Proof

- every instrumented event maps to a written design question with a metric, decision, and threshold
- the client boots and plays a full session with remote-config fetch forced to fail
- a sampled event stream shows schema version, build id, and config version on every event, and no PII
- the experiment log shows hypothesis, arms, cohort, run length, and decision rule dated before launch
- each shipped tuning change references its paired qualitative source and its balance-sheet update
- closed experiments leave no flags or losing arms in the client

Related: [../quality/playtesting.md](../quality/playtesting.md), [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md), [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md), [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md), [../recipes/balance-an-economy.md](../recipes/balance-an-economy.md), [scoping-and-production.md](scoping-and-production.md).
