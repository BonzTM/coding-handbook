# Checklist: Release

Use this for an application deploy or npm library publication.

## Candidate

- [ ] Release commit is approved, green, and contains the intended version and lockfile.
- [ ] Changelog covers config, migration, contract, dependency, security, resource, and manual changes.
- [ ] Breaking changes include compatibility window and consumer/operator migration.
- [ ] `npm ci` and `npm run verify` pass on Node 24 from a clean checkout.
- [ ] Required Testcontainers, contract, accessibility, and security jobs are green.

## Artifact

- [ ] Backend runs emitted JavaScript; frontend assets come from `vite build`.
- [ ] Image or tarball contents are inspected for secrets, local files, and unintended source maps.
- [ ] Library `npm pack` output installs and typechecks in a clean consumer.
- [ ] Artifact digest maps to source revision, lockfile, build metadata, SBOM/provenance, and scans.
- [ ] Same immutable artifact will be promoted; no environment-specific rebuild occurs.

## Deployment And Recovery

- [ ] Protected release environment uses least permissions and short-lived identity.
- [ ] Migration ordering and mixed-version compatibility are proven.
- [ ] Stage/canary smoke covers readiness, critical behavior, and telemetry.
- [ ] Observation window, SLO symptoms, abort thresholds, and accountable operator are named.
- [ ] Prior artifact, config compatibility, schema/message/cache state, and irreversible effects are in rollback plan.

## Proof

- [ ] Published/deployed version and digest equal the approved candidate.
- [ ] Post-release smoke and dashboards show expected behavior and no SLO regression.
- [ ] Registry tag or deployment promotion is auditable.
- [ ] Release notes and operator communication are published through the approved channel.
