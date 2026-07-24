<!--
TEMPLATE -> design/<name>-one-pager.md (one file per game concept or per system)
Fill every <PLACEHOLDER>. Delete the guidance comments before circulating.
Hard constraint: the finished document must fit on one printed/rendered page.
If a section overflows, cut detail or move it to a linked deep-dive doc — never add
a second page. The one-page discipline comes from Stone Librande's "One-Page Designs"
(GDC 2010, `gdcvault.com/play/1012356/One-Page`): design docs exist to
communicate ideas efficiently, and "most people only read the first page anyway."
Governed by [../foundations/design-documentation.md](../foundations/design-documentation.md);
the fill-in procedure is [../recipes/write-a-one-page-gdd.md](../recipes/write-a-one-page-gdd.md).
-->

# <Game Or System Name> — One-Page Design

<One sentence: what this is (game concept | system inside <game>) and its current status (concept | prototyping | in production). Owner: <name>. Last updated: <date>.>

## Hook

<One or two sentences, player-facing, concrete. What the player does and why it is
worth their next hour — not genre labels, not feature lists. If a reader cannot
retell the hook after one read, rewrite it.>

## Pillars

<!-- 3-5 maximum. Each pillar is a decision filter, not a feature: it must be able to
kill a proposed feature. Shape and rules are owned by
../foundations/design-pillars-and-vision.md — do not restate them here, just fill in. -->

1. **<Pillar name>** — <one sentence: the experience it protects and what it excludes.>
2. **<Pillar name>** — <one sentence.>
3. **<Pillar name>** — <one sentence.>

## Core Loop Diagram

<!-- Zoomed-out view of the repeatable action cycle, not a detailed spec — verb-first
player actions, with the reward of one step feeding the next. Loop construction rules
are owned by ../foundations/core-loops.md; the procedure is
../recipes/design-a-core-loop.md. Replace the sketch below with your loop; a photo of
a whiteboard drawing is acceptable, an accurate diagram is required. -->

```text
<ACTION 1: verb phrase>  -->  <ACTION 2: verb phrase>  -->  <ACTION 3: verb phrase>
        ^                                                          |
        +---------------- <REWARD / RESOURCE that re-motivates ACTION 1> ----------+
```

- **Loop frequency**: <seconds per cycle at moment-to-moment scale; minutes per session cycle.>
- **What the player is mastering**: <the skill or pattern the loop teaches; a loop with nothing left to learn goes stale (Koster, *A Theory of Fun*, `theoryoffun.com/press.shtml`).>

## Player Experience

<!-- Experience goals come first and everything else is tested against them —
Fullerton's playcentric process (*Game Design Workshop*, 5th ed.,
`routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`).
Motivation vocabulary is owned by ../foundations/player-psychology.md. -->

- **Target player**: <who, described by motivation (e.g. Quantic Foundry dimensions), not demographics.>
- **Experience goal**: <the feeling the player should report unprompted after a session; this is what playtests verify.>
- **Session shape**: <expected session length and where a session naturally ends.>
- **Reference points**: <up to two existing games and the single element borrowed from each — no "X meets Y" pitches without naming the element.>

## Scope And Risks

<!-- Milestone vocabulary (prototype -> first playable -> vertical slice) is owned by
../operations/scoping-and-production.md. A one-pager without a named riskiest
assumption is not done. -->

- **Riskiest assumption**: <the claim that, if false, kills the concept.>
- **Next proof**: <prototype | first playable | vertical slice> — <what it must demonstrate, by <date>.>
- **Team and budget**: <people x time available for the next proof.>
- **Out of scope**: <two or three tempting features this one-pager explicitly excludes.>
- **Open risks**: <ordered list, worst first; each with the mechanism that retires it (prototype, paper test, spike).>

## Success Criteria

<!-- Each criterion must be checkable by a named mechanism — a playtest observation,
a telemetry read, or a build milestone. "Players have fun" is not a criterion;
"first-time testers replay the loop unprompted" is. Playtest protocol is owned by
../quality/playtesting.md. -->

- [ ] <Criterion tied to the experience goal, verified by playtest — see [../quality/playtesting.md](../quality/playtesting.md).>
- [ ] <Criterion tied to the core loop, e.g. "fresh testers complete N loop cycles without prompting".>
- [ ] <Criterion tied to scope, e.g. "next proof ships by <date> within the stated team/budget".>
- [ ] <Kill criterion: the observable result that means stop — as concrete as the success lines.>
