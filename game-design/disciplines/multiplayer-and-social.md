# Multiplayer And Social Design

Design-side standards for player-versus-player fairness, matchmaking, asymmetric modes, cooperation, disruptive-behavior prevention, and spectating. This doc owns what the social experience is; how state crosses the wire — peers, transports, RPCs, server authority — is owned by [../../godot/systems/multiplayer.md](../../godot/systems/multiplayer.md).

## Default Approach

A multiplayer game ships two products: the game and the crowd the game assembles. Every player-to-player surface — matchmaking, chat, trading, collision, spectating — is designed and reviewed like a mechanic, because it is one: Riot's League of Legends experiments showed that design choices as small as a loading-screen sentence measurably change player behavior (Jeffrey Lin, "The Science Behind Shaping Player Behavior in Online Games", GDC 2013 — `gdcvault.com/play/1017940/The-Science-Behind-Shaping-Player`). The motivation grounding — relatedness independently predicts enjoyment in multiplayer contexts, and is budgeted only where the design has a real social surface — is owned by [../foundations/player-psychology.md](../foundations/player-psychology.md).

### PvP Fairness And Counterplay

- Competitive balance rules — many viable options, no dominant option, counter-cycles at least three deep, executable counterplay — are owned by [difficulty-and-balance.md](difficulty-and-balance.md) (### Viable Options And Counterplay); do not re-derive them here.
- Sirlin's fairness definition — "players of equal skill have an equal chance to win, no matter their start conditions" (`sirlin.net/articles/balancing-multiplayer-games-part-2-viable-options`) — is the acceptance test for every PvP mode: asymmetric characters, factions, and loadouts are allowed; asymmetric win chance at equal skill is a defect.
- Hidden difficulty adjustment is never used in PvP or leaderboard contexts; the rule and its rationale are owned by [difficulty-and-balance.md](difficulty-and-balance.md) (### Dynamic Difficulty Policy).

### Matchmaking And Skill Spread

Defaults from Josh Menke's "Skill, Matchmaking, and Ranking Systems Design" (GDC 2016 — `gdcvault.com/play/1023014/Skill-Matchmaking-and-Ranking-Systems`, slides at `archive.org/details/GDC2016Menke`):

- **A skill rating must be predictive.** Menke defines a skill system as "any method to measure player ability" whose job is to "predict match outcomes correctly": if the system predicts a 75% win probability, the favored side should win about 75% of those matches — not 90%. Validate the rating against realized outcomes, not intuition.
- **Matchmaking optimizes retention, not rating elegance.** Menke's proxy: "hard to say what fun is, but we know what it isn't" — a large skill gap. Minimizing within-match skill gap is the operational form of the equal-chance fairness test above.
- **Queue time versus match quality is a committed tuning decision.** Decide "how long players should wait for a given drop in quality" and write the thresholds down (route through [../decisions/design-decision-records.md](../decisions/design-decision-records.md)); never let the tradeoff emerge from whatever the queue does under load.
- **Displayed rank may diverge from matchmaking rating, transparently.** Hybrid systems where visible rank and matchmaking skill diverge cause "visually strange match-ups" players will notice; Menke's mitigation is transparency — show the match's actual skill balance rather than hiding it.

### Asymmetric Multiplayer

Defaults from Behaviour Interactive's Dead by Daylight postmortem (`gamedeveloper.com/design/crafting-an-asymmetric-multiplayer-horror-experience-in-i-dead-by-daylight-i-`; GDC session: `gdcvault.com/play/1025096/-Dead-by-Daylight-Object`):

- **Commit at concept time**: "either make both sides feel similar with only minor differences, or make two completely different games." Half-committed asymmetry inherits both problem sets.
- **Each side is its own experience for its own player type.** Give each side its own fantasy, loop, onboarding ([ux-and-onboarding.md](ux-and-onboarding.md)), and playtest track — a session that only asks "was the match fun?" hides one side's misery in the other side's average.
- **Balance iteratively, arms-race style**: introduce a mechanic on one side, level the playing field, repeat. There is no mirror to check against, so win-rate parity at equal skill (the Sirlin test) is the balance target — "one of the hardest parts is trying to make each side feel fun without weakening them too much."

### Cooperation And Social Dynamics

- **The interaction verbs you ship are the behaviors you get.** Journey's playtesters "just kept attacking each other and pushing each other into the pit," so thatgamecompany "removed the physics so they couldn't push each other into the pit" (`siliconera.com/journey-producer-reveals-how-gamer-reactions-influenced-the-game/`). Design out negative-sum verbs; do not ship them and moderate the fallout.
- **Communication channels are designed per mode, not defaulted on.** Riot's cross-team chat experiment made the channel opt-in and measured "a significant decrease in all measures of toxicity" with chat usage unchanged (`gamedeveloper.com/design/gdc-riot-experimentally-investigates-online-toxicity`). Start channels restrictive and expand deliberately, with the metric that justifies each expansion named.
- **Cooperation needs a designed social surface, not proximity.** Whether a mode earns its relatedness budget — and how social motivation is profiled — is owned by [../foundations/player-psychology.md](../foundations/player-psychology.md); this doc owns the verbs and channels that surface is built from.

### Disruptive Behavior As A Design Constraint

- **Use the Fair Play Alliance / ADL framework as the shared vocabulary.** The Disruption and Harms in Online Gaming Framework (FPA + ADL Center for Technology and Society, 2020 — `thrivingingames.org/wp-content/uploads/2020/12/FPA-Framework.pdf`, `adl.org/resources/report/disruption-and-harms-online-gaming-framework`) classifies conduct along four axes: expression, delivery channel, impact, and root cause. The FPA has since folded into the Thriving in Games Group (`thrivingingames.org`).
- **Say "disruptive behavior", not "toxicity".** The framework's position: the term toxicity "fails to provide enough actionable information or useful feedback," while "disruptive behavior" forces the designer to name what is disrupted, how, and why.
- **Prevention through design beats moderation after the fact.** Lin's GDC 2013 results (priming, opt-in cross-team chat) demonstrate behavior is a tunable design output, not a community constant; the Journey physics removal is the same move at the mechanics layer.
- **Every new player-to-player channel passes the four axes at design review.** Chat, voice, emotes, trading, body-blocking, invasions: state what each surface can deliver, to whom, at what impact, before it ships — gate it through [../checklists/design-review.md](../checklists/design-review.md).

### Spectating And Streaming

Guidelines from the Chalmers spectator-interface study of Dota 2, StarCraft 2, CS:GO, and Hearthstone (Carlsson & Pelling, Report 2015:129 — `publications.lib.chalmers.se/records/fulltext/224247/224247.pdf`):

- **The spectator view is its own interface, designed on purpose.** Its most important components must be "at least discernible at all distances, screen sizes and at lower resolution streaming quality" — review it at stream resolution, not on a dev monitor.
- **Designate team colors and keep them disciplined** across world, UI, and minimap; pick colorblind-safe colors "from the start rather than implement a color blind mode later." Palette rules coordinate with [accessibility.md](accessibility.md).
- **Hide abundant information; surface it when it matters.** Timely pop-ups over always-on panels — too much information creates confusion, too little strands the viewer.
- **Make game state comprehensible at a glance.** A viewer joining mid-match must be able to judge who is leading from the interface alone (score, round, or resource indicators).
- **Preserve suspense.** Spectators may know more than players, but showing everything destroys the tension that makes the match watchable — decide the spectator's information asymmetry deliberately.
- **Budget for "assumed knowledge".** Spectator interfaces routinely presume the viewer has played the game; every assumption dropped widens the audience the stream can teach.

## Common Mistakes And Forbidden Patterns

- Shipping a PvP mode where start conditions predict wins at equal skill, with a plan to "fix it in tuning" — fairness is the acceptance test, not a polish pass.
- Hidden difficulty or rubber-band adjustment in any PvP or leaderboard context ([difficulty-and-balance.md](difficulty-and-balance.md)).
- A matchmaking rating never validated against outcomes — a system that predicts 75% and delivers 90% is a broken instrument, not a strict one.
- Letting queue-time-versus-quality be decided implicitly by population load instead of committed, recorded thresholds.
- Half-committed asymmetry: two sides neither similar enough to share a balance model nor different enough to be designed as separate experiences.
- Balancing an asymmetric mode by mirroring stats instead of testing win-rate parity at equal skill.
- Shipping grief-capable verbs (shoving, stealing, blocking) and staffing moderation to absorb the consequences instead of redesigning the verb.
- "Toxicity" as a catch-all in design docs and reviews — classify along the framework's four axes so the finding is actionable.
- Default-on, unrestricted communication channels added to a mode without design review or a measurement plan.
- Spectator UI as an afterthought: illegible at stream resolution, no team-color discipline, everything shown all the time, or an all-seeing view that kills suspense.

## Verification And Proof

- Per PvP mode, win-rate data at matched skill (playtest tallies pre-launch, telemetry post-launch via [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)) shows no start condition confers an advantage at equal skill; outliers carry an open tuning task.
- The matchmaking rating is calibration-checked on live data: predicted win probabilities are compared against realized outcomes, and drift triggers a tuning task.
- Queue policy thresholds (acceptable wait per drop in match quality) exist as a decision record ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)), and dashboards track queue time and within-match skill gap against them.
- Asymmetric modes have a separate playtest track per side ([../quality/playtesting.md](../quality/playtesting.md)), and each side's fun is reported separately in the playtest report.
- Every player-to-player channel appears in design-review notes with a four-axes pass ([../checklists/design-review.md](../checklists/design-review.md)), and each channel is instrumented (reports, mutes, usage) so a change can be evaluated the way Riot evaluated cross-team chat.
- The spectator interface is reviewed at target stream resolution, and a fresh viewer who has not played the game can state who is winning at a glance ([../quality/playtesting.md](../quality/playtesting.md)).
