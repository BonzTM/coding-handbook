# CI And Release

Reproducible verification, dependency automation, artifact, changelog, and rollback rules.

## Default Approach

Run one lockfile-honest canonical gate locally and in CI, then release the exact verified artifact.

### Canonical Gate

`npm run verify` is the project contract. Its stable fail-fast sequence covers:

| Stage | Command or behavior | Proof |
|---|---|---|
| install honesty | clean-checkout `npm ci` | manifest and lockfile agree without rewrite |
| format | `prettier --check .` | repository formatting is canonical |
| lint | `eslint . --max-warnings 0` | flat, type-aware policy passes |
| types | `tsc --noEmit` | strict compiler reports zero diagnostics |
| tests | `jest` | deterministic unit/component suites pass |
| audit | repository `npm audit` policy | no unaccepted blocking advisory |
| build | backend `tsc` or frontend `vite build` | production artifact is reproducible |

CI performs `npm ci` before invoking repository scripts; the canonical pipeline must include and report lockfile honesty even when the install cannot literally be nested inside the running npm process. A Makefile is only `verify: ; npm run verify`; it does not duplicate commands or policy.

Integration is explicit, normally `npm run test:integration`, and runs in a separate required CI job with pinned real dependencies. It does not silently depend on a developer's Docker availability during the offline inner loop.

### CI Environment

Pin Node 24, the package manager behavior, runner image, actions, PostgreSQL image, and container builder. Grant least permissions, isolate untrusted pull requests from release secrets, and use short-lived identity for publishing where supported.

Cache only content addressed by lockfile, runtime, platform, and tool version. Cache is an accelerator, never required state. Generated artifacts and builds reproduce after cache deletion.

### Dependency Automation

Use Dependabot by default on GitHub or the platform-approved equivalent. Schedule npm, action, and container-base updates. Group low-risk compatible changes; isolate majors and security changes for focused review.

Every dependency PR reviews manifest, lockfile, transitive graph, install scripts, license, maintenance, advisories, Node/ESM compatibility, and changelog. It runs the same gate as human changes. Auto-merge requires explicit repository policy and never bypasses audit or ownership.

### Release Versioning

Applications tag immutable releases and images with a semantic version or organization-approved release identifier plus source revision. Published libraries follow semantic versioning; incompatible exports, declarations, runtime behavior, or supported-platform changes require a major version.

Build once after verification, attach provenance/SBOM where required, sign according to platform policy, and promote by digest. Do not rebuild separately for each environment.

### Changelog And Release Notes

Maintain `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) form with `Unreleased`, then `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, and `Security` sections as applicable.

Operator-visible changes require entries: configuration, secrets/rotation, ports, migrations, resource requirements, contracts, message replay, timeouts, defaults, and manual steps. Breaking changes include consumer migration and compatibility window.

### Release Pipeline

Release only from an approved green commit. Verify version consistency, build immutable artifacts, scan, publish, deploy to a controlled stage, run smoke/contract checks, then promote with observation and abort criteria.

Published libraries run `npm pack`, inspect included files, and install the tarball into a clean consumer before publish. Protect provenance by using trusted publishing or short-lived scoped credentials.

### Rollback And Recovery

Every release identifies the prior artifact digest and compatibility of config, database schema, messages, caches, and external effects. Prefer rollback of code or forward repair according to the migration plan; never assume a destructive down migration is safe.

After rollback, verify SLO symptoms, data consistency, queue/backlog, migrations, and version identity. Record incident and follow-up when release automation or gates failed to prevent impact.

### Release Security

Release jobs are protected environments with required review, minimal write permissions, immutable tags where supported, and no secret exposure to forked code. Audit publish and deploy actions.

An urgent vulnerability release preserves reproducibility, review, and artifact traceability. Coordinated disclosure notes publish only after affected artifacts and mitigations are available.

## Common Mistakes And Forbidden Patterns

- `npm install` in CI, missing lockfile, or dependency state rewritten during verification.
- Separate CI commands drifting from `npm run verify` or a Makefile duplicating policy.
- Jest passing while the production ESM artifact is never built or started.
- Integration tests skipped indefinitely because the offline gate is green.
- Floating actions, Node, database, or base image versions.
- Dependency bot changes merged without lockfile and advisory review.
- Environment-specific rebuilds, mutable tags, or release from an unverified commit.
- Operator-visible changes missing from changelog and rollback plan.
- Publishing credentials exposed to untrusted pull requests.

## Verification And Proof

- A clean checkout on Node 24 completes the canonical install, format, lint, type, test, audit, and build sequence.
- CI invokes repository scripts and the Makefile remains a one-line shim.
- Required Testcontainers integration runs against pinned PostgreSQL.
- Cache deletion does not change generated output, tests, build, or packed library contents.
- Release artifact digest maps to source revision, lockfile, build metadata, provenance, and scan result.
- Library tarball installs and runs from a clean consumer fixture before publish.
- Changelog covers every operator-visible and breaking change.
- Rollback drill or staged proof covers artifact, configuration, schema, messages, caches, and validation of recovery.

Related: [deployment.md](deployment.md), [security.md](security.md), and [../foundations/git-workflow.md](../foundations/git-workflow.md).
