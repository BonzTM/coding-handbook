# New Project Checklist

Bootstrap checklist for a brand-new Go repo using this handbook.

## Repository Skeleton

- [ ] Run `go mod init <module-path>` in the repo root.
- [ ] Create `cmd/<app>/main.go` and keep it limited to startup, wiring, and shutdown.
- [ ] Create `internal/config`, `internal/core`, `internal/telemetry`, and any needed transport or storage packages.
- [ ] Add tool-only dependencies (such as `govulncheck`, `staticcheck`, `sqlc`, `golangci-lint`, or migration tooling) with `go get -tool <pkg>` (Go 1.24+); they are recorded as `tool` directives in `go.mod` and run with `go tool <name>`. Do not use the legacy `tools.go` blank-import pattern.
- [ ] Decide whether the repo is a service, worker, CLI, library, or a combination, then document that shape in the repo README.
- [ ] Decide which boundaries need explicit contracts in `api/`, transport docs, or schema sources.
- [ ] If the repo publishes or consumes messages, decide event envelope shape, idempotency policy, ordering guarantees, retry limits, and DLQ behavior up front.

## Runtime Contract

- [ ] Configuration loads in one place under `internal/config`.
- [ ] `log/slog` is wired centrally and injected into runtime components.
- [ ] `signal.NotifyContext` drives the process root context.
- [ ] Health and readiness behavior is defined for networked services.
- [ ] Database and external clients have explicit timeout and shutdown behavior.

## Proof And Delivery

- [ ] Copy the committed [Makefile](../templates/Makefile), [.golangci.yml](../templates/.golangci.yml), and [CI workflow](../templates/github-workflows-ci.yml) into the repo so the gate is identical locally and in CI.
- [ ] `make verify` is THE baseline gate from day one — it runs tidy, fmt-check, `golangci-lint` (per [../quality/linting.md](../quality/linting.md)), `go vet`, test, race, `govulncheck`, and build. This is mandatory, not an optional `staticcheck` add-on.
- [ ] CI runs `make verify` on every push and pull request, with no green build possible while it fails.
- [ ] A coverage stance is set per [../quality/testing.md](../quality/testing.md): mandatory paths (domain core, error/status mapping, decode paths) are covered and coverage is tracked rather than allowed to silently regress.
- [ ] `.env.example` exists if the repo has env-driven config.
- [ ] Initial docs link to `golang/AGENTS.md` or the repo's own equivalent fast-path contract.

## Verification

```bash
make verify   # tidy, fmt-check, lint, vet, test, race, vuln, build
```
