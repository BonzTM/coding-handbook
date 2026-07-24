# Design Pillars And Vision

Experience goals and design pillars are the first artifact of any project: the written statement of what the player should feel, and the decision filters every later change is tested against.

## Default Approach

Set experience goals before designing mechanics. Tracy Fullerton's playcentric process (*Game Design Workshop*, 5th ed., `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`) makes this the founding move: define the experience you want players to have, then prototype, playtest, and revise continuously against those goals. Pillars operationalize the goals — they are the short list of commitments a proposal must serve to enter the game. No prototype starts, and no concept passes [../checklists/concept-intake.md](../checklists/concept-intake.md), until both exist in writing.

### Experience Goals First

- Write experience goals as player experiences, not feature lists: what the player feels, decides, or masters — never what the game contains. "The player feels outnumbered but clever" is a goal; "stealth mechanics and 12 enemy types" is an inventory.
- Keep two to four goals. More than four means the team has not decided what the game is.
- Every goal must be checkable in a playtest: phrase each one so a session with fresh players can confirm or refute it. Goals that cannot fail in a playtest are decoration. Fullerton's protocol for running that check is owned by [../quality/playtesting.md](../quality/playtesting.md).
- Goals sit on the player's side of the MDA gap: designers build mechanics bottom-up, but players experience aesthetics top-down (Hunicke, LeBlanc & Zubek, "MDA: A Formal Approach to Game Design and Game Research", 2004, `users.cs.northwestern.edu/~hunicke/MDA.pdf`). Writing goals first forces the team to start from the player's reading, not the builder's.

### Pillars As Decision Filters

- Derive three to five pillars from the experience goals. Each pillar is a short imperative phrase plus one sentence of meaning — e.g. "Every death is the player's fault: no unavoidable damage, no off-screen kills."
- A pillar earns its place by rejecting things. If a pillar has never caused the team to cut, simplify, or redesign a feature, it is a slogan, not a filter. Genre labels ("great combat", "immersive world") reject nothing and are forbidden as pillars.
- Every feature proposal, scope change, and tuning direction names the pillar it serves. A change that serves no pillar is cut, or it becomes evidence the pillar set is wrong — which routes to a pillar revision, never a silent exception.
- Pillars and goals live on the project's one-page overview so the whole team sees them daily — Stone Librande's one-page method exists because design documentation must communicate efficiently and "most people only read the first page anyway" ("One-Page Designs", GDC 2010, `gdcvault.com/play/1012356/One-Page`). The page format is owned by [design-documentation.md](design-documentation.md) and [../templates/one-page-gdd.md](../templates/one-page-gdd.md); produce it via [../recipes/write-a-one-page-gdd.md](../recipes/write-a-one-page-gdd.md).

### Aesthetic Targets

- Name the intended aesthetics in the shared vocabulary this handbook adopts — MDA's aesthetics terms and their successors are defined once in [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md); use that vocabulary rather than inventing per-project adjectives.
- For each experience goal, sketch the dynamics expected to produce it: which runtime behaviors, emerging from which mechanics, should generate the target feeling. This is the designer's hypothesis; playtests are the experiment. Mechanic-level design that follows from these targets is owned by [core-loops.md](core-loops.md) and [mechanics-and-systems.md](mechanics-and-systems.md).
- Ground aesthetic targets in what actually motivates the intended audience — motivation models and their evidence base are owned by [player-psychology.md](player-psychology.md); cite a motivation the target audience demonstrably has, not one the team hopes for.
- Treat MDA as shared vocabulary, not settled law: it is the most widely employed framework and also criticized for neglecting narrative and audiovisual design, with no successor (DDE included) having replaced it (Walk, Görlich & Barrett, "Design, Dynamics, Experience (DDE)", 2017, `link.springer.com/chapter/10.1007/978-3-319-53088-8_3`). The critique and the adoption decision live in [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md).

### Vision Drift Control

- Pillars are stable by default and versioned in the design repo. Changing, adding, or retiring a pillar requires a design decision record via [../decisions/design-decision-records.md](../decisions/design-decision-records.md) — never an edit made in passing.
- Every design review asks "which pillar does this serve?" as a standing gate; the question is part of [../checklists/design-review.md](../checklists/design-review.md).
- Playtest reports evaluate findings against the experience goals, not against feature completeness — the report shape is owned by [../templates/playtest-report.md](../templates/playtest-report.md). A goal that repeatedly fails in playtests is a real signal: revise the design toward the goal, or revise the goal on the record.
- Log each time a pillar rejects a proposal. This rejection log is the cheapest proof the pillars are load-bearing, and it is the first thing a new owner reads during [../onboarding-and-handoff.md](../onboarding-and-handoff.md).
- Scope pressure is the main drift vector: when production forces cuts, cut the features furthest from the pillars first. The cut mechanics are owned by [../operations/scoping-and-production.md](../operations/scoping-and-production.md).

## Common Mistakes And Forbidden Patterns

- Writing mechanics first and retrofitting pillars afterward to bless what already exists.
- Pillars that are genre labels, marketing taglines, or quality adjectives ("fun", "polished", "immersive") — none of these can reject a feature.
- More than five pillars, or pillar sets where every proposal plausibly serves at least one — a filter that passes everything filters nothing.
- Experience goals phrased as feature inventories or platform bullet points instead of player experiences.
- Goals that cannot fail in a playtest, so no session ever produces evidence against them.
- Editing a pillar silently mid-production instead of routing the change through a decision record.
- A pillars document that exists but is not on the one-page overview, so the team stops seeing it and drift goes unnoticed.
- Treating a repeated playtest failure against a goal as a playtester problem rather than a design or goal problem.

## Verification And Proof

- The pillars-and-goals document exists, is versioned, and predates the first prototype — check the repo history.
- Each experience goal is phrased as a player experience and paired with at least one playtest question that could refute it.
- Each pillar has at least one logged rejection or redesign it caused; a pillar with none after a milestone is flagged for revision.
- Concept intake and design review both cite pillars: spot-check recent [../checklists/concept-intake.md](../checklists/concept-intake.md) and [../checklists/design-review.md](../checklists/design-review.md) outcomes for the pillar question answered, not skipped.
- Every pillar change since project start has a corresponding decision record in [../decisions/design-decision-records.md](../decisions/design-decision-records.md).
- The latest playtest report scores findings against the experience goals by name.

Related: [core-loops.md](core-loops.md), [design-documentation.md](design-documentation.md), [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md), [../quality/playtesting.md](../quality/playtesting.md), [../glossary.md](../glossary.md).
