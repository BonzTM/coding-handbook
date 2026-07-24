# Templates

Fill-in skeletons for the design artifacts this handbook requires, so every handbook-following project produces the same document shapes instead of re-deriving them per team. Each template is a fixed structure with explicit `<PLACEHOLDER>` tokens, not finished prose: copy it to the destination in the table below, fill every placeholder, and the result is governed by the linked handbook doc.

For routing a design change to the template it consumes, start at the Change Routing table in [../AGENTS.md](../AGENTS.md). Worked exemplar projects under `reference/` are a planned later phase; until they land, start every new artifact from this tree.

## Template Index

| Template | Destination in a project | Governing handbook doc |
|---|---|---|
| [one-page-gdd.md](one-page-gdd.md) | `design/<system>-one-pager.md`; the game identity page is `design/game-one-pager.md` | [foundations/design-documentation.md](../foundations/design-documentation.md); the identity page is owned by [foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) |
| [playtest-script.md](playtest-script.md) | `design/playtests/<YYYY-MM-DD>-<focus>-script.md` | [quality/playtesting.md](../quality/playtesting.md) |
| [playtest-report.md](playtest-report.md) | `design/playtests/<YYYY-MM-DD>-<focus>-report.md` | [quality/playtesting.md](../quality/playtesting.md) |
| [balance-spreadsheet-spec.md](balance-spreadsheet-spec.md) | the balance workbook itself (shared spreadsheet or committed data files) — the spec defines its tab, column, and formula shape | [disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md), [disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) |
| [ddr-template.md](ddr-template.md) | `design/decisions/NNNN-<slug>.md` | [decisions/design-decision-records.md](../decisions/design-decision-records.md) |

Each template is consumed by a recipe: [../recipes/write-a-one-page-gdd.md](../recipes/write-a-one-page-gdd.md) fills the one-pager, [../recipes/run-a-playtest.md](../recipes/run-a-playtest.md) fills the script before the session and the report after it, and [../recipes/balance-an-economy.md](../recipes/balance-an-economy.md) builds the workbook from the spreadsheet spec. DDRs are filed per the process in [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

## Filename And Destination Conventions

Unlike code-handbook template trees, every template here is markdown and its filename names the artifact type, not an encoded path — destinations contain names only you can choose (`<system>`, `<focus>`, `NNNN-<slug>`), so the destination column carries the path and the filename stays stable and greppable.

- The repo-root `design/` and `design/playtests/` destinations are a convention this handbook adopts for repo-hosted design docs; the `design/decisions/` destination is mandated by [../decisions/design-decision-records.md](../decisions/design-decision-records.md) and sits beside, not inside, the engineering `decisions/` ADR directory. Teams working on a wiki keep the same page names and hierarchy; the tool changes, the shape does not.
- `<YYYY-MM-DD>` is the session date; `<focus>` is the playtest's stated question in slug form. A script and its report share the same date-focus stem so they pair by filename.
- `NNNN` is zero-padded, sequential, and never reused, including for superseded or rejected DDRs.
- Every template carries placeholders for a named owner, a last-updated date, and a sync trigger — the three liveness markers [../foundations/design-documentation.md](../foundations/design-documentation.md) requires. Filling a template without them produces a dead document; never delete those placeholders.
- A filled [one-page-gdd.md](one-page-gdd.md) must still be one page. If filling it overflows the page, the system needs splitting, not a second page ([../foundations/design-documentation.md](../foundations/design-documentation.md) (### One-Page Designs)).
- Tuning values never land in the markdown artifacts; they live in the workbook the [balance-spreadsheet-spec.md](balance-spreadsheet-spec.md) defines, and the one-pagers link to it.

## Governing Docs

Templates carry structure; rules live in the governing docs. A template never restates its governing doc's rules — it encodes them as sections and placeholders. When a template and its governing doc disagree, the doc wins and the template is a bug.

Changing a template is a contract change: the governing doc, the recipe that consumes the template, and the Change Routing row in [../AGENTS.md](../AGENTS.md) move in the same change, per [../CONTRIBUTING.md](../CONTRIBUTING.md). Existing filled artifacts in projects are not retrofitted; they are updated the next time their sync trigger fires.

## Where To Go Next

- Routing a change to its template and obligations: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- The documentation contract these artifacts implement: [../foundations/design-documentation.md](../foundations/design-documentation.md)
- Recipes that fill these templates: [../recipes/README.md](../recipes/README.md)
- Lifecycle gates that check the filled artifacts: [../checklists/README.md](../checklists/README.md)
