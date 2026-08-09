# Runbook: <SERVICE_NAME>

Purpose: <USER_VISIBLE_PROMISE>.

## Ownership And Identity

- Repository: <REPOSITORY_URL>
- Primary/on-call: <OWNER_AND_ROTATION>
- Escalation: <ORDERED_CONTACTS>
- Environments: <ENVIRONMENTS>
- Running version/digest: <HOW_TO_QUERY>

## SLOs And Signals

| SLI | Target/window | Dashboard | Alert |
|---|---|---|---|
| <USER_VISIBLE_INDICATOR> | <TARGET_AND_WINDOW> | <LINK> | <LINK> |

- Error-budget policy: <OWNER_AND_ACTION>
- Logs/traces/metrics: <SAFE_SAVED_QUERIES>

## Alerts

### <ALERT_NAME>

- Symptom and impact: <CONDITION_AND_USER_EFFECT>
- Safe diagnosis: <BOUNDED_READ_ONLY_STEPS>
- Mitigation: <REVERSIBLE_ACTION>
- Verify: <SLI_AND_DATA_CHECK>
- Escalate when: <CONDITION>

## Deploy And Rollback

```text
<DEPLOY_COMMAND_OR_PIPELINE>
<ROLLBACK_COMMAND_OR_PIPELINE>
```

- Migration order: <EXPLICIT_JOB_AND_COMPATIBILITY>
- Observation/abort: <WINDOW_AND_THRESHOLDS>
- Previous artifact: <HOW_TO_IDENTIFY>
- Schema/message/cache/config caveats: <CAVEATS>

## Dependencies And Failure Modes

| Dependency | Timeout/bound | Failure symptom | Safe mitigation |
|---|---|---|---|
| <DEPENDENCY> | <TIMEOUT_AND_CAPACITY> | <SYMPTOM> | <MITIGATION> |

## Configuration And Secrets

- Config source and keys: `.env.example` and <DEPLOY_CONFIG_LINK>.
- Secrets/access: <STORE_PATH_AND_GRANT_OWNER>.
- Rotation: <PROCEDURE_AND_OWNER>.

## Recovery And Replay

- Restore/RPO/RTO: <BACKUP_LOCATION_LAST_RESTORE_AND_OBJECTIVES>
- Queue/DLQ/replay: <BOUNDED_IDEMPOTENT_PROCEDURE>
- Data correction: <APPROVAL_AND_VERIFICATION>
- Incident process: <LINK>
