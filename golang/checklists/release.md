# Release Checklist

Release checklist for Go artifacts that should be traceable, reproducible enough, and safe to deploy.

## Source And CI

- [ ] `go mod tidy` and `go mod verify` are clean.
- [ ] Required CI stages passed on the release commit.
- [ ] `govulncheck ./...` and any repo-mandated static analysis checks are green.
- [ ] Migrations, compatibility notes, and operator-facing config changes are documented.

## Artifact Quality

- [ ] Build uses `-trimpath` and the intended VCS metadata policy.
- [ ] Binaries or containers report correct version/build info.
- [ ] VCS release tags use the canonical `v1.2.3` form (the `v` prefix is required for Go module versions); a changelog or display string may drop the `v`, but the tag must not.
- [ ] If `default.pgo` is used, it came from representative profiles and is still current.
- [ ] Container images are built with a reviewed multi-stage flow when containers are part of the release.

## Deploy Safety

- [ ] Readiness and shutdown behavior were smoke-tested on the release artifact.
- [ ] Schema or API changes have a backward-compatibility or rollback story.
- [ ] Release notes mention new env vars, migration requirements, port changes, or dependency changes that matter to operators.
- [ ] Event or message contract changes have a compatibility, replay, and DLQ or rollback story that operators understand.

## Verification

```bash
go mod verify
go test -race ./...
go build -trimpath ./...
go version -m <binary>
```
