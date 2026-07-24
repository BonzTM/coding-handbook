# Design Documentation

The documentation contract: which design documents exist, what shape each takes, who owns each, and how they stay synced with the playable build.

## Default Approach

Documentation exists to communicate design decisions efficiently, not to prove work happened. The default unit is a one-page visual design per system, backed by living data documents (spreadsheets, decision records, playtest reports) that change with the build. There is no monolithic game design document: design emerges through the prototype-playtest-revise loop ([../quality/prototyping.md](../quality/prototyping.md), [../quality/playtesting.md](../quality/playtesting.md)), and documents record what the loop decided — they do not front-load decisions the loop has not tested yet. This follows Tracy Fullerton's playcentric process ([*Game Design Workshop*, 5th ed.](https://www.routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009)).

### One-Page Designs

Every system that more than one person builds against gets a one-page design. The model is Stone Librande's ["One-Page Designs" (GDC 2010)](https://gdcvault.com/play/1012356/One-Page), built from production practice on Diablo III, Spore, and SimCity: the goal of a design document is "to efficiently communicate ideas", and most readers only read the first page anyway — so make the design *be* one page, dense and visual, rather than the summary of a longer document nobody opens.

- One page, one system. A page that needs a second page is describing two systems; split it.
- Visual first: diagrams, annotated mockups, and flow arrows over paragraphs. A wall of prose on a one-pager is a monolith in disguise.
- Actionable by every discipline: an engineer can scope it, an artist can list assets from it, a tester can derive checks from it. If a discipline cannot act on the page, the page is missing a region, not an appendix.
- The game's identity one-pager is a special case owned by [design-pillars-and-vision.md](design-pillars-and-vision.md); its shape is the [../templates/one-page-gdd.md](../templates/one-page-gdd.md) template, produced via [../recipes/write-a-one-page-gdd.md](../recipes/write-a-one-page-gdd.md).

### Living Documents Over Monoliths

The industry moved from monolithic up-front GDDs to living documents — one-pagers, shared spreadsheets, decision records, and the prototype itself as the spec of record for feel. Do not overcorrect into "no documentation": teams that write nothing down cannot onboard, cannot hand off, and relitigate the same decisions every milestone. Some publishers and platform holders still require milestone documentation; treat that as an export generated from the living docs, not as the working format.

- A document is *living* when it has a named owner, a last-updated marker, and a defined sync trigger (the build event that forces an edit). A document missing any of the three is dead on arrival.
- Decisions with lasting consequences leave the one-pager and become design decision records, owned by [../decisions/design-decision-records.md](../decisions/design-decision-records.md) using [../templates/ddr-template.md](../templates/ddr-template.md). One-pagers state the current design; DDRs preserve why alternatives lost.
- Tuning values never live in prose. Numbers live in the balance workbook ([../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md)) or in the build's data files; documents link to them.
- For questions of feel, the playable build outranks any document ([game-feel.md](game-feel.md)). Documents describe intent and record decisions; they do not adjudicate how the game feels.

### Document Types And Owners

Every document type has exactly one owning role and one owning handbook doc. A design fact lives in exactly one document; everything else links to it.

| Document | Owner | Governing doc | Sync trigger |
|---|---|---|---|
| Identity one-pager (pillars, hook, audience) | Creative/design lead | [design-pillars-and-vision.md](design-pillars-and-vision.md) | Any pillar change (rare; requires a DDR) |
| System one-pagers (loop, economy, combat, ...) | The designer who owns that system | This doc; loop shape in [core-loops.md](core-loops.md), system shape in [mechanics-and-systems.md](mechanics-and-systems.md) | The system changes in the build |
| Balance workbook | Systems designer | [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md), [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) | Any tuning-value change |
| Design decision records | Whoever proposes the decision | [../decisions/design-decision-records.md](../decisions/design-decision-records.md) | A pillar-level or cross-system decision lands |
| Playtest scripts and reports | Playtest coordinator | [../quality/playtesting.md](../quality/playtesting.md), templates in [../templates/playtest-script.md](../templates/playtest-script.md) and [../templates/playtest-report.md](../templates/playtest-report.md) | Every playtest session |
| Telemetry/tuning notes | Live/systems designer | [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md) | Every live tuning change |

### Keeping Docs Synced With The Build

A design document that contradicts the build is worse than no document, because new team members trust it. Sync is enforced by process, not by goodwill:

- Every change that alters a system's player-facing behavior updates that system's one-pager in the same change, or explicitly marks the page stale with an owner and a date. The [../checklists/design-review.md](../checklists/design-review.md) checklist asks for this; a review that skips it is incomplete.
- Playtest findings that change a design flow back into the one-pager and, when pillar-level, into a DDR — the report is the evidence, the one-pager is the decision ([../quality/playtesting.md](../quality/playtesting.md)).
- Stale documents are deleted or stamped, never silently kept. An unowned document is stale by definition.
- Onboarding is the sync audit: a new team member following [../onboarding-and-handoff.md](../onboarding-and-handoff.md) reads the one-pagers, plays the build, and files every mismatch found. Mismatches are doc bugs and get fixed like bugs.

## Common Mistakes And Forbidden Patterns

- Writing a monolithic GDD up front that specifies systems no prototype has tested — it will be wrong, unread, and expensive to keep wrong.
- The opposite failure: no documentation at all, on the theory that "the build is the doc." The build shows what; it cannot record why, and handoff dies with it.
- One-pagers that are one page of prose. Librande's format is visual and spatial; a text page is a compressed monolith, not a one-page design.
- Duplicating a fact across documents (a drop rate in the one-pager, the workbook, and a wiki page). One home per fact; the rest link.
- Tuning numbers embedded in prose documents, guaranteeing they drift from the build's data files.
- Documents without an owner, a last-updated marker, or a sync trigger — vision statements and system pages that nobody is obligated to maintain.
- Recording pillar-level decisions only in chat threads or meeting memory instead of a DDR, so the team relitigates them every milestone.
- Treating a publisher milestone document as the working format instead of an export from the living docs.

## Verification And Proof

- Every system named in the build's current scope has a one-page design; every one-pager names its owner and last-updated date. Spot-check both directions: no undocumented shipped system, no documented cut system.
- Pick three one-pagers and play the build against them; every mismatch is either fixed or stamped stale-with-owner in the same pass.
- Tuning values quoted in any prose doc: count should be zero — grep the docs for magic numbers and replace them with links to the balance workbook.
- Every decision that changed a pillar or crossed system boundaries in the last milestone has a DDR; check the [../decisions/README.md](../decisions/README.md) index against the milestone's change list.
- The [../checklists/design-review.md](../checklists/design-review.md) doc-sync items pass for the most recent design review.

Related: [../recipes/write-a-one-page-gdd.md](../recipes/write-a-one-page-gdd.md) for producing the identity one-pager, [../templates/README.md](../templates/README.md) for all document skeletons, [../decisions/design-decision-records.md](../decisions/design-decision-records.md) for the decision log, and [../onboarding-and-handoff.md](../onboarding-and-handoff.md) for the handoff reading path these documents must support.
