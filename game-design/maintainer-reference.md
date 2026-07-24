# Maintainer Reference

Purpose: hold slower-path architecture, discipline-map, lifecycle, and rationale guidance that is useful but not worth loading for every task.
Audience: maintainers and agents working on game designs that use this handbook.
Read [AGENTS.md](AGENTS.md) first. Use this file when you need the fuller background behind the fast-path rules.

## Architecture Snapshot

This handbook assumes a game project whose design state lives in a small, versioned artifact corpus rather than in one monolithic GDD — a convention this handbook adopts, anchored in Stone Librande's one-page-designs argument that design documentation exists to communicate ideas efficiently and that most readers never get past the first page (`gdcvault.com/play/1012356/One-Page`). The dominant shape is:

```text
game-repo/
  design/
    pillars.md              design pillars and experience goals
    game-one-pager.md       the game identity one-pager
    <system>-one-pager.md   one-page GDDs, one per system
    decisions/              design decision records (DDRs)
    playtests/              one script + one report per session
    balance/                balance spreadsheet specs and exports
  data/                     tunable values as data, not code constants
```

Skeletons for every artifact above live in [templates/](templates/); [templates/README.md](templates/README.md) maps each template to its destination. Reference exemplar projects (a complete worked design corpus for a small game) are a planned later phase.

## Two-Speed Documentation Model

- Fast path: [AGENTS.md](AGENTS.md) for invariants, the task loop, change-type-to-file-set routing, and baseline proof.
- Slow path: this file for architecture, discipline map, proof taxonomy, lifecycle, and rationale.

Use the fast path for most tasks. Use this file when a change crosses disciplines, challenges an existing default, or needs the sourcing behind a rule.

## Discipline Map

| Discipline Area | Owns | Must Not Own |
|---|---|---|
| [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md) | pillars, experience goals, the veto test for features | feature lists, scope and milestone decisions |
| [foundations/core-loops.md](foundations/core-loops.md) | loop diagrams, skill atoms, the loop/arc split | economy tuning numbers, level layouts |
| [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md) | mechanic specs, system interaction rules, MDA vocabulary | narrative content, per-level pacing |
| [foundations/player-psychology.md](foundations/player-psychology.md) | motivation models (SDT, flow, Quantic Foundry), when to apply which | difficulty numbers, monetization decisions |
| [foundations/game-feel.md](foundations/game-feel.md) | feel and juice defaults, the readability constraint | core mechanic selection |
| [foundations/design-documentation.md](foundations/design-documentation.md) | artifact formats, the one-page rule, doc lifecycle | the design decisions the artifacts record |
| [disciplines/level-design.md](disciplines/level-design.md) | level structure, kishōtenketsu pacing, wordless guidance | story ownership, systemic balance |
| [disciplines/narrative-integration.md](disciplines/narrative-integration.md) | story-mechanics fit, arc content planning | loop design, level metrics |
| [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md) | faucet/sink modeling, value chains, Machinations diagrams | live pricing experiments |
| [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md) | cost curves, viable-options analysis, difficulty curve shape | telemetry pipeline design |
| [disciplines/ux-and-onboarding.md](disciplines/ux-and-onboarding.md) | usability, cognitive-load rules, tutorial technique | accessibility tiers and minimum bar |
| [disciplines/accessibility.md](disciplines/accessibility.md) | Game Accessibility Guidelines tiers, APX patterns, the four-issue minimum bar | general usability rules |
| [disciplines/multiplayer-and-social.md](disciplines/multiplayer-and-social.md) | PvP fairness, matchmaking, asymmetric modes, disruptive-behavior prevention, spectator design | netcode and transport, competitive balance rules, motivation models |
| [quality/prototyping.md](quality/prototyping.md) | prototype fidelity ladder, paper-first default, one-question rule | funding milestones, playtest protocol |
| [quality/playtesting.md](quality/playtesting.md) | session protocol, Kleenex-tester rule, observation-over-interrogation | telemetry, prototype construction |
| [operations/scoping-and-production.md](operations/scoping-and-production.md) | milestone ladder, vertical-slice decision, cut policy | design content of any system |
| [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md) | remote config, A/B testing, cohort metrics, the what-not-why caveat | initial balance, playtest protocol |

## Lifecycle Model

The design lifecycle this handbook assumes is Fullerton's playcentric process (set experience goals first, then prototype, playtest, and revise against those goals continuously — `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`) joined to the standard production milestone ladder:

1. Concept intake: pillars, experience goals, and a one-page concept, gated by [checklists/concept-intake.md](checklists/concept-intake.md).
2. Prototype the core loop at the cheapest fidelity that can answer one named design question — paper first ([quality/prototyping.md](quality/prototyping.md)).
3. Playtest against the experience goals, revise, and repeat until the loop proves out or the concept is killed ([quality/playtesting.md](quality/playtesting.md)).
4. First playable, then vertical slice only when a funding or derisking gate demands it — the slice is a pipeline and funding proof, not an automatic best practice ([operations/scoping-and-production.md](operations/scoping-and-production.md)).
5. Production: content is built against proven systems, gated by [checklists/design-review.md](checklists/design-review.md); irreversible decisions get DDRs ([decisions/design-decision-records.md](decisions/design-decision-records.md)).
6. Launch and live tuning: telemetry paired with qualitative playtesting; tunable values change via data, not client patches ([operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md)).

Steps 2–3 are a loop, not a phase; a design that skips them and goes straight to production is untested by definition.

## Proof Taxonomy

Design claims are proven with play evidence, not argument. Each proof type answers a different question; using the wrong one is the design equivalent of mocking the boundary under test.

| Proof Type | Default Stage | What It Proves |
|---|---|---|
| paper prototype | concept | loop viability and rule clarity, at near-zero cost |
| digital prototype | pre-production | real-time feel and control — things paper cannot carry |
| Kleenex test (fresh tester, used once) | onboarding and tutorial work | first-contact comprehension; a used tester can never be first-time again |
| structured playtest | every iteration | observed behavior against written experience goals |
| spreadsheet or Machinations simulation | balance and economy work | resource flows and cost curves before content is built |
| vertical slice | funding gate | all major systems integrating at intended final quality |
| telemetry and A/B cohort | live | what players do at scale — never why they do it |

The Kleenex-tester rule is Will Wright's (`masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`); economy simulation before content is the Machinations position (Adams & Dormans, *Game Mechanics: Advanced Game Design*, `oreilly.com/library/view/game-mechanics-advanced/9780132946728/ch05.html`).

## Design Contracts Worth Remembering

- Every mechanic must trace to a pillar, or be cut.
- Every loop must name the skill it teaches and the reward that closes it — fun is learning patterns, and a loop with nothing left to teach goes boring (Koster, *A Theory of Fun*).
- Every resource must have a written faucet and sink; faucets that outrun sinks are an inflation incident waiting for scale.
- Every new mechanic must be introduced by playing, not reading — teach, develop, twist, prove.
- Every juice effect must preserve game-state readability; juice amplifies a good core, it does not substitute for one.
- Every playtest session must have its question written before the session starts.
- Every tuning value shipped today becomes part of tomorrow's live-tuning surface; keep it in data.

## Contract Surfaces

- The pillars doc is the veto contract: any feature that serves no pillar is cut by default, and changing a pillar requires a DDR.
- One-page GDDs are the per-system contract; when the build and the page disagree, update the page or mark it superseded in the same change — a stale page is worse than none.
- The balance spreadsheet spec ([templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md)) is the source of truth for numbers; values pasted into code without a spreadsheet row are drift.
- Playtest reports are the evidence trail; a design claim with no report behind it is an opinion.
- DDRs record decisions that are expensive to reverse; the frameworks catalog ([decisions/frameworks-and-models.md](decisions/frameworks-and-models.md)) records which model vocabulary the team has standardized on.

## Common Failure Modes

| Symptom | Likely Cause | First Fix |
|---|---|---|
| feature list grows but the game is not fun | no proven core loop; features substituting for a loop | cut to the loop, prototype it, playtest it — [recipes/design-a-core-loop.md](recipes/design-a-core-loop.md) |
| playtests keep returning "it was fine" | interrogation instead of observation; friends and teammates as testers | scripted sessions, fresh testers, watch behavior — [recipes/run-a-playtest.md](recipes/run-a-playtest.md) |
| one strategy dominates everything | a strictly dominant option; benefits scaling faster than costs | cost-curve audit plus intransitive counters — [recipes/tune-a-difficulty-curve.md](recipes/tune-a-difficulty-curve.md) |
| currency becomes worthless in live | faucets outrunning sinks | model the flows, add or widen sinks — [recipes/balance-an-economy.md](recipes/balance-an-economy.md) |
| players quit in the first session | text-heavy, front-loaded teaching | blend the tutorial into play, spread mechanic introduction out — [disciplines/ux-and-onboarding.md](disciplines/ux-and-onboarding.md) |
| more polish, less clarity | juice masking a weak or unreadable core | readability audit before more effects — [foundations/game-feel.md](foundations/game-feel.md) |
| the GDD exists and nobody reads it | monolithic novel-style documentation | one dense visual page per system — [recipes/write-a-one-page-gdd.md](recipes/write-a-one-page-gdd.md) |
| telemetry says what changed, team guesses why | quantitative data with no qualitative pairing | run playtests against the same question the metrics raised |

## Primary Sources Behind These Defaults

- MDA framework: `users.cs.northwestern.edu/~hunicke/MDA.pdf`
- loops, skill atoms, loops vs arcs, value-chain economies: `lostgarden.com/2021/03/13/the-chemistry-of-game-design-2/`, `lostgarden.com/2012/04/30/loops-and-arcs/`, `lostgarden.com/2021/12/12/value-chains/`
- fun as pattern learning: Koster, *A Theory of Fun for Game Design* — `theoryoffun.com/press.shtml`
- player motivation (SDT in games): Ryan, Rigby & Przybylski (2006) — `link.springer.com/article/10.1007/s11031-006-9051-8`; Quantic Foundry model — `quanticfoundry.com/2015/12/15/handy-reference/`
- flow and difficulty: Csikszentmihalyi, *Flow* (1990); Chen, "Flow in Games" — `jenovachen.com/flowingames/Flow_in_games_final.pdf`
- game feel and juice: Swink, *Game Feel* — `routledge.com/Game-Feel-A-Game-Designers-Guide-to-Virtual-Sensation/Swink/p/book/9780123743282`; Nijman, "The Art of Screenshake" — `youtube.com/watch?v=AJdEqssNZ-U`; Jonasson & Purho, "Juice It or Lose It" — `gdcvault.com/play/1016487/Juice-It-or-Lose`
- balance: Schreiber, Game Balance Concepts — `gamebalanceconcepts.wordpress.com/2010/07/21/level-3-transitive-mechanics-and-cost-curves/`; Sirlin, viable options — `sirlin.net/articles/balancing-multiplayer-games-part-2-viable-options`
- economies and Machinations: `oreilly.com/library/view/game-mechanics-advanced/9780132946728/ch05.html`
- level design: Taylor, "Ten Principles of Good Level Design" — `gamedeveloper.com/design/ten-principles-of-good-level-design-part-1-`; kishōtenketsu — `gamedeveloper.com/design/the-secret-to-i-mario-i-level-design`
- playcentric process and playtesting: Fullerton, *Game Design Workshop*, 5th ed. — `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`
- onboarding: Fan, "How I Got My Mom to Play Through Plants vs. Zombies" — `gdcvault.com/play/1015541/How-I-Got-My-Mom`
- game UX: Hodent, *The Gamer's Brain* — `thegamersbrain.com`
- accessibility: `gameaccessibilityguidelines.com/full-list/`; AbleGamers APX — `ablegamers.org/apx/`
- documentation: Librande, "One-Page Designs" — `gdcvault.com/play/1012356/One-Page`
- scoping and slices: Ismail, "Prototypes & Vertical Slice" — `ltpf.ramiismail.com/prototypes-and-vertical-slice/`
- live analytics: Seif El-Nasr, Drachen & Canossa (eds.), *Game Analytics* — `link.springer.com/book/10.1007/978-1-4471-4769-5`

## Related Docs

- Fast path and change routing: [AGENTS.md](AGENTS.md)
- Pillars and vision: [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md)
- Core loops: [foundations/core-loops.md](foundations/core-loops.md)
- Documentation practice: [foundations/design-documentation.md](foundations/design-documentation.md)
- Prototyping and playtesting: [quality/prototyping.md](quality/prototyping.md), [quality/playtesting.md](quality/playtesting.md)
- Frameworks the handbook has standardized on: [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md)
