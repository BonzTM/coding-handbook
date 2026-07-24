# UX And Onboarding

Usability and onboarding rules that get players into the core loop without a manual: signs and feedback for every interaction, teach by doing, minimal words, one new element at a time.

## Default Approach

### Scope

- Game UX is not menus and fonts. Per Celia Hodent (*The Gamer's Brain*, CRC Press 2017, `thegamersbrain.com`), UX covers everything between the player's intent and the game's response, grounded in how perception, attention, and memory actually work — not designer intuition.
- UX decisions are testable claims about player behavior. Every rule in this doc has a proof mechanism in [../quality/playtesting.md](../quality/playtesting.md); do not settle UX arguments by opinion when a five-person test settles them by evidence.
- Onboarding serves the core loop, not the other way around. Before designing a tutorial, the loop it teaches must already be defined in writing — see [../foundations/core-loops.md](../foundations/core-loops.md) and [../recipes/design-a-core-loop.md](../recipes/design-a-core-loop.md).

## Usability And Engage-Ability Split

Hodent splits game UX into two components. Keep them separate in reviews, because their failure modes and fixes differ.

| Component | Question it answers | Failure looks like | Owning material |
|---|---|---|---|
| Usability | Can players do what they intend? | confusion, misclicks, "I didn't know I could do that" | this doc: signs and feedback, cognitive load |
| Engage-ability | Do players want to keep doing it? | boredom, churn despite full comprehension | [../foundations/player-psychology.md](../foundations/player-psychology.md), [../foundations/game-feel.md](../foundations/game-feel.md) |

- Diagnose usability before engage-ability. A player who cannot parse the screen cannot be motivated by it; fix comprehension failures first, then evaluate motivation.
- Never fix a usability failure with an engage-ability tool. More juice on a button players cannot find is noise, not a fix — practitioner pushback on "Juice It or Lose It" makes exactly this point (`gamedeveloper.com/design/video-indies-resist-the-urge-to-juice-it-or-lose-it-`).
- Never fix an engage-ability failure with a usability tool. If players understand the loop and still quit, adding tooltips changes nothing; route the problem to [../foundations/player-psychology.md](../foundations/player-psychology.md) or [difficulty-and-balance.md](difficulty-and-balance.md).

## Signs And Feedback

- Every interactive element carries a sign (it looks actionable before touch) and feedback (the game responds visibly within the same beat as the input). An element missing either is a usability defect, not a polish gap.
- Distinguish the three channels and keep each honest:
  1. **Signs** advertise what can be done: silhouette, color, animation, placement.
  2. **Feedback** confirms what was done: sound, particle, number, state change.
  3. **Feedforward** previews what will happen: telegraphs, range indicators, ghost previews.
- State changes the player must act on get redundant channels (visual plus audio at minimum) so a single missed cue does not strand them; the full input/output-redundancy requirement is owned by [accessibility.md](accessibility.md).
- Feedback intensity scales with consequence. Reserve the loudest effects (hitstop, screen shake, full-screen flash) for the events that matter most; the intensity vocabulary and its readability limits are owned by [../foundations/game-feel.md](../foundations/game-feel.md).
- Never let feedback obscure state. Any effect that hides the information a player needs for the next decision is forbidden regardless of how good it feels — juice amplifies a readable core, it does not substitute for one.
- In-world signs beat UI overlays. Dan Taylor's level design principles put wayfinding in light, geometry, and visual language before words or markers (`gamedeveloper.com/design/ten-principles-of-good-level-design-part-1-`); spatial signposting rules are owned by [level-design.md](level-design.md).

## Onboarding Rules

Defaults from George Fan's "How I Got My Mom to Play Through Plants vs. Zombies" (GDC 2012, `gdcvault.com/play/1015541/How-I-Got-My-Mom`), the canonical tutorial-design talk:

1. **Teach by doing.** Blend the tutorial into the game: the player experiments in a controlled environment rather than reading instructions. A screen of text before play is a design failure, not a shortcut.
2. **Minimal words.** Use as few words as possible; teach with visuals first. Every tutorial sentence must survive the question "what would the player fail to learn without it?" — cut the rest.
3. **One new element at a time.** Spread mechanic introductions out — Fan's guideline is roughly one new element every few levels. Never introduce two systems in the same beat and expect both to land.
4. **Leverage what players already know.** Genre conventions, platform conventions, and real-world metaphors are free teaching; spend explanation budget only on what is genuinely novel.
5. **Structure each introduction as introduce, develop, twist, prove.** The kishōtenketsu four-step used in Nintendo's mechanic-driven levels (`gamedeveloper.com/design/the-secret-to-i-mario-i-level-design`) — the level-scale template is owned by [level-design.md](level-design.md).
6. **No skippable-knowledge traps.** If a tutorial can be skipped, everything it teaches must remain discoverable later through signs and feedback; skipping convenience must never create a permanently confused player.
7. **Onboarding extends past minute one.** Any system introduced mid-game (crafting at hour three, a new faction currency) gets the same treatment: taught by doing, one element, minimal words. Mid-game economy introductions coordinate with [economy-and-progression.md](economy-and-progression.md).

## Cognitive Load Budget

- Working memory and attention are scarce; Hodent's core argument is that onboarding must be designed against those limits, not against what the design team — who cannot un-know the game — finds obvious.
- Treat each simultaneous novel element as spend against a small fixed budget. New mechanic, new resource, new control, new UI region, new vocabulary word: each costs one. When a proposed beat spends more than one, split the beat.
- Distribute learning across sessions. Retention of a taught mechanic is verified by watching a returning player use it unprompted, not by completion of the tutorial that taught it.
- Reduce load with defaults, not documentation: safe starting choices, delayed exposure of advanced options, and interfaces that show only what the current decision needs.
- The team is disqualified from judging load. Only fresh testers who have never seen the game can reveal where it confuses — Will Wright's Kleenex-testing rule (`masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`), owned by [../quality/playtesting.md](../quality/playtesting.md). Each fresh tester is spent once; budget them for onboarding tests specifically.

## Common Mistakes And Forbidden Patterns

- Opening with a text or video wall instead of a playable teaching environment.
- Introducing more than one new element in a single beat because the schedule wants a shorter tutorial.
- Explaining in words what a sign or a level layout could teach silently.
- Locking veteran players in an unskippable tutorial while giving skippers no in-game path to the same knowledge.
- Patching a confusing mechanic with a tooltip instead of fixing the sign or feedback that failed.
- Adding juice to mask a usability defect, or letting feedback effects hide decision-relevant state.
- Validating onboarding with developers or repeat testers instead of fresh Kleenex testers.
- Treating onboarding as a first-hour feature and shipping mid-game systems with no introduction at all.
- Measuring tutorial completion rate and calling it comprehension — completion proves exposure, not learning.

## Verification And Proof

- Run a fresh-tester onboarding test before every milestone: script it with [../templates/playtest-script.md](../templates/playtest-script.md), gate it with [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md), and execute per [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md). Observation over interrogation; no hints from the facilitator.
- Pass condition: the fresh tester reaches the core loop and completes it once without facilitator help, and can state in their own words what they were trying to do at each step.
- Retention check: a returning tester uses each taught mechanic unprompted in a later session; a mechanic that must be re-taught was not taught.
- Sweep every interactive element against the signs-and-feedback rule: actionable look before touch, visible response after. Log misses as defects with the same severity as functional bugs.
- Instrument onboarding funnel steps (tutorial beat reached, first unprompted use of each mechanic, first loop completion) and watch drop-off per beat in [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md); telemetry locates the failing beat, the next fresh-tester session explains it.
- File the outcome with [../templates/playtest-report.md](../templates/playtest-report.md); an onboarding change without a report from a fresh-tester session is unverified.
