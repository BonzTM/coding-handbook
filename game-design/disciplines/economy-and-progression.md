# Economy And Progression

Defaults for designing internal economies and progression systems: model every resource as flows between faucets, pools, and sinks; simulate the flows before building; and pace progression against an interest curve instead of a flat grind.

## Default Approach

- Treat every accumulating quantity — currency, XP, materials, energy, upgrade tiers — as part of one internal economy, and design it as flows, not as isolated reward values. The internal economy is "the flow of tangible and abstract resources" through the game's systems (Adams & Dormans, *Game Mechanics: Advanced Game Design*, ch. 5, `oreilly.com/library/view/game-mechanics-advanced/9780132946728/ch05.html`).
- Specify every resource in the [balance spreadsheet spec](../templates/balance-spreadsheet-spec.md) before it ships: its faucets, its sinks, its intended net flow per session, and the player desire it ultimately serves.
- Number tuning (drop rates, prices, XP tables) is owned by [difficulty-and-balance.md](difficulty-and-balance.md); this doc owns the *structure* of the economy those numbers live in. The end-to-end procedure is [../recipes/balance-an-economy.md](../recipes/balance-an-economy.md).

### Faucets, Pools, And Sinks

- **Faucets** are where resources enter the economy: quest rewards, loot drops, harvesting, timed grants. **Pools** are where they accumulate: wallets, inventories, XP totals. **Sinks** are where they permanently exit: fees, repair costs, consumables, decay, crafting inputs. The design task is balancing the flows: if faucets outrun sinks, the currency inflates toward worthlessness (Game Developer, "The F-Words Of MMOs: Faucets", `gamedeveloper.com/design/the-f-words-of-mmos-faucets`).
- Every faucet ships with at least one sink for the same resource in the same release. A faucet with no sink is a deferred inflation incident, not a reward.
- Spending resources on trade between players or on reversible conversions is not a sink — nothing left the economy. Only permanent removal counts when auditing net flow.
- Ultima Online is the founding cautionary case: an open player-driven economy without adequate drains forced retrofitted sinks such as item decay and NPC-only goods (`en.wikipedia.org/wiki/Gold_sink`). Retrofitting sinks into a live economy is far more painful than shipping them up front, because players experience a new sink as a nerf.

### Value Chains

- Structure each resource as a **value chain**: faucet -> transformations (crafting, conversion, upgrade) -> sink, where the terminal sink delivers something the player actually wants. Daniel Cook's "Value chains" is the working method for constructing and balancing faucet-and-drain economies this way (`lostgarden.com/2021/12/12/value-chains/`).
- A chain that terminates in nothing the player values is dead weight: the resource piles up in its pool and reads as clutter. Trace every chain to a motivation named in [../foundations/player-psychology.md](../foundations/player-psychology.md) — mastery, status, self-expression, progress — or cut the resource.
- Keep chains short by default. Every transformation step is a tuning surface and a tutorialization cost; add a step only when it creates a decision, not just a click. Interesting-decision criteria live in [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md).
- Multiple currencies are a scoping decision, not a default. Each additional currency multiplies the flow-audit surface; add one only when it must be earned and spent on an isolated cadence, and record the reason per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

### Modeling Before Building

- Diagram and **simulate** the economy before writing game code. The Machinations diagram language — pools, sources, drains, converters, gates — exists precisely so an economy's dynamic behavior can be simulated and stress-tested before implementation (Joris Dormans, 2012; Adams & Dormans ch. 5; Ernest Adams, "Machinations, A New Way to Design Game Mechanics", `gamedeveloper.com/design/the-designer-s-notebook-machinations-a-new-way-to-design-game-mechanics`). Tool selection (machinations.io vs. a spreadsheet simulation) is routed through [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md).
- Minimum simulation set before any economy ships: a median player at intended pace, a no-spend floor player, and a maximum-efficiency optimizer running every faucet at cap. If the optimizer's pool curve diverges upward without bound, the economy is broken before launch.
- Simulate in game-hours and sessions, not abstract ticks, so the output answers real questions: "how many sessions to afford the tier-2 upgrade?" is a designable number, not an emergent surprise.
- The simulation is a living artifact: it lives alongside the [balance spreadsheet spec](../templates/balance-spreadsheet-spec.md) and every faucet or sink change re-runs it before the change merges. Simulation predicts; only playtests confirm — validate the modeled pace against real players per [../quality/playtesting.md](../quality/playtesting.md).

### Progression Pacing And Interest Curves

- Pace progression as an **interest curve**, not a monotonic ramp: open with a hook, alternate peaks with rest valleys, and escalate to a climax. The pattern is fractal — it applies to a session, a level, and the whole game (Jesse Schell, *The Art of Game Design*, interest-curve chapter, `routledge.com/The-Art-of-Game-Design-A-Book-of-Lenses-Third-Edition/Schell/p/book/9781138632059`).
- Progression is the delivery schedule for new patterns to learn. Koster's fun-as-learning model sets the failure condition: when the game has no more patterns to teach, it becomes boring (*A Theory of Fun*, `theoryoffun.com`) — so unlock cadence must front-run the point where the current toolset is mastered. The loop being paced is owned by [../foundations/core-loops.md](../foundations/core-loops.md).
- Make each unlock change what the player *does*, not only the numbers. A +5% stat unlock is a pool transaction; a new verb or a new decision is progression.
- Chart the intended pacing: time-to-unlock per milestone at the median simulated pace, plotted against the interest curve. Long flat stretches between peaks are grind by construction — fix them in the model, not with a post-launch faucet buff.
- Difficulty must rise with player power or progression cancels itself into a flat experience; the challenge-side curve is owned by [difficulty-and-balance.md](difficulty-and-balance.md) and tuned via [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md).

### Inflation Watchpoints

- Instrument the economy's design questions from day one, per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md): per-currency faucet volume, sink volume, and net flow per cohort; pool size distribution over time; sink participation rate.
- Watch the top-wealth cohort, not the median. Inflation starts where optimizers concentrate, then repricing against their stockpiles punishes everyone else.
- A sink nobody uses is a faucet problem in disguise: either the sink's output is not desired (dead value chain) or the resource is so abundant that spending it is meaningless.
- Ship economy values as tunable configuration so faucet and sink rates can move without a client update; staged-rollout and A/B mechanics are owned by [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).
- Prefer widening sinks over cutting faucets when correcting live inflation: players experience a faucet cut as taking away income, while a desirable new sink removes currency voluntarily. Either way, telemetry says what changed, not why — pair any correction with playtest or player-facing research before and after.

## Common Mistakes And Forbidden Patterns

- Shipping a faucet with no sink for the same resource, or counting player-to-player trade as a sink.
- Designing rewards as isolated values ("the quest gives 50 gold") with no flow model showing where the 50 gold goes to die.
- Building the economy in code first and discovering its dynamic behavior from live players instead of from a simulation.
- Balancing only for the median player and letting the maximum-efficiency optimizer define the real economy.
- A resource whose value chain terminates in nothing the player wants, left in the game as pool clutter.
- Adding currencies by reflex — every new currency multiplies the audit surface without an isolated-cadence justification.
- Progression that is a monotonic stat ramp: no rest valleys, no new verbs, no interest-curve peaks — numbers go up while the play stays identical.
- Unlock pacing that trails mastery, leaving players with a fully-learned toolset and nothing new to learn.
- Hard-coded economy values that require a client build to retune.
- Correcting live inflation by silently cutting faucet output with no telemetry baseline and no follow-up measurement.
- Restating tuning rules here — number-fitting and cost-curve method belong to [difficulty-and-balance.md](difficulty-and-balance.md).

## Verification And Proof

- Every resource in the design has a completed row in the [balance spreadsheet spec](../templates/balance-spreadsheet-spec.md): faucets, sinks, intended net flow, terminal player value.
- A Machinations diagram or spreadsheet simulation exists for the economy, is checked in alongside the spec, and was re-run for the current change.
- Simulation output for median, no-spend, and optimizer profiles shows bounded pool curves — no divergent stockpile at any profile.
- Time-to-unlock per milestone at simulated median pace is charted against the intended interest curve, with no unplanned flat stretch.
- Playtest evidence ([../quality/playtesting.md](../quality/playtesting.md)) confirms real-player pace lands within tolerance of the simulated median.
- Telemetry emits per-currency faucet, sink, and net-flow measures per cohort, and the live dashboard shows sink participation rate ([../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)).
- Economy values load from tunable configuration, proven by changing one rate without a client rebuild.
- Design review ([../checklists/design-review.md](../checklists/design-review.md)) signed off that each transformation step in every value chain creates a decision, not a click.

## Related

- [../recipes/balance-an-economy.md](../recipes/balance-an-economy.md) - the step-by-step procedure this doc governs.
- [difficulty-and-balance.md](difficulty-and-balance.md) - cost curves and number tuning for the objects inside the economy.
- [../foundations/core-loops.md](../foundations/core-loops.md) - the loop the economy feeds and pays out from.
- [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md) - instrumenting and safely changing a live economy.
