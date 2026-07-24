# Game Design Handbook

This handbook is the default design contract for new game projects. It is not a game design tutorial. It exists to make projects converge on the same decision order, shared vocabulary, documentation shape, and proof of design correctness: what to decide, in what order, and how to prove a design works before production is built on top of it. Monetization and business-model design is currently out of scope beyond the telemetry and economy touchpoints in [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md) and [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md); if adopted, it would live as a new `disciplines/` doc with its own routing row. Reference exemplar projects are a planned later phase.

## Start Here

- Humans: read this file, then follow the reading path for your project shape.
- Agents: read [AGENTS.md](AGENTS.md) first (it includes the change-routing table), then the relevant topical docs and recipes.
- Default assumptions unless a project says otherwise:
  - two to three named design pillars before any mechanics work
  - a diagrammed core loop before any content is built on it
  - one-page docs per system, not a monolithic GDD
  - prototype and playtest before spec; no design claim survives without a playtest
  - MDA as the shared analysis vocabulary; alternatives route through [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md)
  - design decisions that bind future work recorded as DDRs per [decisions/design-decision-records.md](decisions/design-decision-records.md)

## Reading Paths

| If you are designing... | Read in this order |
|---|---|
| A new game concept | [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md) -> [foundations/core-loops.md](foundations/core-loops.md) -> [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md) -> [foundations/design-documentation.md](foundations/design-documentation.md) -> [checklists/concept-intake.md](checklists/concept-intake.md) -> [recipes/write-a-one-page-gdd.md](recipes/write-a-one-page-gdd.md) |
| The core loop and moment-to-moment play | [foundations/core-loops.md](foundations/core-loops.md) -> [foundations/player-psychology.md](foundations/player-psychology.md) -> [foundations/game-feel.md](foundations/game-feel.md) -> [quality/prototyping.md](quality/prototyping.md) -> [quality/playtesting.md](quality/playtesting.md) -> [recipes/design-a-core-loop.md](recipes/design-a-core-loop.md) |
| Levels and content | [foundations/core-loops.md](foundations/core-loops.md) -> [disciplines/level-design.md](disciplines/level-design.md) -> [disciplines/narrative-integration.md](disciplines/narrative-integration.md) -> [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md) -> [recipes/design-a-level.md](recipes/design-a-level.md) -> [recipes/tune-a-difficulty-curve.md](recipes/tune-a-difficulty-curve.md) |
| A systems or economy game | [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md) -> [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md) -> [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md) -> [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md) -> [recipes/balance-an-economy.md](recipes/balance-an-economy.md) |
| Onboarding and the first hour | [disciplines/ux-and-onboarding.md](disciplines/ux-and-onboarding.md) -> [foundations/player-psychology.md](foundations/player-psychology.md) -> [disciplines/accessibility.md](disciplines/accessibility.md) -> [checklists/playtest-readiness.md](checklists/playtest-readiness.md) -> [recipes/run-a-playtest.md](recipes/run-a-playtest.md) |
| A live game being tuned post-launch | [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md) -> [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md) -> [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md) -> [quality/playtesting.md](quality/playtesting.md) |
| A pitch or production plan | [operations/scoping-and-production.md](operations/scoping-and-production.md) -> [foundations/design-documentation.md](foundations/design-documentation.md) -> [templates/one-page-gdd.md](templates/one-page-gdd.md) -> [checklists/design-review.md](checklists/design-review.md) |

Every shape also adopts [quality/playtesting.md](quality/playtesting.md) as its proof discipline, records binding decisions per [decisions/design-decision-records.md](decisions/design-decision-records.md), and runs design reviews against [checklists/design-review.md](checklists/design-review.md). Shipped projects additionally follow [disciplines/accessibility.md](disciplines/accessibility.md) and [operations/scoping-and-production.md](operations/scoping-and-production.md).

## Non-Negotiables

- State the design pillars before designing mechanics; a change that violates a pillar requires a DDR, not a quiet exception. Owned by [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md).
- Diagram the core loop before building it, and know which parts of the game are loops (repeatable, mastery-driven) versus arcs (consumed once) — the distinction decides where production money goes (Daniel Cook, `lostgarden.com/2012/04/30/loops-and-arcs/`).
- Prototype before spec. Set experience goals first, then prototype -> playtest -> revise against those goals; paper and greybox prototypes are the cheapest possible test of a design (Fullerton, *Game Design Workshop*, the playcentric process).
- No design claim survives without a playtest. Observe players; do not interrogate them mid-session. First-time testers are spent after one session and never reused for onboarding tests (Kleenex testing, popularized by Will Wright).
- Document systems as one dense visual page each, not a monolithic GDD; documentation exists to communicate, and "most people only read the first page anyway" (Stone Librande, "One-Page Designs", GDC 2010).
- Difficulty options are player-visible by default. Hidden dynamic difficulty adjustment is contested practice and requires a DDR. Owned by [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md).
- Address accessibility during design, not as a post-hoc patch. The minimum bar is the four most-complained-about issues: input remapping, text size, colorblind-safe presentation, subtitle quality (`gameaccessibilityguidelines.com`).
- Every tunable number lives in the balance sheet with its derivation; no orphan constants scattered across specs. Owned by [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md).
- Telemetry instruments a named design question; do not collect everything and hope. Quantitative data says what players do, never why — pair it with qualitative playtesting.
- Juice amplifies a proven core; it never substitutes for one, and every effect must preserve game-state readability. Owned by [foundations/game-feel.md](foundations/game-feel.md).

## Default Stack

| Concern | Default | Reach for something else when |
|---|---|---|
| Analysis vocabulary | MDA (mechanics -> dynamics -> aesthetics; designers build bottom-up, players experience top-down) | experience-first work justifies DDE framing via [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md) — no successor has displaced MDA as shared vocabulary |
| Motivation model | self-determination theory (autonomy, competence, relatedness) for design; Quantic Foundry's 12 continuous motivations for audience profiling | Bartle types only as historical vocabulary, never as a segmentation tool |
| Loop design | core-loop diagram plus skill-atom decomposition to find where players fail to learn | a pure arc-driven (narrative) production, where content pacing dominates |
| System documentation | one-page design per system, kept living | a publisher or platform milestone contractually requires heavier documentation |
| Economy modeling | faucet/sink flow map plus the balance spreadsheet | flow complexity justifies executable simulation (Machinations-style diagrams) via [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md) |
| Balance method | cost curves for transitive objects; intransitive (counter) relationships to keep multiple options viable (Schreiber, Game Balance Concepts) | single-player games where fairness between players is not a concern |
| Difficulty | flow-channel targeting with player-visible options | implicit player-driven self-selection of challenge (Chen, "Flow in Games") |
| Level structure | four-step kishōtenketsu pacing: introduce, develop, twist, prove mastery | levels not driven by a teachable mechanic |
| Tutorials | in-game contextual teaching with minimal words (Fan, "How I Got My Mom to Play Through Plants vs. Zombies", GDC 2012) | never a front-loaded text manual |
| Playtesting | scripted sessions from [templates/playtest-script.md](templates/playtest-script.md), reported via [templates/playtest-report.md](templates/playtest-report.md) | almost never; unscripted sessions are exploration, not proof |
| Accessibility bar | Game Accessibility Guidelines basic tier | audience reach justifies intermediate/advanced tiers or AbleGamers APX patterns |
| Scoping ladder | prototype -> first playable -> vertical slice -> production | slice-content waste justifies horizontal-first prototyping, documented in [operations/scoping-and-production.md](operations/scoping-and-production.md) |
| Live tuning | remote-config parameters, segment-scoped A/B tests, cohort retention measures | the project has no live service |

## Handbook Map

- [AGENTS.md](AGENTS.md) - fast-path contract and change routing for autonomous agents and reviewers
- [maintainer-reference.md](maintainer-reference.md) - architecture, rationale, and deeper guidance
- `foundations/` - [design pillars and vision](foundations/design-pillars-and-vision.md), [core loops](foundations/core-loops.md), [mechanics and systems](foundations/mechanics-and-systems.md), [player psychology](foundations/player-psychology.md), [game feel](foundations/game-feel.md), and [design documentation](foundations/design-documentation.md)
- `disciplines/` - [level design](disciplines/level-design.md), [narrative integration](disciplines/narrative-integration.md), [economy and progression](disciplines/economy-and-progression.md), [difficulty and balance](disciplines/difficulty-and-balance.md), [UX and onboarding](disciplines/ux-and-onboarding.md), [accessibility](disciplines/accessibility.md), and [multiplayer and social](disciplines/multiplayer-and-social.md)
- `quality/` - [prototyping](quality/prototyping.md) and [playtesting](quality/playtesting.md), the two proof disciplines every design claim passes through
- `operations/` - [scoping and production](operations/scoping-and-production.md) and [live tuning and telemetry](operations/live-tuning-and-telemetry.md)
- `decisions/` ([README.md](decisions/README.md)) - [design decision records](decisions/design-decision-records.md) (DDRs) plus [frameworks and models](decisions/frameworks-and-models.md) selection rules
- `checklists/` ([README.md](checklists/README.md)) and `recipes/` ([README.md](recipes/README.md)) - executable intake, review, readiness, and implementation guidance
- `templates/` ([README.md](templates/README.md)) - committed fill-in skeletons: [one-page GDD](templates/one-page-gdd.md), [playtest script](templates/playtest-script.md), [playtest report](templates/playtest-report.md), [balance spreadsheet spec](templates/balance-spreadsheet-spec.md), and [DDR template](templates/ddr-template.md)
- Team process (human-facing; not read during design builds): [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership transfer, [glossary.md](glossary.md) as a term lookup, and [CONTRIBUTING.md](CONTRIBUTING.md) for changing the handbook itself
- `reference/` exemplar projects composing these patterns end to end are a planned later phase; until they land, the [templates/](templates/) skeletons are the copy-paste starting point

## What This Handbook Optimizes For

- design intent that is still legible six months later
- loops proven by play before content is built on top of them
- decisions recorded once as DDRs, not re-argued every meeting
- evidence from real players over designer intuition
- scope that survives contact with production

## Where To Go Next

- New concept intake: [checklists/concept-intake.md](checklists/concept-intake.md)
- Active agent work: [AGENTS.md](AGENTS.md)
- Routing a change quickly: [AGENTS.md](AGENTS.md) (## Change Routing)
- Writing the first design doc: [recipes/write-a-one-page-gdd.md](recipes/write-a-one-page-gdd.md)
- Designing or fixing a core loop: [recipes/design-a-core-loop.md](recipes/design-a-core-loop.md)
- Running a playtest: [checklists/playtest-readiness.md](checklists/playtest-readiness.md), then [recipes/run-a-playtest.md](recipes/run-a-playtest.md)
- Tuning difficulty or balance: [recipes/tune-a-difficulty-curve.md](recipes/tune-a-difficulty-curve.md), [recipes/balance-an-economy.md](recipes/balance-an-economy.md)
- Building a level: [recipes/design-a-level.md](recipes/design-a-level.md)
- Choosing a framework or model: [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md)
- Recording a design decision: [decisions/design-decision-records.md](decisions/design-decision-records.md)
- Copy-paste skeletons: [templates/README.md](templates/README.md)
- Taking over or handing off a project: [onboarding-and-handoff.md](onboarding-and-handoff.md)
- Looking up a handbook term: [glossary.md](glossary.md)
- Changing the handbook itself: [CONTRIBUTING.md](CONTRIBUTING.md)
