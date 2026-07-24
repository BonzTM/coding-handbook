# Difficulty And Balance

The standardized balance method and difficulty policy: how transitive systems are balanced against a cost curve, how competitive systems stay balanced through viable options and counterplay, and where this handbook stands on dynamic difficulty.

## Default Approach

Balance is a method, not a feel. Classify the system first, because the two families of balance problem have different tools: **transitive** systems (some options are strictly better and priced accordingly — RPG gear tiers, upgrade tracks) are balanced with cost curves; **intransitive** systems (options beat each other in a cycle — matchups, unit counters) are balanced by keeping every option viable with counterplay. The method is Ian Schreiber's Game Balance Concepts curriculum (`gamebalanceconcepts.wordpress.com/2010/07/21/level-3-transitive-mechanics-and-cost-curves/`, expanded in Schreiber & Romero, *Game Balance*, CRC Press 2021); the competitive stance is David Sirlin's (`sirlin.net/articles/balancing-multiplayer-games-part-2-viable-options`). Numbers live in the committed balance model, not in heads or in scattered engine defaults — see [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md).

Balance work is downstream of the resource model. The flows a cost curve prices (currencies, drops, upgrade costs) are owned by [economy-and-progression.md](economy-and-progression.md); do not re-derive faucet/sink policy here.

### Cost Curves

- **Enumerate costs and benefits for every balanceable object.** Everything the player gives up is a cost (resource price, opportunity cost, drawback, setup requirement); everything gained is a benefit. Hidden costs and benefits are the usual source of "mystery" imbalance — list them explicitly in the balance sheet.
- **Plot cost against benefit and fit the curve.** The fitted curve is the pricing contract: a new object is placed on the curve first and hand-tuned second. An object above the curve is undercosted (dominant); below it is overcosted (dead content).
- **Expect benefits to outpace costs late-game.** Schreiber's observation from CCG data is that in transitive systems the curve steepens — top-end objects deliver disproportionate benefit for their cost — so late-game pricing needs its own fit, not a linear extrapolation of early-game values.
- **One curve per resource domain.** Objects priced in different currencies or acquisition paths need an explicit exchange assumption before they can share a curve; that exchange rate is an economy decision owned by [economy-and-progression.md](economy-and-progression.md).
- **The curve is a committed artifact.** It lives in the balance spreadsheet with named columns for every cost and benefit term, so a reviewer can recompute any object's position. Changing the curve is a design decision recorded per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

### Viable Options And Counterplay

- **Many viable options is the goal; a dominant option is the failure.** Sirlin's rule: a strictly-dominant move destroys strategy, because every decision that included it stops being a decision. Balance passes hunt dominance first, dead options second.
- **Fairness is defined at the start line.** Sirlin: "players of equal skill have an equal chance to win, no matter their start conditions." Asymmetric starts (characters, factions, loadouts) are allowed; asymmetric win chances at equal skill are not.
- **Build intransitive (counter) relationships on purpose.** A rock-paper-scissors structure among options is Schreiber's tool for keeping multiple options viable without pricing them identically: each option's weakness is another option's job.
- **Keep the counter-cycle at least three deep.** Sirlin's Yomi-layer construction — a counter, a counter to the counter, and a third option that beats the second — keeps mind-games live even when tuning is imperfect. A two-option counter pair collapses into a coin flip; a three-plus cycle is resilient to tuning error.
- **Counterplay must be executable, not just theoretical.** A counter the player cannot see coming, cannot afford, or cannot execute at realistic skill is not a counter; readability of the threat is a [../foundations/game-feel.md](../foundations/game-feel.md) and [ux-and-onboarding.md](ux-and-onboarding.md) concern that balance depends on.

### Difficulty Curve Design

- **The target is the flow channel: challenge tracks skill.** Csikszentmihalyi's flow model (see [../foundations/player-psychology.md](../foundations/player-psychology.md)) supplies the rationale — challenge far above skill produces anxiety, far below produces boredom — so the difficulty curve rises as player skill rises, with deliberate rest valleys, not monotonically.
- **Teach before you test.** Per-level pacing follows the kishōtenketsu introduce → develop → twist → prove structure owned by [level-design.md](level-design.md); a difficulty spike is only legitimate after the skill it tests has been introduced and practiced.
- **Design self-selection points into the game.** Jenova Chen's "Flow in Games" thesis (`jenovachen.com/flowingames/Flow_in_games_final.pdf`) argues for offering in-fiction choices that let players adjust challenge themselves (optional hard routes, skippable challenges, loadout risk). Prefer these embedded choices over relying solely on a menu setting.
- **Tune the curve from evidence, not intuition.** Fail-point clustering, completion times, and quit points come from playtests ([../quality/playtesting.md](../quality/playtesting.md)) and, post-launch, telemetry ([../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)). The working procedure is [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md).
- **Difficulty options are not accessibility features.** Difficulty modes tune challenge for players who can already execute the inputs; input, presentation, and cognitive accommodations are owned by [accessibility.md](accessibility.md) and are never gated behind a difficulty setting.

### Dynamic Difficulty Policy

Dynamic difficulty adjustment is contested among practitioners — hidden DDA can feel patronizing when noticed and is exploitable in anything score- or skill-expressive, and competitive designers broadly reject it. This handbook's stance:

- **Default: player-visible difficulty options plus Chen-style implicit self-selection.** These are the uncontested forms; use them first.
- **Hidden DDA only with a decision record.** If a project adopts concealed adjustment (rubber-banding, adaptive spawns, pity timers), the DDR must state what is adjusted, the bounds of adjustment, and why visibility would harm the experience — route it through [../decisions/design-decision-records.md](../decisions/design-decision-records.md).
- **Never in competitive modes.** Hidden adjustment in any player-vs-player or leaderboard context violates the equal-chance fairness definition above.
- **Bound every adjuster.** A DDA system has explicit floor and ceiling values in the balance model, so it can ease a struggling player without erasing the challenge or inflating a skilled player's difficulty past the tuned curve.

## Common Mistakes And Forbidden Patterns

- Balancing transitive content by feel with no committed cost curve, so every new object is priced by argument instead of arithmetic.
- Pricing an object while ignoring non-resource costs (drawbacks, setup, opportunity cost), then declaring the curve broken when the object dominates.
- Extrapolating early-game pricing linearly into the late game, where benefits outpace costs.
- Shipping a strictly-dominant option and compensating with a price increase instead of adding a counter — dominance is a structure problem before it is a number problem.
- Two-option counter pairs presented as depth; the cycle needs a third live option.
- "Balancing" by making every option identical — mirror balance is the last resort, not the goal (equal win chance, not equal stats).
- Difficulty spikes that test a skill the game never taught (a [level-design.md](level-design.md) sequencing failure surfacing as a balance complaint).
- Hidden DDA added without a DDR, or any DDA in a competitive or leaderboard mode.
- Using easy mode as the accessibility plan; accommodations belong in [accessibility.md](accessibility.md).
- Reacting to a single loud playtest complaint or forum thread instead of clustered evidence from [../quality/playtesting.md](../quality/playtesting.md) and telemetry.
- Tuning live values ad hoc in the build instead of through the remote-config and rollout process owned by [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).

## Verification And Proof

- Every balanceable object appears in the balance spreadsheet with all cost and benefit terms filled in, and its position relative to the fitted curve is recomputable by a reviewer ([../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md)).
- For each intended counter relationship, a playtest or simulation shows the counter actually winning the matchup at realistic skill — and the cycle has at least three live options.
- Pick/win/usage data (playtest tallies pre-launch, telemetry post-launch) shows no option strictly dominant and no option dead; outliers have an open tuning task or a DDR explaining why they stand.
- Difficulty-curve evidence exists per chapter/level: fail-point and quit-point data from [../quality/playtesting.md](../quality/playtesting.md) sessions, reviewed against the intended curve in [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md).
- Any hidden-DDA system has a DDR on file stating scope and bounds, and its floor/ceiling values are present in the balance model.
- Live tuning changes ship through the segment-first rollout in [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md), with the metric that justified the change named in the change note.
