# Release Checklist

Release checklist for .NET artifacts that should be traceable, reproducible enough, and safe to deploy.

## Source And CI

- [ ] `dotnet restore --locked-mode` succeeds on the release commit — every `packages.lock.json` is in sync with `Directory.Packages.props`, with no drift.
- [ ] Required CI stages passed on the release commit (the full `pwsh ./verify.ps1` matrix: restore (locked), format-check, build (warnings-as-errors), test, audit).
- [ ] `dotnet list package --vulnerable --include-transitive` is clean, or each finding has a documented, justified exception.
- [ ] Migrations, compatibility notes, and operator-facing config changes are documented.

## Artifact Quality

- [ ] Build is `Release` configuration with `ContinuousIntegrationBuild=true` in CI, so the artifact is deterministic and embeds no local filesystem paths.
- [ ] Binaries or containers report correct version/build info: `InformationalVersion` is stamped from the tag and the image label matches it.
- [ ] VCS release tags use the canonical `v1.2.3` form; the NuGet package version is `1.2.3` (package versions never carry the `v`), and a changelog or display string may drop the `v`, but the tag must not.
- [ ] If ReadyToRun or Native AOT is enabled (ADR-gated per [../decisions/framework-selection.md](../decisions/framework-selection.md)), it was validated on the target runtime image and the trade-off is still current.
- [ ] Container images are built with the reviewed multi-stage flow (SDK build stage, chiseled non-root ASP.NET runtime stage) from the committed [Dockerfile](../templates/Dockerfile) when containers are part of the release.

## Deploy Safety

- [ ] Readiness and shutdown behavior were smoke-tested on the release artifact: `/readyz` gates traffic and `SIGTERM` drains in-flight work within `HostOptions.ShutdownTimeout`.
- [ ] Schema or API changes have a backward-compatibility or rollback story; the migration step is explicit (`--migrate` flag or init job), never auto-run on normal startup per [../services/database.md](../services/database.md).
- [ ] Release notes mention new env vars, migration requirements, port changes, or dependency changes that matter to operators.
- [ ] Event or message contract changes have a compatibility, replay, and DLQ or rollback story that operators understand.

## Verification

```powershell
pwsh ./verify.ps1                                        # restore (locked), format-check, build (warnings-as-errors), test, audit
dotnet restore --locked-mode                             # lock files in sync on the release commit
dotnet list package --vulnerable --include-transitive    # audit posture is current
git describe --tags --exact-match                        # the release commit carries the canonical v1.2.3 tag
```
