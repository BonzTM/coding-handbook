<!--
Playtest report template. Copy to the PROJECT repo as design/playtests/<YYYY-MM-DD>-<focus>-report.md.
One report per session, written the same day while observations are fresh.
The session it reports on follows [playtest-script.md](playtest-script.md); process rules live in
[../quality/playtesting.md](../quality/playtesting.md) and [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md).
Delete this comment block when installing.
-->

# Playtest Report: <FOCUS> — YYYY-MM-DD

- **Build:** `<commit / tag / build id>`
- **Script:** link to the filled copy of [playtest-script.md](playtest-script.md) used for this session.
- **Facilitator / observers:** <NAMES>
- **Experience goals under test:** <GOALS> <!-- Copy verbatim from the one-page GDD ([one-page-gdd.md](one-page-gdd.md)). A session without a stated goal produces anecdotes, not findings — Fullerton's playcentric process tests against experience goals, never "see if they like it" (`gamedesignworkshop.com`). -->

## Session Summary

<!-- Three to five sentences: what was tested, what happened, the headline finding.
     End with one verdict line per experience goal: met / partially met / not met, with the finding IDs that justify it. -->

- Goal 1: <VERDICT> (F<N>, F<N>)
- Goal 2: <VERDICT> (F<N>)

## Participants

<!-- Anonymize: IDs only, no names in the committed report. Track first-exposure status — a tester who
     has seen the build can never be a first-time tester again (Will Wright's "Kleenex testing";
     `masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`).
     Onboarding and clarity findings from repeat testers are marked as weak evidence in Findings below. -->

| ID | Profile (target audience fit) | First-time tester? | Prior sessions | Notes |
|----|-------------------------------|--------------------|----------------|-------|
| P1 | <PROFILE> | yes/no | 0 | |

## Findings By Severity

<!-- Severity scale (this handbook's convention, defined in [../quality/playtesting.md](../quality/playtesting.md)):
     Blocker  — prevents testing or invalidates a design pillar; fix before the next session.
     Major    — an experience goal failed for multiple participants.
     Minor    — friction observed but the goal still landed.
     Observation — pattern worth watching; not yet actionable.
     One block per finding, ordered Blocker -> Observation. A finding records what participants DID,
     not what they (or you) think should change — observation over interrogation (Fullerton).
     Findings with no Evidence ID are opinions; delete them or demote to Observation. -->

### F1: <SHORT FINDING TITLE>

- **Severity:** Blocker | Major | Minor | Observation
- **Observed:** <behavior, verbatim — what participants did or failed to do, where, at what point in the session>
- **Participants affected:** <n> of <N> (P1, P3)
- **Evidence:** E1, E2
- **Hypothesis:** <usability failure (signs, feedback, clarity, cognitive load) or engage-ability failure (motivation, flow) — Hodent's split in *The Gamer's Brain* (`thegamersbrain.com`); name which, do not blend>
- **Goal or pillar hit:** <which experience goal or design pillar this threatens>

## Evidence

<!-- Verbatim quotes, video/recording timestamps, observer notes, and telemetry counts. Do not paraphrase
     a quote into your preferred conclusion. Telemetry says what players did, never why — pair every metric
     with an observed behavior or quote before it backs a finding
     ([../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)). -->

| ID | Type (quote / timestamp / observer note / metric) | Content or link | Participant |
|----|---------------------------------------------------|-----------------|-------------|
| E1 | quote | "<VERBATIM>" | P1 |

## Recommended Changes

<!-- Changes are proposals, not decisions. Route anything that moves a pillar, loop, or system contract
     through a DDR ([ddr-template.md](ddr-template.md)); route pure number changes through the tuning
     sheet ([balance-spreadsheet-spec.md](balance-spreadsheet-spec.md)). Every change names the findings
     it addresses; a change with no finding behind it does not belong in this report. -->

| Change | Addresses | Route (DDR / tuning sheet / direct fix) | Owner |
|--------|-----------|-----------------------------------------|-------|
| <CHANGE> | F1 | <ROUTE> | <NAME> |

## Follow-Up Tests

<!-- What the next session must re-test to confirm each change landed, and with whom. If a change touches
     onboarding, first-session clarity, or tutorialization, the retest requires fresh first-time testers —
     everyone in this session is now burned for that purpose. Gate the next session on
     [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md). -->

- Re-test <FINDING/CHANGE> with <fresh | returning> testers; reuse script section <SECTION>.
- <NEXT>
