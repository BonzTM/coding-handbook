# Frameworks And Models

Rules for selecting the default analytical framework or model per design concern. The selection and its defense live only in this table — topic docs apply the chosen framework, cite it, and link here for the choice.

## Default Approach

Default to the framework with the strongest empirical or practitioner grounding for each concern, treat it as shared vocabulary rather than dogma, and route every departure through a DDR (see [design-decision-records.md](design-decision-records.md)). A model earns its place by making a design decision checkable — if adopting it changes no artifact, diagram, or playtest question, do not adopt it.

### Approval Questions

Before adopting a framework or model this table does not already select, answer all of these in writing:

1. What decision does the current default fail to support, concretely, in this project?
2. What is the model's evidentiary basis — peer-reviewed study, large-sample data, or a single practitioner's talk?
3. Is the model contested, and if so, which critique applies to this use (see [Contested Models Policy](#contested-models-policy))?
4. What artifact will the model produce (diagram, curve, spreadsheet, playtest script), and which topic doc owns that artifact?

## Default Choices By Concern

| Concern | Default | Acceptable escalation | Avoid by default |
|---|---|---|---|
| analysis vocabulary | MDA — mechanics, dynamics, aesthetics; designers build bottom-up, players experience top-down (Hunicke, LeBlanc & Zubek 2004, `users.cs.northwestern.edu/~hunicke/MDA.pdf`); applied in [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md) | DDE (Walk, Görlich & Barrett 2017, `link.springer.com/chapter/10.1007/978-3-319-53088-8_3`) via DDR when experience-first or narrative-heavy work exposes MDA's known gaps | inventing per-project vocabulary; presenting any MDA successor as settled canon |
| loop analysis | core-loop diagram plus Daniel Cook's skill atoms and loops-vs-arcs split (`lostgarden.com/2012/04/30/loops-and-arcs/`); applied in [../foundations/core-loops.md](../foundations/core-loops.md) | full skill-atom chain mapping when playtests show players failing to master a specific skill | shipping arcs priced as loops; loop diagrams so detailed they become specs |
| motivation model | self-determination theory — autonomy, competence, relatedness predict enjoyment and continued play (Ryan, Rigby & Przybylski 2006, `link.springer.com/article/10.1007/s11031-006-9051-8`); applied in [../foundations/player-psychology.md](../foundations/player-psychology.md) | Quantic Foundry's 12-motivation model (`quanticfoundry.com/2015/12/15/handy-reference/`) for audience segmentation, since it treats motivations as continuous dimensions over 400,000+ gamers | Bartle types as a design driver — historically foundational, but derived from MUD observation and cautioned against over-generalizing by Bartle himself (`mud.co.uk/richard/hcds.htm`) |
| difficulty policy | flow-channel targeting via player-visible difficulty options and Chen-style implicit self-selection (`jenovachen.com/flowingames/Flow_in_games_final.pdf`); applied in [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) | hidden dynamic difficulty adjustment via DDR only, with genre justification and an exploit review | hidden DDA in competitive modes; a single fixed curve with no self-selection path |
| balance method | Schreiber cost curves for transitive systems, intransitive (rock-paper-scissors) relationships for option viability (`gamebalanceconcepts.wordpress.com/2010/07/21/level-3-transitive-mechanics-and-cost-curves/`), Sirlin's viable-options and fairness tests for competitive play (`sirlin.net/articles/balancing-multiplayer-games-part-2-viable-options`); applied in [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) | telemetry-driven tuning per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md) once live data exists | gut-feel tuning with no cost model; shipping a strictly dominant option knowingly |
| economy modeling | faucet-and-sink flow model diagrammed in Machinations notation — pools, sources, drains, converters, gates (Adams & Dormans, `oreilly.com/library/view/game-mechanics-advanced/9780132946728/ch05.html`), with Cook's value chains for constructing the flows (`lostgarden.com/2021/12/12/value-chains/`); applied in [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md) | the machinations.io simulation tool via DDR when a live or multi-currency economy outgrows static diagrams | adding a currency or source without a sink audit; balancing an economy in code before it is modeled |
| pacing model | Schell interest curves — hook, alternating peaks and valleys, escalation to climax, fractal across scales (*The Art of Game Design*, `routledge.com/The-Art-of-Game-Design-A-Book-of-Lenses-Third-Edition/Schell/p/book/9781138632059`) | kishōtenketsu four-step structure for mechanic-driven level pacing, owned by [../disciplines/level-design.md](../disciplines/level-design.md) | pacing by content volume instead of interest shape |
| game-feel vocabulary | Swink's definition — real-time control of virtual objects, interactions emphasized by polish (*Game Feel*, `routledge.com/Game-Feel-A-Game-Designers-Guide-to-Virtual-Sensation/Swink/p/book/9780123743282`); juice techniques per Nijman and Jonasson & Purho, applied in [../foundations/game-feel.md](../foundations/game-feel.md) | — | juice as a substitute for a working core; effects that cost game-state readability |
| playtest method | Fullerton's playcentric protocol — experience goals first, then prototype, playtest, revise (*Game Design Workshop* 5th ed., `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`), with Kleenex (single-use fresh) testers for onboarding; owned by [../quality/playtesting.md](../quality/playtesting.md) | moderated UX-lab sessions per [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md) when budget allows | demo-and-ask sessions; reusing testers for first-impression questions |
| design interrogation | Schell's lenses as hypothesis generators, checked by playtests — lenses ask, playtests answer | — | treating lens answers as evidence without a playtest |
| documentation format | Librande one-page designs — compress each system to one dense visual page (`gdcvault.com/play/1012356/One-Page`); owned by [../foundations/design-documentation.md](../foundations/design-documentation.md) | heavier milestone documentation when a publisher contract requires it, recorded in a DDR | monolithic up-front GDDs; zero documentation |
| accessibility baseline | Game Accessibility Guidelines basic tier, minimum bar of remapping, text size, colorblindness, subtitle presentation (`gameaccessibilityguidelines.com/full-list/`); owned by [../disciplines/accessibility.md](../disciplines/accessibility.md) | AbleGamers APX patterns (`ablegamers.org/apx/`) or Xbox Accessibility Guidelines for platform or certification needs | accessibility as a post-hoc patch; skipping the four most-complained-about issues |
| telemetry posture | instrument the design questions, not everything; pair quantitative telemetry with qualitative playtesting (*Game Analytics*, Seif El-Nasr et al., `link.springer.com/book/10.1007/978-1-4471-4769-5`); owned by [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md) | A/B testing on segments before global rollout once live | collect-everything pipelines with no design question attached; metrics-first monetization design |

## Contested Models Policy

Several defaults above are contested in the field. The policy: a contested model may be adopted as shared vocabulary, but the doc applying it must name the critique, and taking a position beyond the default stance requires a DDR. The handbook's recorded stances:

- **MDA vs successors**: MDA is "the most widely accepted and practically employed approach" yet criticized for neglecting narrative and audiovisual design (per the DDE paper itself, `link.springer.com/chapter/10.1007/978-3-319-53088-8_3`). Teach MDA as vocabulary; note the critiques; treat no successor as settled.
- **Bartle taxonomy**: historical foundation only. It was derived from MUD observation, not psychometrics; Quantic Foundry's continuous dimensions are the empirical replacement for segmentation.
- **Hidden DDA**: opinion-divided — it can feel patronizing or be exploited, and competitive designers widely reject it. Player-visible options and implicit self-selection are the default; hidden DDA is DDR-gated.
- **Juice**: amplifies a good core, does not substitute for one; practitioners have pushed back that juice can mask weak mechanics and hurt readability (`gamedeveloper.com/design/video-indies-resist-the-urge-to-juice-it-or-lose-it-`). Every effect must preserve game-state readability.
- **Vertical slices**: a funding and pipeline-derisking tool, not an automatic best practice — slice content is often rebuilt and early polish can precede proven systems. Stance owned by [../operations/scoping-and-production.md](../operations/scoping-and-production.md).
- **Documentation weight**: "no GDD" is as wrong as a thousand-page GDD; the mainstream is living one-pagers and prototypes-as-spec, with heavier docs only under contract. Stance owned by [../foundations/design-documentation.md](../foundations/design-documentation.md).
- **Telemetry-first design**: telemetry says what players do, not why; metrics-driven monetization is ethically contested. Quantitative data never overrides qualitative findings without a DDR explaining the conflict.

## Common Mistakes And Forbidden Patterns

- No choosing, re-arguing, or taking a stance on a framework in a topic doc — topic docs apply the selected default (naming it and citing its primary source per [../CONTRIBUTING.md](../CONTRIBUTING.md)) and link here for why it was chosen.
- No presenting a contested model as settled canon, or a successor framework as a replacement the field has not accepted.
- No adopting a framework because it appeared in a talk, without answering the Approval Questions in writing.
- No discrete player typing (Bartle or derivative quizzes) as a targeting or design-driver mechanism.
- No hidden difficulty adjustment landing without a DDR and an exploit review.
- No economy change shipped without the faucet/sink model updated first.
- No framework adopted that produces no artifact a review can inspect.
- No exception to a default in this table without a DDR recorded.

## Verification And Proof

A framework choice is proven, not asserted. Before a departure from this table lands, demonstrate all of:

- The Approval Questions are answered in writing, in the DDR — not left implicit.
- The critique of any contested model is named in the DDR and addressed for this use.
- The artifact the model produces exists and is linked from the owning topic doc.
- A search of `foundations/`, `disciplines/`, `quality/`, and `operations/` shows no topic doc selecting a different framework, re-arguing a selection, or contradicting a recorded stance — the choice and its defense live here and in the DDR, and those docs link here for them.

### Decision Record

When a project departs from a default, the DDR (see [design-decision-records.md](design-decision-records.md), template in [../templates/ddr-template.md](../templates/ddr-template.md)) must write down:

- the model chosen and why the default was insufficient
- the evidentiary basis and the contested-model critique, if one applies
- which topic area applies the model and what artifact it produces
- what playtest or telemetry result would trigger re-evaluation
