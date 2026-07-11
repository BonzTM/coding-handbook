# Templates

Committed, copy-paste-ready starting artifacts for a new C#/.NET repository, so every handbook-following repo converges on the same scaffolding instead of re-deriving it per project.

This tree is the artifact home that the prose docs keep implying. When [foundations/project-setup.md](../foundations/project-setup.md) says "thin `Program.cs`" or [checklists/new-project.md](../checklists/new-project.md) says "baseline CI runs `pwsh ./verify.ps1`", the concrete file lives here. Copy a template into the destination its filename encodes, fill in the placeholders, and the result is governed by the linked handbook doc.

Exact version pins (SDK line, package versions, action majors, base images) live ONLY in these templates and the `reference/` modules — prose docs never pin. The three compiling, `verify.ps1`-green reference modules ([exampleservice](../reference/exampleservice/), [examplegrpc](../reference/examplegrpc/), [exampleworker](../reference/exampleworker/)) carry these same files as proven copies; when bootstrapping a new repo, copy from the reference module matching your shape and fall back to this tree for the pieces it lacks.

## How To Use

1. Create the repo skeleton per [checklists/new-project.md](../checklists/new-project.md).
2. Copy each template to the destination path in the table below, dropping the `project-` filename prefix and decoding `-` segments into directories where noted.
3. Replace every `<placeholder>` token — and the `Orders`/`orders` example names in the Dockerfile, compose file, and k8s manifests with your app name. Markdown templates are fill-in skeletons; `global.json`, `Directory.Build.props`, `Directory.Packages.props`, `nuget.config`, `.editorconfig`, `.gitattributes`, `verify.ps1`, and the `Makefile` shim work as-is once copied.
4. Run `pwsh ./verify.ps1` before the first commit.

## Filename Convention

A template's filename encodes its destination in a fresh repo. Slashes in a destination become `-` in the template filename so the whole tree stays flat and greppable:

- `program-main.cs.txt` -> `src/<App>.Api/Program.cs` (C# source templates carry a trailing `.txt` so this docs repo holds no compilable `.cs` files; drop the `.txt` when you copy)
- `gitignore` -> `.gitignore` (shipped without the leading dot so this docs repo does not behave like an app repo; add the dot when you copy)
- `project-readme.md` -> `README.md`
- `project-agents.md` -> `AGENTS.md`
- `codeowners.md` -> `.github/CODEOWNERS`
- `github-workflows-ci.yml` -> `.github/workflows/ci.yml`
- `k8s-deployment.yaml` -> `k8s/deployment.yaml`
- `docker-compose.yml` keeps its literal name at the repo root.
- dotfiles and root configs (`global.json`, `Directory.Build.props`, `Directory.Packages.props`, `nuget.config`, `verify.ps1`, `Makefile`, `Dockerfile`, `.editorconfig`, `.gitattributes`, `.dockerignore`) keep their literal name.

Where the destination contains a name you choose (`<App>`), pick it when you copy; the template does not guess for you.

## Template Index

| Template | Destination in a new repo | Governing handbook doc |
|---|---|---|
| [global.json](global.json) | `global.json` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [Directory.Build.props](Directory.Build.props) | `Directory.Build.props` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [Directory.Packages.props](Directory.Packages.props) | `Directory.Packages.props` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [nuget.config](nuget.config) | `nuget.config` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [.editorconfig](.editorconfig) | `.editorconfig` | [quality/linting.md](../quality/linting.md) |
| [.gitattributes](.gitattributes) | `.gitattributes` | [foundations/cross-platform.md](../foundations/cross-platform.md) |
| [gitignore](gitignore) | `.gitignore` | [foundations/git-workflow.md](../foundations/git-workflow.md) |
| [verify.ps1](verify.ps1) | `verify.ps1` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [Makefile](Makefile) | `Makefile` (one-line shim delegating to `pwsh ./verify.ps1`) | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [program-main.cs.txt](program-main.cs.txt) | `src/<App>.Api/Program.cs` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [github-workflows-ci.yml](github-workflows-ci.yml) | `.github/workflows/ci.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [github-workflows-release.yml](github-workflows-release.yml) | `.github/workflows/release.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [Dockerfile](Dockerfile) | `Dockerfile` | [operations/deployment.md](../operations/deployment.md) |
| [.dockerignore](.dockerignore) | `.dockerignore` | [operations/deployment.md](../operations/deployment.md) |
| [docker-compose.yml](docker-compose.yml) | `docker-compose.yml` | [operations/deployment.md](../operations/deployment.md) |
| [k8s-deployment.yaml](k8s-deployment.yaml) | `k8s/deployment.yaml` | [operations/deployment.md](../operations/deployment.md) |
| [dependabot.yml](dependabot.yml) | `.github/dependabot.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [project-readme.md](project-readme.md) | `README.md` | [README.md](../README.md) |
| [project-agents.md](project-agents.md) | `AGENTS.md` | [AGENTS.md](../AGENTS.md) |
| [project-contributing.md](project-contributing.md) | `CONTRIBUTING.md` | [foundations/git-workflow.md](../foundations/git-workflow.md) |
| [codeowners.md](codeowners.md) | `.github/CODEOWNERS` | [AGENTS.md](../AGENTS.md) (## Change Routing) |
| [adr-template.md](adr-template.md) | `decisions/NNNN-<slug>.md` | [decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) |
| [changelog.md](changelog.md) | `CHANGELOG.md` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [runbook.md](runbook.md) | `docs/runbook.md` | [operations/operability.md](../operations/operability.md) |
| [security-policy.md](security-policy.md) | `.github/SECURITY.md` | [operations/security.md](../operations/security.md) |
| [pull_request_template.md](pull_request_template.md) | `.github/pull_request_template.md` | [checklists/pr-review.md](../checklists/pr-review.md) |

What each foundational file does, in one line:

- `global.json` pins the SDK line (`rollForward: latestFeature`) so every machine and CI leg builds with the same toolchain.
- `Directory.Build.props` sets the solution-wide build contract once: nullable, implicit usings, `TreatWarningsAsErrors`, `AnalysisLevel=latest-all`, lock files.
- `Directory.Packages.props` is the single place package VERSIONS live (Central Package Management); projects reference packages by name only.
- `nuget.config` pins the package feeds and keeps restore deterministic.
- `.editorconfig` is the single style/severity source, enforced by `dotnet format` and the build.
- `.gitattributes` normalizes line endings (`eol=lf` for source) so the three-OS matrix agrees on file content.
- `verify.ps1` is the ONE canonical gate — restore (locked), format-check, build (warnings-as-errors), test, audit — for humans and CI alike; the `Makefile` exists only so `make verify` habits still land on it.
- `program-main.cs.txt` is the thin composition root: config, DI, telemetry, endpoints, graceful shutdown — no business logic.

`pwsh ./verify.ps1` is the single verification entrypoint for both humans and CI: docs and workflows invoke the script (and its `-Integration` switch) rather than restating raw command lists.

## Where To Go Next

- Bootstrapping a repo: [checklists/new-project.md](../checklists/new-project.md)
- The layout these templates land in: [foundations/project-setup.md](../foundations/project-setup.md)
