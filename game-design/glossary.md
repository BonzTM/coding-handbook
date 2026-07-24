# Glossary

> **Lookup aid, not required reading.** Consult a single entry when a handbook term is unclear; nothing here needs to be read front-to-back to design.

Canonical vocabulary for this handbook. Every term below has exactly one meaning across all projects; use the word for nothing else, and define it nowhere else. Each entry is one or two sentences plus the doc that owns the full rule — read that doc before relying on the term. Start from [README.md](README.md) for how these pieces fit together.

Terms are alphabetical.

## Arc
Content that is consumed once and does not repeat — story beats, set pieces, scripted moments — as opposed to a loop, which is replayed toward mastery (Daniel Cook, "Loops and Arcs"). The distinction decides what a production is actually spending money on. Owned by [foundations/core-loops.md](foundations/core-loops.md).

## Core loop
The repeatable sequence of player actions that defines the primary flow of the game, diagrammed as a deliberately zoomed-out view of how actions feed into each other — not a detailed spec. Loops are fractal: moment-to-moment, session, and meta loops nest inside each other. Owned by [foundations/core-loops.md](foundations/core-loops.md); built via [recipes/design-a-core-loop.md](recipes/design-a-core-loop.md).

## Cost curve
The plotted relationship of an object's total costs to its total benefits, used to balance transitive objects against each other (Ian Schreiber, Game Balance Concepts). An object above the curve is overpowered; below it, dead content. Owned by [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md).

## DDR (Design Decision Record)
A short, immutable document stating one design decision, the context that forced it, and the consequences accepted; once accepted it is frozen and only superseded, never edited. The design-side counterpart of an ADR. Owned by [decisions/design-decision-records.md](decisions/design-decision-records.md); skeleton in [templates/ddr-template.md](templates/ddr-template.md).

## Design pillar
One of a small fixed set of statements naming what the game is optimizing for, used as the veto test for every feature: a proposal that serves no pillar is cut. Owned by [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md).

## Dynamic difficulty adjustment (DDA)
Runtime tuning of challenge in response to measured player performance. Hidden DDA is contested — it can feel patronizing or exploitable — so this handbook defaults to player-visible options and Chen-style implicit choice. Owned by [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md).

## Faucet and sink
The two halves of an internal economy: faucets inject resources (rewards, loot, harvesting) and sinks remove them (fees, repair, decay, consumables). When faucets outrun sinks the currency inflates toward worthless — the founding cautionary case is Ultima Online. Owned by [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md); balanced via [recipes/balance-an-economy.md](recipes/balance-an-economy.md).

## First playable
The first build in which the core loop can actually be played end to end, however ugly; it sits between prototype and vertical slice on the milestone ladder. Owned by [operations/scoping-and-production.md](operations/scoping-and-production.md).

## Flow channel
The band between anxiety (challenge exceeds skill) and boredom (skill exceeds challenge) where engagement lives, from Csikszentmihalyi's flow model as applied to games by Jenova Chen's "Flow in Games". Difficulty curves exist to keep the player inside it. Owned by [foundations/player-psychology.md](foundations/player-psychology.md); applied in [recipes/tune-a-difficulty-curve.md](recipes/tune-a-difficulty-curve.md).

## Game feel
"Real-time control of virtual objects in a simulated space, with interactions emphasized by polish" (Steve Swink, *Game Feel*). A property of input response, simulation, and polish together — not a synonym for juice alone. Owned by [foundations/game-feel.md](foundations/game-feel.md).

## Interest curve
Jesse Schell's pacing tool from *The Art of Game Design*: plot expected engagement over time — hook, alternating peaks and rest valleys, escalation to a climax — and apply it fractally to moments, levels, and the whole game. Owned by [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md).

## Intransitive mechanic
A rock-paper-scissors relationship in which no option dominates; the standard tool for keeping multiple options viable without exact numeric parity (Schreiber; Sirlin's viable-options rule). Owned by [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md).

## Juice
Maximal feedback for minimal input — tweens, particles, screen shake, hitstop, sound (Jonasson & Purho, "Juice It or Lose It"; Nijman, "The Art of Screenshake"). Juice amplifies a good core; it never substitutes for one, and every effect must preserve game-state readability. Owned by [foundations/game-feel.md](foundations/game-feel.md).

## Kishōtenketsu
The Nintendo four-step level structure (Koichi Hayashida, Super Mario 3D World): introduce a mechanic, develop it, twist it, then let the player demonstrate mastery. The default pacing template for mechanic-driven levels. Owned by [disciplines/level-design.md](disciplines/level-design.md); applied in [recipes/design-a-level.md](recipes/design-a-level.md).

## Kleenex test
A playtest using a fresh tester exactly once (term popularized by Will Wright): someone who has never seen the game reveals where it confuses, and once used can never be a first-time tester again. Mandatory for onboarding and tutorial validation. Owned by [quality/playtesting.md](quality/playtesting.md).

## Lens
One of Schell's named sets of questions for interrogating a design from a specific perspective. Lenses generate hypotheses; playtests check them — a lens answer is never accepted as proof on its own. Owned by [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md).

## Machinations
Joris Dormans's visual diagram language — pools, sources, drains, converters, gates — for modeling and simulating an internal economy before writing code (Adams & Dormans, *Game Mechanics: Advanced Game Design*). Owned by [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md).

## MDA
Mechanics → Dynamics → Aesthetics (Hunicke, LeBlanc & Zubek, 2004): designers build bottom-up from mechanics, players experience top-down from aesthetics, and design lives in the gap. Shared vocabulary in this handbook, taught with its known critiques, not as settled law. Owned by [decisions/frameworks-and-models.md](decisions/frameworks-and-models.md).

## Mechanic
A designer-authored rule of the game — the M in MDA — as distinct from a dynamic, the runtime behavior that emerges when mechanics meet players. Mechanics are what you can edit; dynamics are what you can only observe and test. Owned by [foundations/mechanics-and-systems.md](foundations/mechanics-and-systems.md).

## One-page design
A single dense, visual page that compresses one system to what the team can act on (Stone Librande, "One-Page Designs"): documentation exists to communicate efficiently, and "most people only read the first page anyway." The house GDD format. Owned by [foundations/design-documentation.md](foundations/design-documentation.md); written via [recipes/write-a-one-page-gdd.md](recipes/write-a-one-page-gdd.md) from [templates/one-page-gdd.md](templates/one-page-gdd.md).

## Paper prototype
A physical, code-free prototype whose rules can change mid-test — the fastest, cheapest way to answer a design question, and the heart of Fullerton's playcentric process. Owned by [quality/prototyping.md](quality/prototyping.md).

## Playcentric process
Fullerton's iteration contract from *Game Design Workshop*: set experience goals first, then prototype → playtest → revise continuously against those goals, never against taste alone. Owned by [quality/playtesting.md](quality/playtesting.md); prototyping side in [quality/prototyping.md](quality/prototyping.md).

## Skill atom
Daniel Cook's atomic feedback loop ("The Chemistry of Game Design"): the smallest unit in which a player performs an action, gets feedback, and gains a skill. Mapping skill atoms shows where players get lost and fail to master the game. Owned by [foundations/core-loops.md](foundations/core-loops.md).

## Vertical slice
A small, fully playable portion of the game showing all major systems working together at intended final quality — the artifact that proves the team *can* make the game, where a prototype proves it *should* (Rami Ismail). A funding and de-risking tool, not an automatic best practice. Owned by [operations/scoping-and-production.md](operations/scoping-and-production.md).

## Related

- [README.md](README.md) — how the pillars, loops, disciplines, and proof practices fit together.
- [AGENTS.md](AGENTS.md) — the fast-path contract that uses this vocabulary.
- [AGENTS.md](AGENTS.md) (## Change Routing) — which doc owns each change surface.
