<!--
Destination: design/playtests/<YYYY-MM-DD>-<FOCUS>-script.md

Playtest session script for <GAME/BUILD>. Contract:
- Every moderated playtest runs from a filled copy of this file; an unscripted session
  produces anecdotes, not evidence.
- Fill this file BEFORE recruiting; the readiness gate is
  game-design/checklists/playtest-readiness.md. Session results land in
  game-design/templates/playtest-report.md — one report per session, same date and focus.
- The protocol this template implements (recruit, script, observe over interrogate,
  iterate) is Fullerton's playcentric process; see game-design/quality/playtesting.md.
- Replace every <PLACEHOLDER>. Delete sections that genuinely do not apply, and say why.
-->

# Playtest Script: <FOCUS> — <YYYY-MM-DD>

One-line purpose: test <WHICH SYSTEM, LEVEL, OR LOOP> against <WHICH EXPERIENCE GOAL OR DESIGN QUESTION>.

## Session Goals

State what this session must answer before it starts. A session without a falsifiable question is a demo, not a test. Lenses and design reviews generate the hypotheses; the playtest checks them (Schell, *The Art of Game Design*, `routledge.com/The-Art-of-Game-Design-A-Book-of-Lenses-Third-Edition/Schell/p/book/9781138632059`).

- **Build under test:** <BUILD ID / COMMIT / VERSION> on <PLATFORM>.
- **Experience goal being tested:** <THE GOAL FROM YOUR PILLARS DOC — e.g. "players feel tension when X">. See game-design/foundations/design-pillars-and-vision.md for where goals live.
- **Design questions (max 3):**
  1. <QUESTION — e.g. "Do players discover the dash without prompting by encounter 2?">
  2. <QUESTION>
  3. <QUESTION>
- **Success/failure criteria per question:** <WHAT OBSERVED BEHAVIOR CONFIRMS OR REFUTES EACH — decide now, not after>.
- **Tester profile:** <TARGET-AUDIENCE FIT; PRIOR GENRE EXPERIENCE REQUIRED OR EXCLUDED>.
- **Fresh-tester requirement:** onboarding and first-time-comprehension questions require testers who have never seen the game; a used tester can never be a first-time tester again (Will Wright's Kleenex testing, `masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`). Mark each question: <FRESH REQUIRED: yes/no>.

## Setup

- **Roles:** facilitator <NAME>, observer/note-taker <NAME>. One person cannot do both for more than <N> testers; recruit a second observer above that.
- **Location / remote tooling:** <ROOM OR SCREEN-SHARE + RECORDING TOOL>.
- **Recording:** <SCREEN CAPTURE / FACE CAM / NONE> — consent obtained via <CONSENT FORM / VERBAL, LOGGED>.
- **Build state:** installed and smoke-tested on the test machine <WHEN>; save/checkpoint prepared at <STARTING STATE>; debug overlays and cheats <DISABLED / ENABLED FOR: REASON>.
- **Telemetry:** <EVENTS THIS SESSION SHOULD CAPTURE, IF INSTRUMENTED>. Telemetry records what testers do; this session exists to observe why — pair both, trust neither alone (see game-design/operations/live-tuning-and-telemetry.md).
- **Session length:** <MINUTES> play + <MINUTES> debrief. Hard stop at <TIME>.
- **Facilitator opening statement (read verbatim):** "<WE ARE TESTING THE GAME, NOT YOU. NOTHING YOU DO IS WRONG. PLEASE THINK ALOUD AS YOU PLAY. WE WILL NOT HELP UNLESS YOU ARE STUCK FOR SEVERAL MINUTES — THAT SILENCE IS US LEARNING, NOT RUDENESS.>"

## Warm-Up

Two minutes, maximum. The warm-up calibrates the tester's baseline and gets them talking; it must not teach the game.

- **Background questions (before play):**
  - <"WHAT GAMES HAVE YOU PLAYED RECENTLY?">
  - <"HAVE YOU PLAYED ANYTHING IN THIS GENRE? WHICH?">
  - <QUESTION THAT CHECKS THE TESTER PROFILE ASSUMPTION ABOVE>
- **Think-aloud practice:** <ONE TRIVIAL NON-GAME TASK TO REHEARSE NARRATING, e.g. "talk me through unlocking your phone">.
- **What the facilitator may say about the game before play:** <THE EXACT ONE-OR-TWO-SENTENCE FRAMING — WRITE IT HERE SO EVERY TESTER HEARS THE SAME THING>. Nothing beyond this; if a tester needs more to start, that is a finding.

## Tasks And Prompts

Order tasks from open to directed: free play first (what do testers do unprompted?), directed tasks after (can they do what the design intends?). Write every prompt verbatim — improvised prompts leak hints and make sessions incomparable.

### Task 1: Free Play

- **Prompt (verbatim):** "<PLAY HOWEVER YOU LIKE. THINK ALOUD.>"
- **Duration:** <MINUTES>.
- **Watching for:** <WHICH DESIGN QUESTION THIS FEEDS; EXPECTED VS INTERESTING BEHAVIOR>.
- **Intervention rule:** help only after <N> minutes stuck; log the intervention and the words used.

### Task 2: <DIRECTED TASK NAME — e.g. Reach The End Of Level 2>

- **Prompt (verbatim):** "<...>"
- **Start state:** <SAVE / CHECKPOINT>.
- **Success looks like:** <OBSERVABLE COMPLETION>.
- **Abandon after:** <TIME OR FAILURE COUNT> — abandonment is data, not a session failure.
- **Watching for:** <WHICH DESIGN QUESTION THIS FEEDS>.

### Task 3: <TASK NAME>

- **Prompt (verbatim):** "<...>"
- <SAME FIELDS AS ABOVE>

<!-- One block per task. Every task must feed a Session Goals question; cut tasks that feed none. -->

- **Mid-play probes (only these, only when the tester pauses):** "<WHAT ARE YOU TRYING TO DO?>", "<WHAT DO YOU EXPECT THAT TO DO?>", "<WHAT ARE YOU LOOKING AT?>". Never ask "do you like it?" during play; never explain a mechanic the game is supposed to teach.

## Observation Notes

Structure the note-taking before the session so observers record behavior, not opinions. Watching what testers do outweighs asking what they think — the protocol is observation over interrogation (Fullerton, *Game Design Workshop*, `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`).

| Timestamp | Task | Observed behavior (verbatim quote or action) | Maps to question # | Severity <blocker/friction/note> |
|---|---|---|---|---|
| <MM:SS> | <TASK> | <WHAT HAPPENED, NOT WHAT IT MEANS> | <1/2/3> | <...> |

- **Always log:** deaths/failures and where; hesitations longer than <SECONDS>; every facilitator intervention and its wording; misreads of UI or feedback (usability signal — signs and feedback, per Hodent, *The Gamer's Brain*, `amazon.com/Gamers-Brain-Neuroscience-Impact-Design/dp/1498775500`); moments the tester stops thinking aloud (often absorption or confusion — note which it looks like).
- **Do not log interpretations mid-session.** Diagnosis happens in the report, not the note sheet.

## Debrief Questions

Ask after play ends, open questions first, specifics after. Answers are self-report — weight them below observed behavior when they conflict, and record them anyway.

1. "<DESCRIBE WHAT YOU JUST PLAYED TO A FRIEND IN A SENTENCE OR TWO.>" (comprehension check against the experience goal)
2. "<WHAT WAS THE BEST MOMENT? WALK ME BACK TO IT.>"
3. "<WHAT WAS THE MOST FRUSTRATING OR CONFUSING MOMENT?>"
4. "<WHAT DID <MECHANIC/UI ELEMENT UNDER TEST> DO? HOW DID YOU FIGURE THAT OUT?>"
5. "<WHAT WOULD YOU DO NEXT IF YOU KEPT PLAYING?>" (pull/retention signal)
6. <QUESTION TARGETING AN UNRESOLVED SESSION GOAL — WRITE ONE PER OPEN QUESTION>
7. "<ANYTHING YOU EXPECTED THE GAME TO LET YOU DO THAT IT DIDN'T?>"

- **Never ask:** leading questions ("did you notice the tutorial hint?"), design-solution questions ("should the dash be faster?"). Testers report experience; the team designs the fix.
- **Close:** thank the tester, log compensation <AMOUNT/FORM>, and record whether they are now burned for fresh-tester purposes.

## Related

- Protocol and cadence this script implements: game-design/quality/playtesting.md
- Go/no-go gate before scheduling: game-design/checklists/playtest-readiness.md
- Where findings land: game-design/templates/playtest-report.md
- End-to-end walkthrough: game-design/recipes/run-a-playtest.md
- Onboarding-specific comprehension testing: game-design/disciplines/ux-and-onboarding.md
