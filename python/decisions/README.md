# Decisions

The rules for making and recording hard-to-reverse choices: what process a non-obvious decision follows, and the default package and framework picks that keep agents from inventing new architecture each task.

This directory holds handbook-level decision process and defaults. Project ADRs — records for a specific service or library — live in that project's own `decisions/` directory, not here.

## Contents

- [architecture-decision-records.md](architecture-decision-records.md) - the ADR process and when a decision requires one before merge.
- [framework-selection.md](framework-selection.md) - default package/framework choices and the bar a dependency must clear.

## Recording A Decision

Copy [../templates/adr-template.md](../templates/adr-template.md) into the project's `decisions/` directory, fill status, alternatives, consequences, and re-evaluation triggers, then follow [architecture-decision-records.md](architecture-decision-records.md).

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Decision-shaped change routing: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- ADR skeleton: [../templates/adr-template.md](../templates/adr-template.md)
