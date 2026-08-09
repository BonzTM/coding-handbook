# Templates

Committed, copy-paste-ready starting artifacts for a new Python repository, so every handbook-following repo converges on the same scaffolding instead of re-deriving it per project.

This tree is the artifact home that the prose docs govern. Copy each template to the destination its filename encodes, replace every `<placeholder>`, generate `uv.lock`, and run the canonical gate.

## How To Use

1. Create the repo skeleton per [the new-project checklist](../checklists/new-project.md).
2. Copy the applicable files to the destinations below. Decode filename prefixes and add leading dots where documented.
3. Replace every `<placeholder>`; remove optional runtime dependencies and artifact jobs that the project shape does not use.
4. Run `uv lock`, commit `uv.lock`, then run `make verify` before the first commit.

Exact Python, uv, package, action, and image pins live in templates. The canonical `pyproject.toml` uses `uv_build` because the default is one pure-Python package under `src/`; Hatchling requires the escalation in [framework selection](../decisions/framework-selection.md).

## Filename Convention

Filenames encode destinations and keep this documentation tree non-executable:

- `src-app-main.py.txt` -> `src/<app>/__main__.py` and `src-app-config.py.txt` -> `src/<app>/config.py`; Python source templates carry `.txt`, which is dropped when copied.
- `gitignore` -> `.gitignore`, `editorconfig` -> `.editorconfig`, `python-version` -> `.python-version`, and `env.example` -> `.env.example`; these dotfile templates omit the leading dot in this repo.
- `project-readme.md` -> `README.md`, `project-agents.md` -> `AGENTS.md`, and `project-contributing.md` -> `CONTRIBUTING.md`.
- `codeowners.md` -> `.github/CODEOWNERS`, `pull_request_template.md` -> `.github/pull_request_template.md`, and workflow filenames decode into `.github/workflows/`.
- `k8s-deployment.yaml` -> `k8s/deployment.yaml`; literal root files keep their names.

## Template Index

| Template | Destination in a new repo | Governing handbook doc |
|---|---|---|
| [pyproject.toml](pyproject.toml) | `pyproject.toml` | [project setup](../foundations/project-setup.md), [linting](../quality/linting.md) |
| [Makefile](Makefile) | `Makefile` | [CI and release](../operations/ci-and-release.md) |
| [python-version](python-version) | `.python-version` | [project setup](../foundations/project-setup.md) |
| [gitignore](gitignore) | `.gitignore` | [git workflow](../foundations/git-workflow.md) |
| [editorconfig](editorconfig) | `.editorconfig` | [style and review](../foundations/style-and-review.md) |
| [env.example](env.example) | `.env.example` | [configuration](../foundations/configuration.md) |
| [src-app-main.py.txt](src-app-main.py.txt) | `src/<app>/__main__.py` | [project setup](../foundations/project-setup.md) |
| [src-app-config.py.txt](src-app-config.py.txt) | `src/<app>/config.py` | [configuration](../foundations/configuration.md) |
| [Dockerfile](Dockerfile) | `Dockerfile` | [deployment](../operations/deployment.md) |
| [.dockerignore](.dockerignore) | `.dockerignore` | [deployment](../operations/deployment.md) |
| [docker-compose.yml](docker-compose.yml) | `docker-compose.yml` | [deployment](../operations/deployment.md) |
| [k8s-deployment.yaml](k8s-deployment.yaml) | `k8s/deployment.yaml` | [deployment](../operations/deployment.md) |
| [github-workflows-ci.yml](github-workflows-ci.yml) | `.github/workflows/ci.yml` | [CI and release](../operations/ci-and-release.md) |
| [github-workflows-release.yml](github-workflows-release.yml) | `.github/workflows/release.yml` | [CI and release](../operations/ci-and-release.md) |
| [dependabot.yml](dependabot.yml) | `.github/dependabot.yml` | [CI and release](../operations/ci-and-release.md) |
| [project-readme.md](project-readme.md) | `README.md` | [Python handbook](../README.md) |
| [project-agents.md](project-agents.md) | `AGENTS.md` | [agent contract](../AGENTS.md) |
| [project-contributing.md](project-contributing.md) | `CONTRIBUTING.md` | [git workflow](../foundations/git-workflow.md) |
| [codeowners.md](codeowners.md) | `.github/CODEOWNERS` | [agent contract](../AGENTS.md) |
| [adr-template.md](adr-template.md) | `decisions/NNNN-<slug>.md` | [ADRs](../decisions/architecture-decision-records.md) |
| [changelog.md](changelog.md) | `CHANGELOG.md` | [CI and release](../operations/ci-and-release.md) |
| [runbook.md](runbook.md) | `docs/runbook.md` | [operability](../operations/operability.md) |
| [security-policy.md](security-policy.md) | `.github/SECURITY.md` | [security](../operations/security.md) |
| [pull_request_template.md](pull_request_template.md) | `.github/pull_request_template.md` | [PR review](../checklists/pr-review.md) |

The Makefile is the single verification entrypoint. `make verify` runs lock-check, frozen sync, format-check, lint, imports, types, test, and audit in order; named targets remain available for fast feedback.

## Where To Go Next

- Bootstrapping: [new-project checklist](../checklists/new-project.md)
- Layout: [project setup](../foundations/project-setup.md)
- Complete worked service: `reference/exampleservice/` lands in phase 2.
