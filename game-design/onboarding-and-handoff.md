# Onboarding And Handoff

> **Team-process document.** This governs design ownership transfer between people. It is not part of the design-production contract; agents drafting or changing design docs do not read it.

This guide is for taking over as design owner of a game project that was run with this handbook. It defines what a new owner reads on day one, the questions they must be able to answer before they truly own the design, and what the outgoing owner is responsible for.

This is not the handbook's own "Start Here" in [README.md](README.md). That section is about working *inside the handbook*. This guide is about owning a *design that was built with it*. If you are an agent or human contributing a single design change, you want [AGENTS.md](AGENTS.md) (including its Change Routing table), not this file.

## Who This Is For

- A new design owner inheriting a game in prototyping, production, or live operation.
- The outgoing owner running the transfer.
- A producer or lead confirming a handoff is actually complete before sign-off.

If the project follows the handbook, every artifact this guide references already exists in the project. If one is missing, that is a handoff defect, not an optional extra; surface it before accepting ownership.

## Day-One Reading Path, In Order

Read these in the project, not in the handbook. Do not skip ahead; each step assumes the previous one.

| Step | Read | What you must come away knowing |
|---|---|---|
| 1 | The project's one-page GDD | What the game is, who it is for, and the target experience in one page. Its shape is governed by [templates/one-page-gdd.md](templates/one-page-gdd.md). |
| 2 | The pillars document | The three-to-five pillars every feature is judged against, and at least one thing that was cut because of them. See [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md). |
| 3 | The core loop diagram | The repeatable action sequence at each timescale — moment, session, meta — and which loop is the game. See [foundations/core-loops.md](foundations/core-loops.md). |
| 4 | The project's DDRs | Why the load-bearing design choices were made, what was rejected, and which decisions are still open. Read newest and any proposed-but-unimplemented records first. Format: [decisions/design-decision-records.md](decisions/design-decision-records.md). |
| 5 | The balance sheets | Where every tuning value lives and how each is derived, not just its current number. Shape: [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md). |
| 6 | The last three playtest reports | What players actually did versus what the design intended, and what changed as a result. Shape: [templates/playtest-report.md](templates/playtest-report.md). |
| 7 | Play the current build | The design as it exists, not as documented. Play through the full core loop yourself before you change anything. |

Step 7 is the gate between reading and owning. If you cannot get the current build running and play its core loop end to end, you do not yet have a working picture of the design and the handoff is not done. Docs describe intent; the build is the ground truth, and the two drift — the playcentric process in [quality/playtesting.md](quality/playtesting.md) exists to close exactly that gap.

## Questions A New Owner Must Be Able To Answer

You do not own the design until you can answer every one of these unaided. Treat any "I'd have to ask the previous owner" as an open handoff item.

### Vision And Loops

- What are the pillars, and what recent feature was cut or reshaped because it failed one? (Pillars doc; see [foundations/design-pillars-and-vision.md](foundations/design-pillars-and-vision.md).)
- What is the core loop at each timescale, and which mechanic is it built to teach mastery of? (Loop diagram; see [foundations/core-loops.md](foundations/core-loops.md).)
- Which parts of the game are loops and which are arcs — what is replayable versus consumed once — and does the production budget reflect that split? (See [foundations/core-loops.md](foundations/core-loops.md).)

### Balance And Economy

- Where does every tuning value live, and can I trace any one of them back to its cost-curve or economy derivation? (Balance sheets; see [disciplines/difficulty-and-balance.md](disciplines/difficulty-and-balance.md).)
- What are the economy's faucets and sinks, and which flow is being watched for inflation or starvation right now? (See [disciplines/economy-and-progression.md](disciplines/economy-and-progression.md).)
- Which options are currently dominant or dead, and what is the open tuning plan for them?

### Playtesting And Evidence

- When was the last playtest, what did it find, and what changed because of it? (Playtest reports; see [quality/playtesting.md](quality/playtesting.md).)
- Where do fresh first-time testers come from, and how is the pool protected? A tester who has seen the game can never be a first-time tester again, so the recruiting pipeline is an owned asset, not an errand. (See [quality/playtesting.md](quality/playtesting.md).)
- Which telemetry dashboard answers which design question, and which live values ship via remote config versus a client build? (See [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md).)

### Decisions And Direction

- Why are the load-bearing design choices the way they are? (Project DDRs; see [decisions/design-decision-records.md](decisions/design-decision-records.md).)
- Which decisions are still open or proposed, and what evidence is blocking them?
- What is in scope for the current milestone, what was cut to get there, and what does the next milestone have to prove? (See [operations/scoping-and-production.md](operations/scoping-and-production.md).)

If a question has no documented answer, the fix is to write it down in the project, not to keep it in your head. Undocumented design intent is the failure this guide exists to prevent.

## Outgoing Owner Responsibilities

The outgoing owner runs the handoff and is responsible for it being complete; do not delegate it to the newcomer to discover.

- Walk the incoming owner through the pillars and at least one real pillar-driven cut, so the pillars transfer as a working tool, not a poster.
- Hand over balance sheets with derivations intact. A tuning value whose rationale lives only in your head is tribal knowledge; restore the derivation per [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md) before you leave.
- Surface every open or proposed DDR and every undocumented decision; convert tribal knowledge into a DDR before you leave. Format: [templates/ddr-template.md](templates/ddr-template.md).
- Transfer the playtest operation: past reports, the session script in use ([templates/playtest-script.md](templates/playtest-script.md)), the fresh-tester recruiting pipeline, and any scheduled sessions.
- Transfer ownership of telemetry dashboards, remote-config values, and any running A/B tests, and confirm the new owner can read and change them independently. (See [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md).)
- Confirm the design docs match the current build: the one-page GDD, loop diagram, and balance sheets describe today's game, not the game as pitched. Stale docs are worse than missing ones because they are trusted.
- Hand off the stakeholder map: who approves scope changes, who owns the milestone gates, and who the new owner escalates to when design and production conflict.

The transfer is complete only when the new owner can run a playtest independently — per [recipes/run-a-playtest.md](recipes/run-a-playtest.md) and [checklists/playtest-readiness.md](checklists/playtest-readiness.md) — and can answer the day-one questions without you.

## Where To Go Next

- The fast-path contract and its Change Routing table: [AGENTS.md](AGENTS.md)
- Why decisions are recorded and how: [decisions/design-decision-records.md](decisions/design-decision-records.md)
- Running and reporting a playtest: [recipes/run-a-playtest.md](recipes/run-a-playtest.md)
- Tuning values and their derivations: [templates/balance-spreadsheet-spec.md](templates/balance-spreadsheet-spec.md)
- Live values, dashboards, and experiments: [operations/live-tuning-and-telemetry.md](operations/live-tuning-and-telemetry.md)
- Milestones and scope control: [operations/scoping-and-production.md](operations/scoping-and-production.md)
