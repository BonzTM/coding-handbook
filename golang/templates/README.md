# Templates

Committed, copy-paste-ready starting artifacts for a new Go repository, so every handbook-following repo converges on the same scaffolding instead of re-deriving it per project.

This tree is the artifact home that the prose docs keep implying. When [foundations/project-setup.md](../foundations/project-setup.md) says "thin `main`" or [checklists/new-project.md](../checklists/new-project.md) says "baseline CI runs `make verify`", the concrete file lives here. Copy a template into the destination its filename encodes, fill in the placeholders, and the result is governed by the linked handbook doc.

## How To Use

1. Create the repo skeleton per [checklists/new-project.md](../checklists/new-project.md).
2. Copy each template to the destination path in the table below, dropping the `cmd-app-` / `project-` filename prefix and decoding `-` segments into directories where noted.
3. Replace every `<placeholder>` token — and in Go templates the `github.com/org/repo` module path and the stubbed `internal/...` package APIs, and in `.golangci.yml` the module path. Markdown templates are fill-in skeletons; the `Makefile` works as-is once copied.
4. Run `make verify` before the first commit.

## Filename Convention

A template's filename encodes its destination in a fresh repo. Slashes in a destination become `-` in the template filename so the whole tree stays flat and greppable:

- `cmd-app-main.go.txt` -> `cmd/<app>/main.go` (Go source templates carry a trailing `.txt` so this docs repo holds no unbuildable `.go` files; drop the `.txt` when you copy)
- `project-readme.md` -> `README.md`
- `project-agents.md` -> `AGENTS.md`
- `codeowners.md` -> `.github/CODEOWNERS`
- `github-workflows-ci.yml` -> `.github/workflows/ci.yml`
- `k8s-deployment.yaml` -> `k8s/deployment.yaml`
- `docker-compose.yml` keeps its literal name at the repo root.
- dotfiles and root configs (`Makefile`, `Dockerfile`, `.golangci.yml`, `.dockerignore`) keep their literal name.

Where the destination contains a name you choose (`<app>`), pick it when you copy; the template does not guess for you.

## Template Index

| Template | Destination in a new repo | Governing handbook doc |
|---|---|---|
| [Makefile](Makefile) | `Makefile` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [.golangci.yml](.golangci.yml) | `.golangci.yml` | [quality/linting.md](../quality/linting.md) |
| [cmd-app-main.go.txt](cmd-app-main.go.txt) | `cmd/<app>/main.go` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [project-readme.md](project-readme.md) | `README.md` | [README.md](../README.md) |
| [project-agents.md](project-agents.md) | `AGENTS.md` | [AGENTS.md](../AGENTS.md) |
| [codeowners.md](codeowners.md) | `.github/CODEOWNERS` | [AGENTS.md](../AGENTS.md) (## Change Routing) |
| [adr-template.md](adr-template.md) | `docs/adr/NNNN-<slug>.md` | [decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) |
| [github-workflows-ci.yml](github-workflows-ci.yml) | `.github/workflows/ci.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [github-workflows-release.yml](github-workflows-release.yml) | `.github/workflows/release.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [Dockerfile](Dockerfile) | `Dockerfile` | [operations/deployment.md](../operations/deployment.md) |
| [.dockerignore](.dockerignore) | `.dockerignore` | [operations/deployment.md](../operations/deployment.md) |
| [docker-compose.yml](docker-compose.yml) | `docker-compose.yml` | [operations/deployment.md](../operations/deployment.md) |
| [k8s-deployment.yaml](k8s-deployment.yaml) | `k8s/deployment.yaml` | [operations/deployment.md](../operations/deployment.md) |
| [runbook.md](runbook.md) | `docs/runbook.md` | [operations/operability.md](../operations/operability.md) |
| [dependabot.yml](dependabot.yml) | `.github/dependabot.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [security-policy.md](security-policy.md) | `.github/SECURITY.md` | [operations/security.md](../operations/security.md) |
| [pull_request_template.md](pull_request_template.md) | `.github/pull_request_template.md` | [checklists/pr-review.md](../checklists/pr-review.md) |
| [project-contributing.md](project-contributing.md) | `CONTRIBUTING.md` | [foundations/git-workflow.md](../foundations/git-workflow.md) |
| [changelog.md](changelog.md) | `CHANGELOG.md` | [operations/ci-and-release.md](../operations/ci-and-release.md) |

The `Makefile` is the single verification entrypoint for both humans and CI: `make verify` runs the full ordered safety gate, and the other targets (`help`, `tidy`, `fmt`, `fmt-check`, `lint`, `vet`, `test`, `race`, `cover`, `vuln`, `build`) are the named steps. Docs and CI invoke `make` targets rather than restating raw command lists.

## The Reference Service

A complete, compiling, `make verify`-green example lives at [../reference/exampleservice/](../reference/exampleservice/). It composes the build-and-verify templates above (`Makefile`, `Dockerfile`, `.dockerignore`, `.golangci.yml`) with the handbook's code patterns, proving they hold together in a building repo; the repo-governance artifacts (README/AGENTS/CODEOWNERS skeletons, workflows, k8s manifests, check-list docs) are deliberately not duplicated there — copy those from this tree. The reference is also the canonical, proven source for the artifacts that are project-shaped rather than copy-verbatim — copy them from the reference instead of from a static snippet so they cannot drift out of a building state:

- `internal/config` loader (env load, fail-fast validation, no globals) — [foundations/configuration.md](../foundations/configuration.md)
- `internal/buildinfo` (version/commit stamping via `-ldflags`) — [operations/observability.md](../operations/observability.md)
- `.env.example`, `.gitignore`, `.editorconfig` — proven copies in the reference root
- `sqlc.yaml` (module root) and the goose `migrations/` embedded via `go:embed` in `internal/db/` — [services/database.md](../services/database.md)

## Where To Go Next

- Bootstrapping a repo: [checklists/new-project.md](../checklists/new-project.md)
- The layout these templates land in: [foundations/project-setup.md](../foundations/project-setup.md)
