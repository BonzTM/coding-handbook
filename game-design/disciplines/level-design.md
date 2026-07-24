# Level Design

Level design defaults for building spaces that teach and exercise mechanics: the principle baseline, the four-beat level structure, wordless guidance, and encounter-space rules.

## Default Approach

- Treat every level as a gameplay delivery system for mechanics, not as scenery. Dan Taylor's "Ten Principles of Good Level Design" (GDC 2013) makes this the anchor principle: good level design is driven by mechanics (`gamedeveloper.com/design/ten-principles-of-good-level-design-part-1-`).
- Start from the mechanic the level teaches or stresses, sourced from [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md) — never from a layout sketch that goes looking for a purpose afterward.
- Blockout first, art last. A level must prove itself as untextured geometry in playtest before any art pass; process and pacing for that pass live in [../recipes/design-a-level.md](../recipes/design-a-level.md) and [../quality/playtesting.md](../quality/playtesting.md).
- Every level names, in its one-page spec ([../templates/one-page-gdd.md](../templates/one-page-gdd.md)), the mechanic it introduces or develops and where it sits on the difficulty curve owned by [difficulty-and-balance.md](difficulty-and-balance.md).

### Principles Baseline

The handbook adopts Taylor's principle list as the default review vocabulary. The load-bearing subset:

- **Fun to navigate**: players always know where to go next, communicated through visual language — light, geometry, landmarks — not map markers or text.
- **Mechanics-driven**: the level exists to deliver the game's mechanics; a level that would work equally well in a different game is not doing its job.
- **Tells *what*, never *how***: the level states the goal and leaves the solution to the player. Signposting the solution deletes the learning moment that makes the level fun — fun is pattern-learning per Raph Koster's *A Theory of Fun*, the rationale owned by [../foundations/core-loops.md](../foundations/core-loops.md).
- **No reliance on words**: story and instruction are carried by the space itself (see Guidance Without Words below).

Full list and rationale: `gamedeveloper.com/design/ten-principles-of-good-level-design-part-1-` and `...-part-2-`.

### Introduce-Develop-Twist-Prove Structure

Default structure for any mechanic-teaching level is the Nintendo four-step (kishōtenketsu) pattern, as described by Koichi Hayashida for *Super Mario 3D World* (`gamedeveloper.com/design/the-secret-to-i-mario-i-level-design`):

1. **Introduce** — present the mechanic in a safe, low-stakes space where failure is cheap and the mechanic is the only new thing on screen.
2. **Develop** — repeat the mechanic in harder configurations: added hazards, tighter timing, combination with one already-mastered mechanic.
3. **Twist** — subvert the established expectation: invert the mechanic, recontextualize it, or collide it with a system the player has not seen it touch.
4. **Prove** — a final test the player must pass using the mechanic without scaffolding, demonstrating mastery before the level ends.

Rules on top of the template:

- One new mechanic per level by default. Spreading introductions out is the same rule George Fan applies to onboarding pacing, owned by [ux-and-onboarding.md](ux-and-onboarding.md).
- The beats escalate in that order. A twist before the develop beat tests a skill the player was never given room to acquire.
- The prove beat is a check, not a spike: it certifies the level's own teaching, not the whole difficulty curve. Curve-level escalation belongs to [difficulty-and-balance.md](difficulty-and-balance.md) and [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md).
- This structure is the level-scale form of fun-as-learning: each beat is a step in mastering one pattern, which is why it pairs with the loop model in [../foundations/core-loops.md](../foundations/core-loops.md).

### Guidance Without Words

Default to zero instructional text inside a level. The space carries the guidance:

- **Light and contrast** pull the eye toward the intended path; geometry frames the next objective in the player's default camera.
- **Landmarks** give the player a persistent orientation reference; offer sightlines to the same landmark from multiple vantage points so orientation survives detours (`book.leveldesignbook.com`).
- **A controlled first encounter** replaces a tutorial popup: the introduce beat is staged so the player experiments with the mechanic in an environment where the mechanic's feedback does the teaching — Fan's "use as few words as possible" and teach-with-visuals techniques, owned with the rest of tutorial design by [ux-and-onboarding.md](ux-and-onboarding.md).
- Readability is a hard constraint on decoration: no prop, effect, or lighting choice may make game state ambiguous. The readability rule and its cognitive grounding (Hodent) live in [ux-and-onboarding.md](ux-and-onboarding.md); the polish budget that competes with it lives in [../foundations/game-feel.md](../foundations/game-feel.md).
- Environmental storytelling — the level telling its story without words — is the level-design half of a contract shared with [narrative-integration.md](narrative-integration.md); that doc owns how narrative beats map onto spaces.
- Guidance must not assume full sight or hearing: pair every light/audio cue with a redundant channel per [accessibility.md](accessibility.md).

### Combat And Encounter Spaces

Encounter-space vocabulary and defaults follow the community-maintained Level Design Book (`book.leveldesignbook.com/process/combat/balance`):

- **Cover taxonomy**: soft cover blocks sight but not projectiles; hard cover blocks projectiles and usually sight; half cover protects a standing figure's top half; full cover protects a standing figure entirely. Name cover in these terms in specs and reviews.
- **No cover boxes**: avoid repetitive waist-high boxes; integrate cover into the architecture with varied shapes.
- **Sightlines are designed objects**: a sightline is the uninterrupted line from the player camera to an important part of the level. Vary sightline length and quantity per area so different spaces afford different visibility; rounded corners widen sightlines, sharp corners cut them.
- **Cover density by mode**: for PvE encounters err on too much cover (it encourages creative routing); for PvP err on too little (players make their own cover through mechanics).
- **Chokepoints and territory**: in competitive maps, distribute three to four chokepoints and make it impossible to defend all of them from one position; map control — the proportion of territory held — is the balance object.
- **No dominant positions**: every part of the map should be useful in some game state and weak in others. This is the spatial form of the no-dominant-strategy rule owned by [difficulty-and-balance.md](difficulty-and-balance.md).
- Encounter difficulty comes from configuration — enemy placement, sightlines, cover, timing — before stat inflation; stat-based escalation is a tuning decision that belongs to [difficulty-and-balance.md](difficulty-and-balance.md).

## Common Mistakes And Forbidden Patterns

- Designing the layout first and retrofitting mechanics into it — the level must be generated by what it teaches.
- Shipping instructional popups or on-screen text where a staged first encounter would teach the mechanic wordlessly.
- Introducing more than one new mechanic in a level without a documented reason in the level's spec.
- Placing the twist before the mechanic has been developed, or ending a teaching level without a prove beat.
- Signposting the solution — telling the player *how* instead of *what* deletes the mastery loop the level exists to serve.
- Cover boxes: repeated identical waist-high blocks standing in for designed cover.
- A single position from which all chokepoints can be held, or any always-correct location on a competitive map.
- Difficulty via enemy stat inflation where encounter configuration has not been exhausted.
- Guidance that exists only in one sensory channel (a light-only or audio-only cue) — an accessibility failure owned by [accessibility.md](accessibility.md).
- Art-passing a level that has not survived a blockout playtest.

## Verification And Proof

- Blockout playtest before art: a fresh (Kleenex) tester completes the level with no verbal or written instruction from the team; protocol in [../quality/playtesting.md](../quality/playtesting.md) and [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md).
- Beat audit: the level spec names its mechanic and points at the geometry implementing each of the four beats; a beat with no geometry is a finding.
- Navigation proof: in playtest, testers state where they think they should go next at each decision point; wrong answers mark guidance failures, not tester failures.
- Encounter audit: cover classified with the soft/hard/half/full taxonomy, sightlines walked from the player camera, and no dominant position found in adversarial playtest.
- Review gates: [../checklists/design-review.md](../checklists/design-review.md) before build, [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md) before the first external test.
- Process on rails: [../recipes/design-a-level.md](../recipes/design-a-level.md) is the step-by-step form of this doc.
