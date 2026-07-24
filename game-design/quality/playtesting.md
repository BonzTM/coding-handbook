# Playtesting

Owns the playtest as this handbook's primary proof mechanism: the session protocol, tester recruiting, observation discipline, and the routing that turns findings into changes. A design claim that has not survived a playtest is a hypothesis, not a fact.

## Default Approach

Follow the playcentric process: set experience goals first, then prototype, playtest, and revise continuously against those goals (Fullerton, *Game Design Workshop*, 5th ed. — `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`). The experience goals a session tests against are owned by [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md); the prototype under test is staged by [prototyping.md](prototyping.md). Playtesting is not a milestone event — it runs at every prototype stage, and no stage gate passes without one ([prototyping.md](prototyping.md)).

### Playtest Taxonomy

Match the session type to the question, the same way a test suite matches test type to risk. The stage names are a convention this handbook adopts for routing, built on Fullerton's structured protocol.

| Session Type | Use for | Testers |
|---|---|---|
| self and team test | does the mechanic function at all; daily iteration | the designers themselves |
| fresh-tester (Kleenex) test | first-time experience, onboarding, discoverability, clarity | people who have never seen the build, used exactly once |
| target-audience test | do the experience goals land for the intended player | recruits matching the audience in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) |
| tuning test | difficulty curve, balance, economy pacing after a change | returning testers with known skill; familiarity is acceptable here |
| readiness smoke | does the build survive a session without blocking bugs | anyone, before every external session |

Self-testing never substitutes for external testing: the team cannot experience the game as a first-time player, and the questions that matter most — clarity, onboarding, whether the loop is fun — are exactly the ones self-testing cannot answer.

### Session Protocol

Every session follows the same fixed shape. The script lives in [../templates/playtest-script.md](../templates/playtest-script.md), the step-by-step run procedure in [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md), and the build gate in [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md).

1. **State the questions first.** Write down what this session must answer — specific, falsifiable questions derived from lenses ([Lenses As Hypothesis Sources](#lenses-as-hypothesis-sources)), an open design decision ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)), or a prior session's unresolved findings. A session with no questions produces anecdotes, not proof.
2. **Gate the build.** Run [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md). A session lost to a blocking bug burns testers you cannot reuse.
3. **Script the session.** Fixed intro, fixed task framing, fixed time box, fixed observation items, fixed post-play interview questions. Ad-hoc sessions are not comparable across testers.
4. **Run it silent.** One facilitator delivers the script; observers log and never speak. No coaching, no explaining, no defending the design mid-session ([Observation Over Interrogation](#observation-over-interrogation)).
5. **Interview after play, not during.** Open, non-leading questions from the script; player commentary is recorded as claims, not findings.
6. **Report within a day.** File [../templates/playtest-report.md](../templates/playtest-report.md) while observations are fresh, then route every finding ([Turning Findings Into Changes](#turning-findings-into-changes)).

The intro script must state that the game is being tested, not the player, and that getting stuck is useful data — otherwise testers perform instead of playing.

### Recruiting And Kleenex Testers

Fresh testers are a consumable resource. A Kleenex tester — the term popularized by Will Wright — is used exactly once: someone who has never seen the game reveals where the design is confusing, and once they have played, they can never be a first-time tester again (`masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`; practitioner write-up: `gamedeveloper.com/business/gamedev-weekly-digest-02-the-power-of-anticipation-in-game-design-using-kleenex-testers`).

- **Keep a tester roster.** Record who has seen which build and when. Spending a fresh tester on a build that was not ready is an unrecoverable loss; the readiness checklist exists to prevent it.
- **Kleenex testers are mandatory for onboarding questions.** Tutorial, first-session, and discoverability findings are only valid from first-time players; the onboarding contract they validate is owned by [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md).
- **Teammates are never fresh.** Developers, and anyone who has watched a session or a stream of the build, are disqualified from first-time-experience questions. They remain useful for tuning sessions.
- **Recruit against the audience, not availability.** Target-audience sessions require testers matching the player defined in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md); friends-and-family convenience samples answer "can anyone play this," not "does this land for our player."
- **Reuse deliberately.** Returning testers are the right instrument for tuning and difficulty sessions, where accumulated skill mirrors the shipped player's — route those findings through [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md).

### Observation Over Interrogation

The protocol privileges what players do over what they say (Fullerton's structured playtest protocol — `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`). Behavior is evidence; self-report is a claim about evidence. The grounding is game UX practice: usability failures — missed signs, misread feedback, overloaded attention — show up in behavior whether or not the player can articulate them (Hodent, *The Gamer's Brain*, 2017 — `thegamersbrain.com`).

- Log observable events against the script's observation items: where the player hesitated, what they never found, what they tried that failed, where they died or quit, what they misread during effects ([../foundations/game-feel.md](../foundations/game-feel.md) readability items), and visible emotional reactions with timestamps.
- Never help a stuck player before the script's stuck-time limit expires. The struggle is the data; rescuing the tester destroys the finding and contaminates everything after it.
- Answer in-session questions with questions ("what do you think it does?") or silence per the script — an explained mechanic can never again be tested for discoverability with that tester.
- In the interview, ask open questions about what happened before any question about opinions, and never lead ("did you find the shop?" teaches the tester the shop exists).
- Record player-proposed solutions as symptoms, not designs. Testers are authoritative about what they experienced and unreliable about why and what to change; the designer owns the fix.

### Lenses As Hypothesis Sources

Schell's lens method — sets of questions that interrogate a design from different perspectives (*The Art of Game Design: A Book of Lenses*, 3rd ed. — `routledge.com/The-Art-of-Game-Design-A-Book-of-Lenses-Third-Edition/Schell/p/book/9781138632059`) — is complementary to playtesting: lenses generate hypotheses, playtests check them. The frameworks themselves, and when to use which, are owned by [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md).

- Derive each session's questions from the lenses relevant to the current prototype stage: the loop-and-motivation lenses for a core-loop prototype ([../foundations/core-loops.md](../foundations/core-loops.md), [../foundations/player-psychology.md](../foundations/player-psychology.md)), curiosity and clarity lenses for onboarding builds, challenge lenses for tuning builds.
- A lens answered from the armchair is still a hypothesis. The lens tells you what to look for; only the session tells you what is there.
- Keep the per-session question list short enough that observers can actually watch for all of it. Unanswered questions carry forward to the next session's list rather than bloating this one.

### Turning Findings Into Changes

A playtest that does not change something — the design, a tuning value, a documented decision, or a now-validated hypothesis — was entertainment. The report template ([../templates/playtest-report.md](../templates/playtest-report.md)) separates three layers, and reviews hold the separation:

1. **Observation** — what happened, verbatim from the log. Never edited to fit a theory.
2. **Interpretation** — the designer's claimed cause, marked as confirmed (seen across testers) or suspected (needs another session).
3. **Decision** — change, investigate further, or accept as intended, each with an owner.

Route every decision to the doc that owns the surface, mirroring the Change Routing table in [../AGENTS.md](../AGENTS.md):

- Confused in the first minutes, missed mechanics, unread UI → [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md).
- Loop is boring, no reason to continue, goals unclear → [../foundations/core-loops.md](../foundations/core-loops.md) and [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md) — never juice ([../foundations/game-feel.md](../foundations/game-feel.md) forbids polish as the fix for a boring loop).
- Difficulty spikes, dominant strategies, quit-points at walls → [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md).
- Currency or reward pacing complaints, hoarding, nothing to spend on → [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md).
- Lost in the space, backtracking, missed critical path → [../disciplines/level-design.md](../disciplines/level-design.md).
- Controls feel bad, effects obscure state → [../foundations/game-feel.md](../foundations/game-feel.md).
- Story beats skipped or contradicted by play → [../disciplines/narrative-integration.md](../disciplines/narrative-integration.md).

Findings that would change a pillar or a recorded decision do not get patched silently — they go through a DDR ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)). One tester's outlier reaction is a note; the same failure across testers is a finding. Do not redesign after every session — batch, look for the pattern, then change one thing at a time so the next session can attribute the effect.

### Pairing With Telemetry

Telemetry tells you what players do; playtests tell you why — the canonical caveat is to pair quantitative telemetry with qualitative sessions rather than trust either alone (*Game Analytics*, Seif El-Nasr, Drachen & Canossa, eds. — `link.springer.com/book/10.1007/978-1-4471-4769-5`). Instrumentation, funnels, and live tuning are owned by [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md); the contract at this boundary:

- Instrument the design questions, not everything. If a playtest question will recur at scale — where do players quit, which option is never picked — it becomes a telemetry event with a named owner.
- Telemetry anomalies route back into playtest questions: a drop-off spike at level 3 is a *what*; the next session's script watches level 3 to find the *why*.
- Neither source overrides the other by default. A metric that contradicts observed sessions means the instrument or the interpretation is wrong — investigate before tuning.

## Common Mistakes And Forbidden Patterns

- Sessions with no written questions, producing anecdotes that confirm whatever the team already believed.
- Spending Kleenex testers on a build that fails the readiness checklist, or letting a teammate stand in for a first-time player.
- No tester roster, so nobody knows who is still fresh and first-time-experience findings quietly come from repeat players.
- Coaching a stuck tester before the stuck-time limit, or explaining a mechanic mid-session and then trusting later discoverability observations.
- Leading interview questions, or interviewing during play instead of after.
- Treating player-proposed solutions as design decisions instead of symptoms.
- Reporting interpretations as observations — the log edited to fit the theory.
- Findings filed with no route and no owner, or a session report that changed nothing and validated nothing.
- Redesigning after every single session, or changing several things at once so the next session cannot attribute the effect.
- Fixing "the game is boring" with polish instead of routing to the loop ([../foundations/game-feel.md](../foundations/game-feel.md)).
- Pillar-breaking findings patched silently instead of going through a DDR.
- Trusting telemetry to explain why, or a single observed session to establish what happens at scale.

## Verification And Proof

- Every session has a script ([../templates/playtest-script.md](../templates/playtest-script.md)), a passed readiness check ([../checklists/playtest-readiness.md](../checklists/playtest-readiness.md)), and a report ([../templates/playtest-report.md](../templates/playtest-report.md)) filed within a day.
- The report separates observation, interpretation, and decision, and every decision names an owner and a routed doc.
- The tester roster shows which testers are spent, and all first-time-experience findings trace to genuinely fresh testers.
- Each session's questions trace to a lens, an open DDR, or a prior session's carry-forward list — and each gets an answer, a follow-up session, or an explicit carry-forward.
- Prototype stage gates in [prototyping.md](prototyping.md) reference a specific session report as their evidence, not a verbal "it tested fine."
- Design changes since the last review each cite the finding that motivated them; pillar-level changes each have a DDR ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)).
- Recurring playtest questions have corresponding telemetry events registered in [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).

Playtesting is done for a build when its stage-gate questions have evidence-backed answers — not when a session has merely been held.

Related: [prototyping.md](prototyping.md), [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md), [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md), [../templates/playtest-script.md](../templates/playtest-script.md), [../templates/playtest-report.md](../templates/playtest-report.md), [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md), [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)
