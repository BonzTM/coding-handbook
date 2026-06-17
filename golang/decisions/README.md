# Decisions

The rules for making and recording hard-to-reverse choices: what process a non-obvious decision goes through, and the default dependency and framework picks that keep agents from inventing new architecture each task.

This directory holds the handbook-level decision *process and defaults*. Project-level ADRs — the actual records of decisions a specific service or library made — live in **that project's own `decisions/` directory**, not here. This directory governs how those records are written and which defaults they may safely assume.

## Contents

- [architecture-decision-records.md](architecture-decision-records.md) - the ADR process and when a decision requires one before it merges. Use it whenever a change is non-obvious or hard to reverse.
- [framework-selection.md](framework-selection.md) - the default dependency and framework choices, and the bar a new dependency must clear. Use it before adding any third-party library.

## Recording A Decision

Project ADRs use [../templates/adr-template.md](../templates/adr-template.md). Copy it into the project's `decisions/` directory, fill in status, alternatives, and consequences, and follow the process in [architecture-decision-records.md](architecture-decision-records.md).

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Routing a decision-shaped change: [../maintainer-map.md](../maintainer-map.md)
- The ADR file to copy: [../templates/adr-template.md](../templates/adr-template.md)
