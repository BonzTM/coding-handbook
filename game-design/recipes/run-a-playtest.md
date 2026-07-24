# Recipe: Run A Playtest

Use this when a prototype or build is ready to be tested against open design questions and you need findings a design review can act on. Method and rationale are owned by [../quality/playtesting.md](../quality/playtesting.md); this is the session-execution procedure.

## Files To Touch

- the design questions for this session, pulled from the open questions in the one-page GDD or an open DDR ([../templates/ddr-template.md](../templates/ddr-template.md))
- a session script instantiated from [../templates/playtest-script.md](../templates/playtest-script.md)
- [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md) — the gate before any tester is scheduled
- a report instantiated from [../templates/playtest-report.md](../templates/playtest-report.md), filed with the project's design docs

## Steps

1. Write the design questions first. Each session tests the build against stated experience goals, not against "do they like it" — that is the core of Fullerton's playcentric process (*Game Design Workshop*, 5th ed., `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`). Schell's lenses are the hypothesis generator; the playtest is the check (*The Art of Game Design*, `routledge.com/The-Art-of-Game-Design-A-Book-of-Lenses-Third-Edition/Schell/p/book/9781138632059`). A session with no named questions is canceled, not run.
2. Run the readiness gate. Every box in [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md) passes or the session is rescheduled; a session that dies to a known crash or a missing build wastes the one thing you cannot refund — tester freshness.
3. Recruit testers matched to the questions. Onboarding, first-session, and clarity questions require Kleenex testers — people who have never seen the game, used exactly once, because a used tester can never be a first-time tester again (Will Wright, `masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`). Balance and mastery questions require experienced testers instead. Never mix the two pools in one finding.
4. Fill in the script: greeting and framing ("we are testing the game, not you"), consent for recording, the task or free-play block per design question, observation prompts, and the debrief questions. The fixed shape lives in [../templates/playtest-script.md](../templates/playtest-script.md); improvised sessions produce unattributable findings.
5. Run the session. Observe; do not interrogate, coach, or defend the design. If a tester is stuck, the stuck point is the finding — log it and let them struggle to the script's intervention threshold before helping.
6. Capture during play, not from memory afterward: timestamped observation notes tied to the design question each supports, plus recordings or instrumented build data where available. Telemetry records what players did; only observation and debrief get at why — pair them, never substitute one for the other (Seif El-Nasr, Drachen & Canossa, *Game Analytics*, `link.springer.com/book/10.1007/978-1-4471-4769-5`).
7. Debrief with the scripted questions only after play ends. Ask about what happened, not what they would change; testers are reliable witnesses of their own confusion and unreliable designers.
8. File the report against the design questions using [../templates/playtest-report.md](../templates/playtest-report.md): raw observations separated from interpretation, each question marked answered, partially answered, or unanswered, with evidence. Route resulting design changes through [../checklists/design-review.md](../checklists/design-review.md) and record reversals of standing decisions as DDRs per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

## Invariants To Preserve

- every session runs against named design questions written before recruiting
- readiness gate passes before any tester is scheduled; tester freshness is never spent on a broken build
- Kleenex testers are used exactly once and only for first-time-experience questions
- the facilitator observes and never explains, hints, or defends the design mid-session
- raw observations stay separated from interpretation in the report
- findings cite evidence (timestamped note, recording, or metric), not facilitator memory
- one session produces one filed report; sessions without a report did not happen

## Proof

- the filled script and the filed report exist and reference the same design questions
- each design question in the report carries a verdict and its supporting evidence
- observation notes are timestamped and attributable to a tester and a session block
- design changes triggered by the report appear in the design-review queue or as DDRs, not as silent doc edits

If the question is which numeric values to change rather than whether the design works, run this recipe to gather evidence, then continue with [tune-a-difficulty-curve.md](tune-a-difficulty-curve.md) or [balance-an-economy.md](balance-an-economy.md).
