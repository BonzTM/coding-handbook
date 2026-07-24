# Concept Intake Checklist

Pre-design gate run BEFORE any system design, loop diagramming, or prototyping starts. The handbook supplies the HOW — pillars, loops, playtest protocol, the milestone ladder — and this checklist covers the WHAT the concept must supply. Each item names the doc that consumes the answer. A box is "answered" only when the answer is concrete enough to design against (a written goal, a named platform, a number), not "TBD" or "we'll see in production".

## How To Resolve Each Box

Resolve every applicable box in this order. Intake never stalls the project; it decides *where each answer comes from* and makes that traceable.

1. **From the pitch.** If the concept document or requester's brief already answers it, record the answer and move on.
2. **Ask.** Asking beats inferring. When the requester is reachable, ask the open questions that materially change the design — batched, once, up front. Do not interrogate box-by-box, and do not ask about boxes the defaults table below already covers unless the answer would change the game's shape.
3. **Default.** When the requester is unreachable, has said "just make it", or the box is low-stakes for a prototype: take the entry from [Defaults When The Pitch Is Silent](#defaults-when-the-pitch-is-silent), record it as a stated assumption, and proceed. Two boxes are irreversible-grade — **single-player vs multiplayer** and **premium vs live-service** — because they reshape every system downstream (economy sinks, difficulty tuning, telemetry, production plan). Ask about them whenever interaction is possible; when it is not, take the default and flag the assumption at the top of the one-pager, not buried in a decision record appendix.
4. **Skip.** Sections marked as not applying to the project's shape are skipped, not answered. A weekend jam entry has no funding milestone ladder. Do not ask questions the shape makes meaningless.

Never invent an answer that is neither in the pitch, nor from the requester, nor in the defaults table. Every defaulted answer is disclosed on the concept one-pager so the requester can veto it cheaply; irreversible-grade defaults get a design decision record per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

## Section Applicability By Shape

| Section | Jam / experiment | Premium release | Live-service |
|---|---|---|---|
| Experience Goals | yes | yes | yes |
| Audience And Constraints | platform, input, and accessibility rows only | yes | yes |
| Scope Signals | riskiest-assumption and bounded-scope rows only | yes | yes, including the live rows |
| Open Questions | yes | yes | yes |

## Experience Goals

- [ ] Two to four experience goals are written as player experiences — what the player feels, decides, or masters — not feature inventories. Setting experience goals before designing mechanics is the founding move of Fullerton's playcentric process; the goal format and pillar rules are owned by [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md).
- [ ] Each goal is phrased so a playtest with fresh players could refute it. A goal that cannot fail in a session per [../quality/playtesting.md](../quality/playtesting.md) is decoration, not a goal.
- [ ] Three to five pillars are derived from the goals, each able to reject a feature ([../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md)). Genre labels and quality adjectives are not pillars.
- [ ] Aesthetic targets are named in the shared MDA-derived vocabulary owned by [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md), not per-project adjectives.
- [ ] The hypothesized core loop is nameable in one sentence — the repeatable action-feedback-repeat cycle, not a feature list. Intake needs only the hypothesis; designing and diagramming it happens later via [../foundations/core-loops.md](../foundations/core-loops.md) and [../recipes/design-a-core-loop.md](../recipes/design-a-core-loop.md).

## Audience And Constraints

- [ ] The target audience is described by motivations that audience demonstrably has — using the empirical motivation models (Quantic Foundry's twelve motivations, SDT's autonomy/competence/relatedness) owned by [../foundations/player-psychology.md](../foundations/player-psychology.md) — not by demographics or by motivations the team hopes for.
- [ ] Two to four comparable titles are listed, with what their players expect and what this concept deliberately breaks. Comparables are the cheapest calibration for audience, session shape, and scope.
- [ ] Target platform(s) and input methods are named. Input drives the control-feel budget ([../foundations/game-feel.md](../foundations/game-feel.md)) and the onboarding surface ([../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)).
- [ ] Expected session length and play context are stated (ten-minute commute sessions and two-hour couch sessions produce different loops and different save models).
- [ ] The accessibility floor is committed at concept time, not patched later — design-time accessibility is the stated posture of AbleGamers APX (`ablegamers.org/apx/`). Minimum bar: input remapping, readable text size, colorblind-safe signaling, and subtitle presentation — the four most commonly complained-about issues per `gameaccessibilityguidelines.com`. The full tiered guideline set is owned by [../disciplines/accessibility.md](../disciplines/accessibility.md).
- [ ] Team size, budget ceiling, and target date are written down. These are design inputs, not production trivia — they bound the loop-versus-content mix below ([../operations/scoping-and-production.md](../operations/scoping-and-production.md)).

## Scope Signals

- [ ] The concept is classified along Daniel Cook's loops-versus-arcs split (`lostgarden.com/2012/04/30/loops-and-arcs/`): how much is repeatable, mastery-driven systems (loops) versus consumed-once content (arcs). Arcs cost per minute of play; loops amortize. An arc-heavy concept on a small team is a scope red flag that intake must surface, not bury.
- [ ] The riskiest assumption is named — the one claim that, if false, kills the concept — and it is testable by a cheap paper or graybox prototype per [../quality/prototyping.md](../quality/prototyping.md) before any production-quality work.
- [ ] The milestone ladder expectation is set: prototype, then first playable, then vertical slice. Rami Ismail's framing (`ltpf.ramiismail.com/prototypes-and-vertical-slice/`): prototypes answer whether you *should* make the game; the vertical slice answers whether you *can*. Which rungs this project needs, and whether a slice is a funding artifact or skipped, is owned by [../operations/scoping-and-production.md](../operations/scoping-and-production.md).
- [ ] Single-player or multiplayer is decided — irreversible-grade; multiplayer pulls fairness and viable-options balance work into scope from day one ([../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md)).
- [ ] Premium (shipped-complete) or live-service is decided — irreversible-grade; live-service pulls in economy sink/faucet planning ([../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md)) and telemetry and remote-tuning design ([../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)) at concept time.
- [ ] The v1 feature set is bounded in writing: what ships, and explicitly what does not. Scope creep mid-production is the default failure mode intake exists to prevent.

## Open Questions

Every box not resolved from the pitch becomes an open question. Batch them into one message to the requester; do not trickle questions as design proceeds. Questions still open when design must start take the defaults below.

### Defaults When The Pitch Is Silent

"Record as" says where the assumption must be disclosed: **DDR** means a design decision record per [../decisions/design-decision-records.md](../decisions/design-decision-records.md); **one-pager** means a line in the assumptions list on the concept one-pager ([../templates/one-page-gdd.md](../templates/one-page-gdd.md)). Anything not in this table has no silent default — it comes from the pitch or the requester. Experience goals in particular are never silently defaulted: if the pitch lacks them, draft them from the pitch's strongest fantasy, put them at the top of the one-pager, and invite veto.

| Decision | Default when the pitch is silent | Record as |
|---|---|---|
| Audience | players of the named comparable titles, with motivations read from those titles per [../foundations/player-psychology.md](../foundations/player-psychology.md) | one-pager |
| Platform and input | the team's established platform; with zero signal, PC with full keyboard/mouse and gamepad support | one-pager |
| Session length | the median session shape of the comparables | one-pager |
| Accessibility floor | the four-issue minimum bar above, tiered up later per [../disciplines/accessibility.md](../disciplines/accessibility.md) | one-pager |
| Single vs multiplayer | single-player | DDR, flagged at the top of the one-pager |
| Premium vs live-service | premium, shipped-complete, no live-ops | DDR, flagged at the top of the one-pager |
| Milestone ladder | prototype and first playable always; vertical slice only if external funding or greenlight requires one | one-pager |
| First prototype form | paper or graybox against the riskiest assumption, per [../quality/prototyping.md](../quality/prototyping.md) | one-pager |

## Proof

Intake is complete, and design may start, when:

- every applicable box is resolved — from the pitch, from the requester, or from the defaults table — and skipped sections are skipped because of shape, not convenience;
- the concept one-pager exists per [../templates/one-page-gdd.md](../templates/one-page-gdd.md) (produced via [../recipes/write-a-one-page-gdd.md](../recipes/write-a-one-page-gdd.md)), carrying the goals, pillars, loop hypothesis, audience, constraints, and the assumptions list — one page because design documentation exists to communicate efficiently and "most people only read the first page anyway" (Stone Librande, "One-Page Designs", GDC 2010, `gdcvault.com/play/1012356/One-Page`);
- both irreversible-grade decisions (player count, premium vs live-service) are either answered by the requester or defaulted with a DDR and a flag at the top of the one-pager;
- open questions were batched and sent once, and the riskiest assumption has a named prototype that will test it;
- the first design review can cite this intake: [design-review.md](design-review.md) assumes goals, pillars, and bounded scope already exist.

An unanswerable box is never a silent stall: ask when you can, default and disclose when you cannot.
