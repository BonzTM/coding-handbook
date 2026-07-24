# Playtest Readiness Checklist

Gate before spending testers: [../quality/playtesting.md](../quality/playtesting.md) owns the method and [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md) owns the session procedure; this checklist gates whether the session is worth running at all. Walk it per playtest, before confirming recruits — a fresh tester who meets a broken build or an unscripted session is spent forever (Will Wright's Kleenex-testing rule: first-timers can only be first-timers once, `masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`).

## Build Readiness

- [ ] The build launches, reaches the content under test, and survives a full scripted session without a crash — proven by a dry run on the exact device or machine testers will use, not on a dev workstation.
- [ ] The build is pinned to an exact version (tag, commit, or build ID) so every finding in the report maps to one known artifact.
- [ ] Known blockers on the tested path are fixed or fenced off; anything not under test is either stable-but-reachable or cut from the build, never half-working.
- [ ] Debug overlays, cheats, and placeholder text that would contaminate a first impression are hidden — first-time-experience data is unrepeatable per tester.
- [ ] A mid-session recovery path exists (save, restart point, or facilitator reset) so one bug does not end the session.

## Script And Questions

- [ ] The design questions this session must answer are written before the script, and each traces to a pillar in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) or an open decision — lenses generate the hypotheses, the playtest checks them (Schell, *The Art of Game Design*).
- [ ] What would count as an answer is defined per question before the session, so results cannot be reinterpreted afterward to flatter the build.
- [ ] A session script built from [../templates/playtest-script.md](../templates/playtest-script.md) exists — greeting, framing, tasks, prompts, debrief — so every tester gets the same session and results are comparable.
- [ ] The script observes rather than interrogates: no leading questions, no explaining the game mid-session beyond the scripted framing (Fullerton's playcentric protocol, *Game Design Workshop*).
- [ ] The facilitator knows the intervention rule in advance: rescue only on blocker bugs; a stuck or confused player is the data, not a problem to fix live.

## Recruiting

- [ ] Recruits match the design question: first-time-experience and onboarding questions get Kleenex testers who have never seen the game; depth and balance questions get returning or experienced testers — see [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md).
- [ ] Nobody is double-counted as fresh: the roster tracks who has already seen the game, and no prior tester is booked against a first-time question.
- [ ] Nobody who built the feature under test is a tester for it; team self-play answers different questions than external play.
- [ ] Session count, session length, and a no-show backup are set before invitations go out, so the schedule does not force cutting the debrief.

## Capture And Consent

- [ ] Every design question has a capture channel — observer notes, screen or gameplay recording, or telemetry — and telemetry is instrumented for the specific questions, not "collect everything" (see [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)).
- [ ] A note-taker separate from the facilitator is assigned; one person cannot run the script and capture behavior at once.
- [ ] Written consent for recording and data use is collected before the session starts; testers who are minors have guardian consent.
- [ ] Confidentiality (NDA, if the build is sensitive) is handled as a separate signature before exposure, not bundled into the consent form.
- [ ] Recordings and notes have a named storage location, an access list, and a deletion date; raw footage never leaves the team.

## Proof

- [ ] A full dry run of script plus build plus capture setup was completed end to end by someone who did not write the script.
- [ ] Every design question maps to at least one script task and one capture channel; any question with no mapping is cut from the session, and any task answering no question is cut from the script.
- [ ] The [../templates/playtest-report.md](../templates/playtest-report.md) skeleton is prefilled with build ID, questions, and answer criteria before the first tester arrives.
