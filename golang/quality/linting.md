# Linting

The single source of truth for static linting policy: what runs, why, and how it is pinned and invoked.

## Default Approach

`golangci-lint` is the mandated meta-linter. It aggregates `gofmt`, `go vet`, `staticcheck`, and a curated set of correctness, error-handling, security, and concurrency analyzers behind one config and one command. The committed config is [../templates/.golangci.yml](../templates/.golangci.yml); copy it to the repo root as `.golangci.yml`.

This doc owns lint policy. The `gofmt -s`, `go vet`, and `staticcheck` mentions in [../foundations/style-and-review.md](../foundations/style-and-review.md), [../operations/ci-and-release.md](../operations/ci-and-release.md), and [../AGENTS.md](../AGENTS.md) describe the underlying tools and the baseline proof commands; `golangci-lint` is how the repo runs them together. Nothing here weakens those: `gofmt -s` is still mandatory, `go vet ./...` is still a standalone gate, and the linter only adds to that floor.

### Versioning And Pinning

The config targets golangci-lint **v2** (`version: "2"` schema; current stable is v2.12.2). The v2 schema is not compatible with v1 configs: linters split into a `linters:` block and a `formatters:` block, and the file must declare `version: "2"`. Do not paste v1 examples into this config.

Pin the linter as a module tool so every developer and CI run uses the same version. The v2 module path includes `/v2`:

```bash
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
```

This records a `tool` directive in `go.mod` (the reason for the `go 1.24` baseline). Never use the obsolete `tools.go` / `//go:build tools` pattern. Run it through the toolchain, never a globally installed binary that drifts between machines.

### Running It

Use the Makefile, which both humans and CI invoke:

```bash
make lint   # -> go tool golangci-lint run
make fmt    # applies the formatters golangci-lint owns (gofumpt, gci)
make verify # the full ordered gate, which includes lint
```

`go tool golangci-lint run` is the raw equivalent. The formatter half (`gofumpt`, `gci`) is applied in place by `make fmt`; the linter half is a read-only gate in `make lint`. Keep formatting out of the lint gate so a CI lint failure always means a real defect, not whitespace.

### Enabled-Linter Policy And Why

The policy is opinionated but low-noise: start from the upstream default set, then add high-signal linters whose findings are almost always real defects. We do not enable everything (`default: all`); whole-program style linters such as `wsl`, `nlreturn`, `varnamelen`, and `exhaustruct` produce churn without catching bugs, and a linter that cries wolf gets disabled wholesale. Each category earns its place:

- **Correctness** — `govet` (with `enable-all`, minus `fieldalignment`), `staticcheck`, `ineffassign`, `unused`, `unconvert`, `wastedassign`, `unparam`. These catch dead assignments, unreachable code, impossible type assertions, and unused results — defects that compile cleanly but behave wrong.
- **Error handling** — `errcheck` (with `check-type-assertions` and `check-blank`), `errorlint`, `nilerr`, `nilnesserr`. These enforce the repo's error contract from [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md): no silently dropped errors, `%w` wrapping with `errors.Is`/`errors.As` instead of `==`, and no returning `nil` right after checking a non-nil error.
- **Security** — `gosec`. Flags weak crypto, unchecked file paths, command injection, and SQL built by string concatenation — the supply-chain and input-handling risks from [../operations/security.md](../operations/security.md). `G104` is excluded because it duplicates `errcheck`; double-reporting trains people to ignore findings.
- **Concurrency and resource safety** — `bodyclose`, `rowserrcheck`, `sqlclosecheck`, `noctx`. Leaked HTTP/SQL bodies, unchecked `Rows.Err()`, and context-less network calls are the classic leaks behind exhausted connection pools and untimed-out requests; these align with [../foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md) and [../services/database.md](../services/database.md).
- **Loop and iteration safety** — `copyloopvar`. Catches loop-variable capture in closures and goroutines. (It replaces the v1 `exportloopref`; the Go 1.22 loop-var change removed most cases but not closure capture.)
- **Idiom and style** — `revive` (the maintained replacement for the retired `golint`), `gocritic`, `usestdlibvars`, `perfsprint`, `prealloc`, `testifylint`. These keep code matching [../foundations/style-and-review.md](../foundations/style-and-review.md): exported-symbol docs, lower-case unpunctuated error strings, stdlib constants over magic values.

### Formatter Policy

Formatting is owned by the `formatters:` block, not the linters, and is applied by `make fmt`:

- `gofumpt` — a strict superset of `gofmt -s`. The `gofmt -s` contract from [../foundations/style-and-review.md](../foundations/style-and-review.md) still holds; `gofumpt` only tightens it (no behavior `gofmt` rejects is allowed by `gofumpt`).
- `gci` — deterministic import grouping: standard, third-party, then the local module prefix. This replaces ad-hoc `goimports` ordering with a single enforced order.

Set the module path in both `gofumpt.module-path` and the `gci` `prefix(...)` section to the real module before committing. Editors may run `goimports` on save, but the committed contract is `gofumpt` + `gci` as applied by `make fmt`.

## Common Mistakes And Forbidden Patterns

- Using a globally installed `golangci-lint` binary instead of the pinned `go tool` version, so local and CI results diverge.
- Pasting a v1 config (no `version:` key, formatters listed under `linters`) into the v2 schema.
- Enabling `default: all` and then drowning real findings in style noise, or littering the code with `//nolint` to silence it.
- Putting formatters in the lint gate, so a CI lint failure might just be whitespace.
- Blanket-disabling a linter to dodge one finding instead of fixing the finding or scoping a `//nolint:linter // reason` with a justification.
- Excluding `_test.go` from everything; tests still need correctness and error checks, only the noisy security/dup checks are relaxed.
- Re-adding the obsolete `tools.go` pattern instead of `go get -tool`.

## Verification And Proof

```bash
make fmt    # apply gofumpt + gci
make lint   # go tool golangci-lint run -- must exit 0
```

Linting is clean when `make lint` exits 0 with no `//nolint` added to dodge a real defect, and `make fmt` leaves no diff. `make verify` runs lint as part of the full ordered gate, so a green `verify` proves the lint policy held.

## Where To Go Next

- Error contract the error linters enforce: [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)
- Style rules the idiom linters enforce: [../foundations/style-and-review.md](../foundations/style-and-review.md)
- How lint fits the delivery pipeline: [../operations/ci-and-release.md](../operations/ci-and-release.md)
- Test proof beyond static analysis: [testing.md](testing.md)
