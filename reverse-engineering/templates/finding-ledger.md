# Template: Finding Ledger (`findings.md`)

Copy to the workspace root as `findings.md`. One row per conclusion. Rules live in [../foundations/agentic-re-workflow.md](../foundations/agentic-re-workflow.md#observation-inference-unknown).

```markdown
# Findings — <game> <build-id>

Target build: <build-id> (exe sha256 <sha256>)
Last reviewed: <yyyy-mm-dd> by <name>

## Findings

| ID | Claim | Type | Evidence | Confidence | Depends on | Status |
|---|---|---|---|---|---|---|
| F-001 | <one-sentence claim> | observation \| inference | <evidence_id>[, <evidence_id>] (<bundle file>) | high \| medium \| low | <F-xxx or -> | open \| verified \| superseded by F-xxx |

## Unknowns

| ID | Question | Blocking | What would settle it | REA unknown id |
|---|---|---|---|---|
| U-001 | <question> | <F-xxx or work item> | <second build \| runtime trace \| server response \| ...> | <id from record_unknown> |

## Superseded

| ID | Was | Superseded by | Why |
|---|---|---|---|
```

## Column Rules

- **Type**: `observation` only when the claim restates what a tool returned. Anything else is `inference`.
- **Evidence**: at least one REA `evidence_id` and the bundle file under `evidence/` that contains it. A screenshot or chat quote is not evidence.
- **Confidence**: `high` needs two independent evidence types (for example static + recorded capture). `low` is allowed; unmarked is not.
- **Status**: rows are never deleted. A wrong row moves to Superseded with the reason.
- Runtime-derived rows say whether the runtime was recorded (a saved capture) or live (an approved session) in the Claim text.
