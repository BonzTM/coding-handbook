# Player Psychology

The motivation models this handbook standardizes on: self-determination needs as the base layer, empirical motivation profiles instead of legacy type taxonomies, and flow as the contract between challenge and skill.

## Default Approach

Design decisions that claim to be "for the player" must name which need or motivation they serve, in the vocabulary below. Three models are standard: self-determination theory for *why* play is satisfying, the Quantic Foundry motivation dimensions for *which* satisfactions a given audience seeks, and flow for *how hard* the game should push at any moment. Anything outside these three routes through [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md).

### Self-Determination Needs

The base model is self-determination theory (SDT) applied to games: perceived in-game **autonomy** and **competence** predict enjoyment, preference for future play, and well-being changes, and **relatedness** independently predicts enjoyment in multiplayer contexts (Ryan, Rigby & Przybylski, "The Motivational Pull of Video Games", *Motivation and Emotion* 30, 2006 — `link.springer.com/article/10.1007/s11031-006-9051-8`). This is the strongest academically grounded motivation model available; the PENS model derives from the same line of work.

Use the three needs as a review lens on every system:

- **Competence** — the game must let the player feel effective. The same study tied competence perception to intuitive controls; competence is delivered by readable feedback and a mastery curve, which is why this need is co-owned in practice by [core-loops.md](core-loops.md) (Koster's fun-as-learning grounds the loop) and [game-feel.md](game-feel.md) (feedback clarity).
- **Autonomy** — meaningful choice over goals and approach. Forced paths, single dominant strategies, and interrupt-driven engagement mechanics all spend autonomy. Viable-option balance work in [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) is autonomy work.
- **Relatedness** — mattering to others, human or believable NPC. Only budget for it where the design actually has a social surface; a single-player game does not fail review for lacking it.

A feature that satisfies none of the three needs is a candidate for cutting, not for more polish.

### Motivation Profiles

For describing audiences, default to the Quantic Foundry Gamer Motivation Model: 12 motivations in 6 correlated pairs (Action, Social, Mastery, Achievement, Immersion, Creativity), derived from survey data on 400,000+ gamers (Yee & Ducheneaut — reference chart `quanticfoundry.com/2015/12/15/handy-reference/`, GDC 2019 "A Deep Dive into the 12 Motivations" `gdcvault.com/play/1025742/A-Deep-Dive-into-the/`).

Rules of use:

- Motivations are **continuous dimensions, not buckets**. A target audience is a profile — high Mastery, mid Immersion, low Social — never "our players are Achievers."
- Write the target profile into the concept intake and pillar work ([design-pillars-and-vision.md](design-pillars-and-vision.md), [../checklists/concept-intake.md](../checklists/concept-intake.md)) so scope debates can appeal to it: a feature serving a motivation the profile ranks low is off-profile until an explicit pillar change says otherwise.
- Bartle's 1996 Achiever/Explorer/Socializer/Killer taxonomy (`mud.co.uk/richard/hcds.htm`) is historical vocabulary, not a design input. It was derived from observation of MUD players, not survey psychometrics, and Bartle himself cautioned against over-generalizing beyond virtual worlds. Cite it when reading older literature; do not segment a modern audience with it.

### Flow And Challenge

Flow is the handbook's challenge-skill contract: optimal experience requires challenge balanced against skill, clear goals, focused attention, and continuous feedback — too much challenge produces anxiety, too little produces boredom (Csikszentmihalyi, *Flow: The Psychology of Optimal Experience*, 1990 — `archive.org/details/flowpsychologyof00csik`). Jenova Chen's "Flow in Games" thesis (`jenovachen.com/flowingames/Flow_in_games_final.pdf`) applies this to games: keep players inside the flow channel, and prefer giving players choices that let them self-select difficulty over hidden system-driven adjustment.

Operational consequences:

- Every difficulty curve is a flow-channel claim: challenge tracks the skill the game has actually taught, in the order it taught it. The tuning mechanics live in [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) and [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md); this doc owns the rationale.
- Clear goals and continuous feedback are flow preconditions, so onboarding and moment-to-moment readability ([../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md), [game-feel.md](game-feel.md)) are motivation work, not cosmetic work.
- Hidden dynamic difficulty adjustment is a contested practice — it can feel patronizing or exploitable, and many competitive designers reject it. Player-visible difficulty options and Chen-style implicit self-selection are the defaults; hidden DDA requires a design decision record ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)).

### Applying Models Without Stereotyping

These models describe tendencies in populations; they do not predict an individual playtester, and they never override observed behavior.

- Models generate hypotheses; playtests check them. A motivation-profile claim ("our audience is high Mastery, so they will tolerate this failure rate") is testable and must be tested — protocol in [../quality/playtesting.md](../quality/playtesting.md).
- When telemetry and the model disagree, the model loses. Telemetry tells you what players do, not why, so pair it with qualitative sessions before rewriting the profile — see [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).
- Needs-based framing is also the ethical guardrail: engagement built on satisfying competence, autonomy, and relatedness is durable; engagement built on interrupting or obligating the player spends autonomy and shows up as churn. If a retention mechanic cannot be described as serving one of the three needs, escalate it to design review ([../checklists/design-review.md](../checklists/design-review.md)) rather than shipping it by default.
- One vocabulary, one owner: the terms above are defined once in [../glossary.md](../glossary.md); do not coin per-project synonyms for the same constructs.

## Common Mistakes And Forbidden Patterns

- Segmenting an audience with Bartle types ("we need content for the Explorers") instead of a continuous motivation profile.
- Treating a motivation profile as a per-player predictor and overriding what a live playtester actually did with what the model says they should want.
- Justifying a feature with "players will love it" and no named need or motivation dimension — an unfalsifiable claim that cannot lose a scope debate.
- Raising challenge without checking what the game has taught: a spike that outruns taught skill is an anxiety exit from the flow channel, not "hardcore appeal."
- Shipping hidden DDA without a decision record, or using it to mask a broken difficulty curve.
- Retention mechanics that work by obligation or interruption (appointment pressure, loss-on-absence) defended as "engagement" with no needs-based account.
- Citing SDT, flow, or the motivation model from memory instead of the primary sources above, then designing against a misremembered version.

## Verification And Proof

Player-psychology work is done when:

- the target motivation profile is written down in the concept intake / pillars doc as ranked continuous dimensions, with the source of the ranking (comparable-title data, survey, or explicit assumption flagged for test).
- every major system in the design doc names the need(s) it serves — competence, autonomy, relatedness — and features serving none carry a cut-or-justify note.
- each motivation claim that shaped a design decision has a matching playtest hypothesis or telemetry question logged per [../quality/playtesting.md](../quality/playtesting.md) and [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).
- the difficulty curve has been reviewed as a flow-channel claim: challenge steps map to skills the game demonstrably taught earlier, verified in [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md).
- any hidden difficulty adjustment has a design decision record; player-facing difficulty controls are the documented default otherwise.

## Where To Go Next

- [core-loops.md](core-loops.md) — the loop structures that deliver competence through learning and mastery.
- [game-feel.md](game-feel.md) — the feedback layer that makes competence perceptible.
- [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) — turning the flow contract into tuned numbers.
- [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md) — teaching skill fast enough for challenge to track it.
- [../quality/playtesting.md](../quality/playtesting.md) — testing motivation hypotheses against real players.
