# Recipe: Write A One-Page GDD

Use this when a game concept, system, feature, or mode needs a spec the team will actually read: one dense, visual page built from [../templates/one-page-gdd.md](../templates/one-page-gdd.md), not a chapter of prose. Documentation practice is owned by [../foundations/design-documentation.md](../foundations/design-documentation.md); this recipe is the fast path.

## Files To Touch

- a copy of [../templates/one-page-gdd.md](../templates/one-page-gdd.md), renamed for the system and placed where the project keeps design docs
- the project's pillar statement (read, not edited — every one-pager must name the pillar it serves, per [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md))
- a DDR via [../templates/ddr-template.md](../templates/ddr-template.md) only if the page overturns a previously recorded decision

## Steps

1. Scope the page to ONE system or ONE concept. Stone Librande's rule ("One-Page Designs", GDC 2010, `gdcvault.com/play/1012356/One-Page`): documentation exists to communicate ideas efficiently, and "most people only read the first page anyway" — so each system gets its own page. If the material demands two pages, split it into two one-pagers; never shrink fonts or spill over.
2. Copy the template and fill the header first: system name, owner, date, and the design pillar this system serves. A one-pager that cannot name its pillar is describing a system the game may not need — route that back through [../checklists/concept-intake.md](../checklists/concept-intake.md).
3. Build the central visual before writing any prose. The page is organized around one load-bearing diagram — a loop diagram (shape defined in [../foundations/core-loops.md](../foundations/core-loops.md)), a state machine, a level flow, an economy graph, or an annotated mock — and the labels on that diagram ARE the spec. Prose only annotates what the visual cannot carry.
4. Compress the numbers. Tunable values go in a small table, not sentences; each value is either traced to the project's balance sheet ([../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md)) or explicitly marked provisional. No paragraph exceeds two sentences; cut adjectives, keep decisions.
5. Separate decisions from open questions in distinct blocks. A reader must be able to tell at a glance what is settled and what is still being argued. Delete every unused template section — a placeholder left in ships ambiguity.
6. Run the review pass that proves the page communicates: hand it to a fresh reader who has not seen the system (the doc-level equivalent of Kleenex testing, per [../quality/playtesting.md](../quality/playtesting.md)) and have them explain the system back without your help. Wherever their read diverges from intent, fix the page, not the reader, and re-test with another fresh reader. You are playtesting the document.
7. Keep it alive. When the system changes in a prototype or playtest, the page changes in the same pass — the industry consensus is living one-pagers over monolithic up-front GDDs, and a stale page is worse than none ([../foundations/design-documentation.md](../foundations/design-documentation.md)).

## Invariants To Preserve

- one page, one system — overflow splits into a second one-pager, never a denser page
- the central visual carries the spec; prose annotates, it does not duplicate
- the header names owner, date, and the pillar served — no orphan systems
- decisions and open questions never share a block
- every number on the page is traced or marked provisional; no confident-looking guesses
- no template placeholder survives to review

## Proof

- a fresh reader explains the system back accurately, unaided — divergences fixed on the page and re-tested
- the page passes [../checklists/design-review.md](../checklists/design-review.md)
- every tunable on the page reconciles against the balance sheet or carries a provisional mark
- if the page overturned a recorded decision, the DDR landed in the same change ([../decisions/design-decision-records.md](../decisions/design-decision-records.md))
