<!--
This file is the handbook template for a repo's PR template.
Install it at: .github/pull_request_template.md
It mirrors checklists/pr-review.md — that checklist is the source of truth for review depth.
Delete this comment block when installing.
-->

## Summary

<!-- What changed and why, in two or three sentences. Link the issue/ticket. -->

-

## Change Type

<!-- Check all that apply. -->

- [ ] Feature
- [ ] Bug fix
- [ ] Refactor (no behavior change)
- [ ] Performance
- [ ] Docs / tooling only
- [ ] Dependency bump
- [ ] Breaking change (export, wire, schema, config, or event contract)

## Gates

- [ ] `make verify` is green locally (tidy, fmt-check, lint, vet, test, race, vuln, build).
- [ ] Tests added or updated that prove the actual behavior change; if this ships a feature it meets every gate in [feature-definition-of-done.md](../checklists/feature-definition-of-done.md).
- [ ] Docs and `.env.example` updated for any new or changed config key.
- [ ] `CHANGELOG.md` entry added for every operator-visible change (behavior, config, migration, contract).
- [ ] ADR linked below if this change is architectural or shifts a contract (per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- [ ] Compatibility / migration impact noted below if exports, wire, schema, or event payloads changed.

## Compatibility / Migration

<!-- Required if "Breaking change" is checked or any contract moved. State the rollback story
     and any migration step operators must run. "None" is a valid answer for non-breaking changes. -->

None.

## ADR

<!-- Link the ADR for architectural changes, or "N/A". -->

N/A
