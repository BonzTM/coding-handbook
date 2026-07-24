# AGENTS.md - Game Design Project Contract

This is the authoritative fast-path contract for autonomous agents doing game design work in a new project.
Read this file first: it carries the repo-wide design invariants, the change-routing table, and the verification bar. Use [maintainer-reference.md](maintainer-reference.md) when you need slower-path architecture and rationale.

## Purpose

- Use this file for repo-wide design invariants, change defaults, change-to-doc routing, and the verification bar.
- Use [maintainer-reference.md](maintainer-reference.md) for the handbook map, the design lifecycle model, the proof taxonomy, and troubleshooting.
- For the full catalogs, see the [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), [decisions/README.md](decisions/README.md), and [templates/README.md](templates/README.md) indexes.

## Source Of Truth

- This file is the fast path. More detailed docs refine it; they do not weaken it.
- Vision and scope authority lives in [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md).
- The loop vocabulary (core loop, skill atom, loops vs arcs) lives in [foundations/core-loops.md](foundations/core-loops.md); system and mechanic rules in [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md).
- Motivation and flow models live in [foundations/player-psychology.md](foundations/player-psychology.md); moment-to-moment control and polish rules in [foundations/game-feel.md](foundations/game-feel.md).
- Documentation shape (one-pagers, living docs) lives in [foundations/design-documentation.md](foundations/design-documentation.md).
- Discipline rules live in [disciplines/level-design.md](disciplines/level-design.md), [disciplines/narrative-integration.md](disciplines/narrative-integration.md), [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md), [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md), [disciplines/ux-and-onboarding.md](disciplines/ux-and-onboarding.md), [disciplines/accessibility.md](disciplines/accessibility.md), and [disciplines/multiplayer-and-social.md](disciplines/multiplayer-and-social.md).
- Proof methodology lives in [quality/prototyping.md](quality/prototyping.md) and [quality/playtesting.md](quality/playtesting.md).
- Production and live-game rules live in [operations/scoping-and-production.md](operations/scoping-and-production.md) and [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md).
- Framework selection and contested-model stances (MDA, SDT, Bartle, Quantic Foundry) live only in [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md); topic docs apply the selected framework in their own domain. Design decisions and their rationale live in [decisions/design-decision-records.md](decisions/design-decision-records.md).
- Team-process docs — [onboarding-and-handoff.md](onboarding-and-handoff.md), [CONTRIBUTING.md](CONTRIBUTING.md), and the [glossary.md](glossary.md) lookup aid — serve humans running the team and the handbook; they are not needed to design a feature.
- Copy-paste artifacts live in [templates/](templates/). Worked exemplar projects under `reference/` are a planned later phase; until they land, start new artifacts from the templates.

## Fast Path

1. Read this file and identify the project shape from [README.md](README.md). For a brand-new concept, run [checklists/concept-intake.md](checklists/concept-intake.md) first — it says which WHAT decisions to take from the pitch, which to ask about, and which defaults apply when the pitch is silent.
2. Route the change through the [Change Routing](#change-routing) table below; do not guess where a design decision belongs.
3. Read the relevant foundations or disciplines doc before changing design in a new area.
4. Design with the repo defaults unless the project has already documented an exception in a DDR.
5. Prove the change with the narrowest meaningful playtest or simulation first, then the repo-wide baseline.

## Repo-Wide Invariants

- **Pillars gate scope**: every feature traces to a named design pillar in [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md); a feature that serves no pillar is cut or the pillar set changes via DDR.
- **Loop before content**: the core loop is proven in a playable prototype before content is authored against it. Loops are repeatable mastery systems; arcs are consumed once and priced accordingly (Daniel Cook, "Loops and Arcs", `lostgarden.com/2012/04/30/loops-and-arcs/`).
- **Playtest evidence over intuition**: design claims graduate only on observed playtest behavior against stated experience goals — the playcentric process (Fullerton, *Game Design Workshop*, 5th ed.). Observation beats interrogation; record what players did, then what they said.
- **One page per system**: design documentation defaults to dense, visual one-pagers and living documents, not a monolithic up-front GDD (Stone Librande, "One-Page Designs", GDC 2010, `gdcvault.com/play/1012356/One-Page`).
- **Teach by playing**: onboarding introduces mechanics inside controlled play with minimal text, spread out over time (George Fan, "How I Got My Mom to Play Through Plants vs. Zombies", GDC 2012).
- **No strictly dominant option**: balance work preserves multiple viable options; a strictly dominant move destroys strategy (Sirlin, `sirlin.net/articles/balancing-multiplayer-games-part-2-viable-options`).
- **Every faucet has a sink**: no resource enters the economy without a matching drain and a modeled flow; unbalanced faucets are an inflation incident waiting for scale (Cook, "Value chains", `lostgarden.com/2021/12/12/value-chains/`).
- **Juice amplifies, never masks**: polish effects must preserve game-state readability and never substitute for a weak core (Swink, *Game Feel*; counterpoint coverage on Game Developer).
- **Accessibility at design time**: the four-item minimum bar — full remapping, text size, colorblind-safe signaling, subtitles — ships on every project and is dropped only via DDR; the basic tier of `gameaccessibilityguidelines.com` is adopted by default, with each skip named and reasoned in the adoption sheet ([disciplines/accessibility.md](disciplines/accessibility.md)). Decided during design, not patched after.
- **Telemetry answers a question**: instrument the design question, not everything; telemetry says what players do, playtests say why — always pair them ([operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md)).
- **Vocabulary is single-homed**: every term of art has one definition in [glossary.md](glossary.md) and one owning doc; framework selection and contested-model stances live only in [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md), while topic docs apply the selected framework.

## Change Routing

Use this when you know what kind of change you are making but not the file set. Start Here is what you read and touch first; Also Update is the sync surface the change normally drags along; Verify Or Confirm is the proof.

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| New concept or pitch intake | [checklists/concept-intake.md](checklists/concept-intake.md), [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md) | one-page GDD, experience goals, open questions or defaulted assumptions | intake resolved (answered, asked, or defaulted-and-disclosed) before design work; pillars named |
| Design pillars, vision statement, target experience | [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md), [decisions/design-decision-records.md](decisions/design-decision-records.md) | one-page GDD, affected discipline docs, scope plan | DDR recorded; every active feature still maps to a pillar |
| Core loop definition or change | [foundations/core-loops.md](foundations/core-loops.md), [recipes/design-a-core-loop.md](recipes/design-a-core-loop.md) | mechanics list, economy hooks, one-page GDD, prototype backlog | loop diagram current; loop playtested in prototype before content is built on it |
| New mechanic or system rule | [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md), [foundations/core-loops.md](foundations/core-loops.md) | loop diagram, balance sheet, glossary terms, one-page GDD | prototype answers the mechanic's question; [checklists/design-review.md](checklists/design-review.md) passes |
| Target audience, motivation profile, player-type assumptions | [foundations/player-psychology.md](foundations/player-psychology.md), [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md) | pillars, onboarding plan, playtest recruiting profile | motivation claims cite a model from the frameworks doc; playtest recruits match the profile |
| Controls, camera, feedback, juice, polish pass | [foundations/game-feel.md](foundations/game-feel.md) | readability review, accessibility toggles for intense effects, playtest script | fresh-tester session confirms control intent lands; game state stays readable with effects on |
| GDD, one-pager, spec, or design wiki change | [foundations/design-documentation.md](foundations/design-documentation.md), [recipes/write-a-one-page-gdd.md](recipes/write-a-one-page-gdd.md), [templates/one-page-gdd.md](templates/one-page-gdd.md) | glossary, linked DDRs, superseded doc versions | one page per system; the team can act on it without the author present |
| Level, encounter, or space design | [disciplines/level-design.md](disciplines/level-design.md), [recipes/design-a-level.md](recipes/design-a-level.md) | difficulty curve position, narrative beats, playtest plan | level teaches-develops-twists-proves its mechanic; testers navigate without prompting |
| Story beats, dialogue hooks, narrative-mechanic fit | [disciplines/narrative-integration.md](disciplines/narrative-integration.md) | level docs, pillars, glossary names | narrative serves the loop or is costed as an arc; beats survive a player who skips text |
| Currency, resource, reward, or progression change | [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md), [recipes/balance-an-economy.md](recipes/balance-an-economy.md), [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md) | difficulty curve, live tuning parameters, telemetry hooks | faucet/sink flows modeled and simulated before ship; no unmatched faucet |
| Difficulty curve, tuning pass, balance change | [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md), [recipes/tune-a-difficulty-curve.md](recipes/tune-a-difficulty-curve.md) | balance sheet, affected levels, playtest plan, telemetry | cost-curve check; no strictly dominant option; playtest confirms challenge tracks skill |
| Tutorial, onboarding, first-session UX | [disciplines/ux-and-onboarding.md](disciplines/ux-and-onboarding.md), [templates/playtest-script.md](templates/playtest-script.md) | game-feel signs and feedback, accessibility text sizing, level 1 design | fresh (never-before-seen) testers complete onboarding unaided; words minimized |
| Accessibility options or review | [disciplines/accessibility.md](disciplines/accessibility.md) | control remapping, subtitle and text presentation, effect toggles, design review checklist | minimum bar met or any drop recorded with a DDR; basic-tier skips named with a reason in the adoption sheet |
| PvP mode, matchmaking, social or spectator feature | [disciplines/multiplayer-and-social.md](disciplines/multiplayer-and-social.md) | balance sheet, per-side playtest plan, telemetry hooks, queue-policy DDR, [checklists/design-review.md](checklists/design-review.md) | win-rate parity at equal skill; matchmaking rating validated against realized outcomes; every player-to-player channel passes the four-axes review |
| New prototype (question to answer) | [quality/prototyping.md](quality/prototyping.md), [operations/scoping-and-production.md](operations/scoping-and-production.md) | prototype backlog, the question it answers, disposal plan | one question per prototype; answer recorded; throwaway build actually thrown away |
| Planning or running a playtest | [quality/playtesting.md](quality/playtesting.md), [recipes/run-a-playtest.md](recipes/run-a-playtest.md), [checklists/playtest-readiness.md](checklists/playtest-readiness.md), [templates/playtest-script.md](templates/playtest-script.md) | recruiting profile, session build, findings backlog | readiness checklist green before the session; [templates/playtest-report.md](templates/playtest-report.md) filed after |
| Scope, milestones, vertical slice, cut list | [operations/scoping-and-production.md](operations/scoping-and-production.md) | pillars, prototype backlog, cut-list DDRs | milestone ladder explicit (prototype → first playable → slice); cuts recorded, not silent |
| Live tuning, remote config values, A/B test | [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md) | balance sheet, economy model, event schema, rollback plan | change tested on a segment before global rollout; paired qualitative check scheduled |
| Choosing or citing a design framework or model | [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md), [decisions/README.md](decisions/README.md) | glossary, docs that cite the model | framework named from the decisions table, with its known critiques; not restated elsewhere |
| Non-obvious or hard-to-reverse design decision | [decisions/design-decision-records.md](decisions/design-decision-records.md), [templates/ddr-template.md](templates/ddr-template.md) | superseded DDRs, pillars if scope shifted, README links | DDR recorded with status, alternatives, and consequences before the change lands |
| Design review before build commitment | [checklists/design-review.md](checklists/design-review.md), [checklists/README.md](checklists/README.md) | open DDRs, one-page GDD, balance sheet | every box tied to evidence a reviewer can point at |
| Glossary term added or changed | [glossary.md](glossary.md) | the single owning doc for the term | exactly one definition; one named owner; other docs link, not restate |
| New recipe, checklist, template, or handbook doc | [CONTRIBUTING.md](CONTRIBUTING.md), [recipes/README.md](recipes/README.md), [templates/README.md](templates/README.md) | a routing row in this table, the owning index README, [maintainer-reference.md](maintainer-reference.md) map, glossary | house doc shape matched; all links resolve; no orphan doc |
| Design ownership transfer or onboarding | [onboarding-and-handoff.md](onboarding-and-handoff.md), [README.md](README.md) | pillars doc currency, open DDRs, playtest cadence, live tuning access | new owner answers the handoff questions unaided |

## High-Value Boundaries

- [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md) owns scope authority; no other doc may approve or cut features.
- foundations/ owns shared vocabulary and models; disciplines/ owns domain-specific rules and may not redefine foundation terms.
- quality/ owns proof methodology; a discipline doc states *what* to prove, [quality/playtesting.md](quality/playtesting.md) and [quality/prototyping.md](quality/prototyping.md) state *how*.
- operations/ owns production reality — milestones, cuts, live changes; design docs do not carry schedule promises.
- decisions/ owns named frameworks, models, and hard-to-reverse choices; topic docs link to it rather than naming frameworks inline.
- templates/ owns fill-in artifacts with explicit placeholders; prose docs do not duplicate template bodies.

## Proof Hints

- Loop and mechanic changes usually need a paper or greybox prototype session before any content is committed.
- Balance and economy changes usually need a spreadsheet or Machinations-style simulation plus a targeted playtest; simulation alone is not enough.
- Onboarding changes are not done until fresh testers — people who have never seen the game — get through unaided; a tester can only be fresh once.
- Feel and polish changes need a readability check: can a new observer state the game state with all effects firing.
- Live tuning changes are not done until the rollback path is written and a qualitative follow-up is scheduled.

## Working Norms

- Prefer small, testable design changes over broad redesigns.
- Do not introduce a new system because it feels richer; match the game's current loop unless the task is explicitly a pillar change.
- Do not bypass boundaries: discipline docs do not redefine vocabulary, prototypes do not become production content, telemetry does not replace playtests.
- When citing a framework, take it from [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md) with its critiques attached; do not present a contested model as settled.
- When a design claim changes, run or schedule the proving playtest before claiming success whenever practical.
- If verification fails — the playtest contradicts the design, the simulation diverges — fix the design or report it clearly. Do not claim the change is done.

## Baseline Verification

| Goal | Check | Expectation |
|---|---|---|
| pillar alignment | trace the change to a pillar in [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md) | a named pillar, or a cut, or a pillar DDR |
| loop integrity | current core-loop diagram vs the change | loop still closes; new actions feed an existing feedback loop |
| balance sanity | cost/benefit placement on the balance sheet per [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md) | no strictly dominant option; costs and benefits on curve |
| economy sanity | faucet/sink flow model per [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md) | every faucet matched by a sink; simulated before ship |
| onboarding proof | fresh-tester session per [quality/playtesting.md](quality/playtesting.md) | target players get through unaided; findings filed in a playtest report |
| accessibility floor | minimum bar and adoption sheet per [disciplines/accessibility.md](disciplines/accessibility.md) | remapping, text size, colorblind signaling, subtitles met or DDR'd; basic-tier skips reasoned in the adoption sheet |
| documentation sync | the system's one-pager and glossary terms | doc reflects the shipped design; no drift, no orphan terms |
| decision trail | [decisions/design-decision-records.md](decisions/design-decision-records.md) | hard-to-reverse choices have a DDR with alternatives and consequences |

Run the narrowest applicable checks first; the full bar applies before a milestone or handoff. There is no compiler for design — the gate is evidence, and [quality/playtesting.md](quality/playtesting.md) defines what counts.

## Slow Path Docs

- Handbook map and lifecycle: [maintainer-reference.md](maintainer-reference.md)
- Vision and vocabulary: [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md), [foundations/core-loops.md](foundations/core-loops.md), [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md)
- Proof and verification: [quality/prototyping.md](quality/prototyping.md), [quality/playtesting.md](quality/playtesting.md)
- Production reality: [operations/scoping-and-production.md](operations/scoping-and-production.md), [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md)
