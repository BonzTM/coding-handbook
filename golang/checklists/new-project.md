# New Project Checklist

Bootstrap checklist for a brand-new Go repo using this handbook.

## Repository Skeleton

- [ ] Run `go mod init <module-path>` in the repo root.
- [ ] Create `cmd/<app>/main.go` and keep it limited to startup, wiring, and shutdown.
- [ ] Create `internal/config`, `internal/core`, `internal/telemetry`, and any needed transport or storage packages.
- [ ] Add `tools.go` for tool-only dependencies such as `govulncheck`, `staticcheck`, `sqlc`, or migration tooling.
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

- [ ] Baseline CI runs `gofmt -s -l .`, `go vet ./...`, `go test ./...`, `go test -race ./...`, and `go build -trimpath ./...`.
- [ ] `govulncheck ./...` is part of release or CI policy.
- [ ] `.env.example` exists if the repo has env-driven config.
- [ ] Initial docs link to `golang/AGENTS.md` or the repo's own equivalent fast-path contract.

## Verification

```bash
gofmt -s -l .
go vet ./...
go test ./...
go test -race ./...
go build -trimpath ./...
```
