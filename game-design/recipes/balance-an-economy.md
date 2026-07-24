# Recipe: Balance An Economy

Use this when adding, removing, or retuning a resource, currency, reward source, or spend point. Rules are owned by [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md); this recipe is the procedure.

## Files To Touch

- the balance spreadsheet, per [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md)
- the economy flow diagram (pools, faucets, sinks) owned by [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md)
- the telemetry spec for inflation indicators, per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)
- a DDR via [../decisions/design-decision-records.md](../decisions/design-decision-records.md) if the change touches an economy invariant

## Steps

1. Model the change as flows in the diagram: name every faucet the resource enters through (rewards, drops, harvesting) and every sink it exits through (fees, repairs, consumables, decay). Balancing the two flows is the design task; a faucet with no downstream sink is the defect, not a tuning problem (Daniel Cook, "Value chains", `lostgarden.com/2021/12/12/value-chains/`).
2. Enter costs and benefits for every changed object in the balance spreadsheet and plot them against the cost curve; an object off the curve is over- or under-priced until the deviation is deliberate and documented (Ian Schreiber, Game Balance Concepts Level 3, `gamebalanceconcepts.wordpress.com/2010/07/21/level-3-transitive-mechanics-and-cost-curves/`).
3. Simulate before shipping values: run the Machinations model or the spreadsheet's flow tab across low, median, and high engagement player profiles for a fixed session count and record net currency drift per profile (Adams & Dormans, *Game Mechanics: Advanced Game Design*, ch. 5).
4. Tune until drift lands in the target band. When net inflow is positive, prefer adding or widening sinks over cutting faucets — faucet cuts punish earning, sink additions give players something to want.
5. Define the inflation indicators the change will be watched by: net currency per player per session, currency held by cohort percentile, sink participation rate, and median player-trade price if trading exists.
6. Ship the values as remote-config tunables on a test segment before global rollout, per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).
7. Record the predicted drift and the rationale in the spreadsheet changelog, and in a DDR if an invariant moved.

## Invariants To Preserve

- every faucet has at least one reachable sink; unsunk currency inflates until it is worthless (Ultima Online is the canonical failure case — `en.wikipedia.org/wiki/Gold_sink`)
- the spreadsheet is the single source of truth for values; shipped configs are exported from it, never hand-edited
- no strictly dominant purchase — intransitive relationships between options stay intact, per [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md)
- progression pacing is unchanged unless the change intends it and the DDR says so

## Proof

- simulation output showing net drift within the target band for all three engagement profiles
- cost-curve plot with every changed object on the curve, or its deviation documented
- inflation indicators live on the telemetry dashboard before the values ship
- segment result reviewed before global rollout, per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)
