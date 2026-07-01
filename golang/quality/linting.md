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
make lint      # -> go tool golangci-lint run
make fmt       # applies the formatters golangci-lint owns (gofumpt, gci)
make fmt-check # read-only formatter gate: fails if `make fmt` would produce a diff
make verify    # the full ordered gate, which includes fmt-check and lint
```

`go tool golangci-lint run` is the raw equivalent. The formatter half (`gofumpt`, `gci`) is applied in place by `make fmt`; its read-only CI counterpart is `make fmt-check` (`golangci-lint fmt --diff`), which fails without writing anything. The linter half is a read-only gate in `make lint`. Keep formatting out of the lint gate so a CI lint failure always means a real defect, not whitespace; `verify` runs `fmt-check` as its own step for the same reason.

### Enabled-Linter Policy And Why

The policy is opinionated but low-noise: start from the upstream default set, then add high-signal linters whose findings are almost always real defects. We do not enable everything (`default: all`); a linter that cries wolf gets disabled wholesale. Each category earns its place:

- **Correctness** — `govet` (with `enable-all`, minus `fieldalignment`), `staticcheck`, `ineffassign`, `unused`, `unconvert`, `wastedassign`, `unparam`, `exhaustive`. These catch dead assignments, unreachable code, impossible type assertions, unused results, and enum switches that silently skip a value — defects that compile cleanly but behave wrong. `exhaustive` runs with `default-signifies-exhaustive: true`: a `default:` branch is an explicit decision that covers the remaining values, so only switches with neither full coverage nor a `default:` are flagged.
- **Error handling and logging** — `errcheck` (with `check-type-assertions` and `check-blank`), `errorlint`, `nilerr`, `nilnesserr`, `sloglint`. These enforce the repo's error and logging contract from [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md): no silently dropped errors, `%w` wrapping with `errors.Is`/`errors.As` instead of `==`, no returning `nil` right after checking a non-nil error, and no `log/slog` call that mixes key-value arguments with `slog.Attr` (`no-mixed-args`) — mixed calls make it easy to miscount the key-value pairs, which shows up only as `!BADKEY` attributes at runtime. `sloglint` does not mandate one attribute style beyond that, because the logging doc does not either.
- **Security** — `gosec`. Flags weak crypto, unchecked file paths, command injection, and SQL built by string concatenation — the supply-chain and input-handling risks from [../operations/security.md](../operations/security.md). `G104` is excluded because it duplicates `errcheck`; double-reporting trains people to ignore findings.
- **Concurrency and resource safety** — `bodyclose`, `rowserrcheck`, `sqlclosecheck`, `noctx`, `fatcontext`, `spancheck`. Leaked HTTP/SQL bodies, unchecked `Rows.Err()`, context-less network calls, contexts rebuilt inside loops (an unbounded value-chain leak), and OTel spans that are never ended or never record errors are the classic leaks behind exhausted pools, untimed-out requests, and broken traces; these align with [../foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md), [../services/database.md](../services/database.md), and [../operations/observability.md](../operations/observability.md). Like `testifylint`, `spancheck` is a no-op in repos that never import the OTel SDK, so it costs nothing where it does not apply.
- **Loop and iteration safety** — `copyloopvar`, `intrange`. `copyloopvar` catches loop-variable capture in closures and goroutines (it replaces the v1 `exportloopref`; the Go 1.22 loop-var change removed most cases but not closure capture). `intrange` rewrites manual counted loops to the Go 1.22 `for i := range n` form — fewer off-by-one opportunities.
- **Idiom and style** — `revive` (the maintained replacement for the retired `golint`), `gocritic`, `usestdlibvars`, `perfsprint`, `prealloc`, `misspell`, `modernize`, `recvcheck`, `gochecknoinits`, `testifylint`, `protogetter`. These keep code matching [../foundations/style-and-review.md](../foundations/style-and-review.md): exported-symbol docs, lower-case unpunctuated error strings, stdlib constants over magic values. `misspell` catches common misspellings in comments and strings before reviewers do. `modernize` flags legacy patterns with direct modern replacements (`slices.Contains` over hand-rolled loops, `sync.WaitGroup.Go`, `errors.AsType`) so the codebase tracks the stdlib the `go` directive already requires. `recvcheck` forbids mixing value and pointer receivers on one type, which the style doc calls a smell. `gochecknoinits` bans `init()` functions outright — the style doc hard-forbids `init()` that wires, configures, or registers, and generated code (the one legitimate `init()` user) is already excluded via `generated: strict`. `protogetter` enforces getters over direct field access on generated protobuf structs (nil-message safety); like `testifylint` and `spancheck`, it is a no-op without generated protobuf code in the repo.
- **Test hygiene** — `usetesting`, `thelper`, `tparallel`. `usetesting` replaces raw `os`/`context` calls in tests with `t.TempDir()`, `t.Setenv()`, and `t.Context()`, which clean up and cancel automatically per [testing.md](testing.md). `thelper` requires `t.Helper()` in test helpers so failures point at the calling test, not the helper. `tparallel` catches inconsistent `t.Parallel()` between parent tests and subtests — misuse that silently serializes or races.
- **Suppression hygiene** — `nolintlint` (with `require-explanation` and `require-specific`). Every `//nolint` must name the linter it silences and carry a reason, so suppressions stay auditable; see [Suppressing A False Positive](#suppressing-a-false-positive).

### Considered And Excluded

Deliberate exclusions, so nobody re-litigates them one PR at a time. Whole-program style linters such as `wsl`, `nlreturn`, `varnamelen`, and `exhaustruct` produce churn without catching bugs. Beyond those:

- `wrapcheck` — demands wrapping at every return; [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md) mandates judgment-based wrapping at package/subsystem boundaries, not mechanically at every stack frame.
- `contextcheck` — false-positives on deliberately detached contexts (a bounded `context.Background()` for shutdown flush and terminal-settle paths), which the reference apps use on purpose.
- `mnd` — magic-number detection flags far more honest literals (ports, sizes, test fixtures) than real magic; `usestdlibvars` already covers the stdlib-constant cases.
- `nilnil` — returning `(nil, nil)` is a legitimate "not found, no error" contract in places; `nilerr`/`nilnesserr` already catch the actual bug class.
- `paralleltest` — flags every test that lacks `t.Parallel()`, but [testing.md](testing.md) forbids parallelizing tests that mutate process-global state; `tparallel` catches the real misuse without demanding blanket parallelism.
- `depguard` / `gomodguard_v2` — import/module allow-deny lists are per-repo policy with no meaningful handbook-wide default; configure them in the repo that needs them. (Plain `gomodguard` is deprecated; use `gomodguard_v2` if you adopt it.)
- `forbidigo` — useful as a per-repo tool (e.g. banning `time.Now` in core packages per [../foundations/time.md](../foundations/time.md)), but the ban list is repo-specific, so the shared config ships none.
- `gochecknoglobals` — would flag the package-level sentinel errors and immutable lookup tables that [../foundations/style-and-review.md](../foundations/style-and-review.md) explicitly permits; the style doc's review rules police mutable globals better than a blanket linter.

### Changing The Enabled Set

The template is the source of truth and the copies must not drift: any change to the enabled linters or their settings updates [../templates/.golangci.yml](../templates/.golangci.yml), all three reference-app copies (`reference/exampleservice`, `reference/examplegrpc`, `reference/exampleworker` — identical to the template except the module-path and `prefix(...)` substitutions), and the rationale in this document — one line per enable, one line per deliberate exclusion — in the same PR. A PR that changes one of those without the others is incomplete; `make verify` must pass in all three reference apps before it merges.

### Suppressing A False Positive

When a finding is genuinely wrong (not merely inconvenient), suppress it at the narrowest possible scope:

```go
data, err := os.ReadFile(path) //nolint:gosec // path is validated against the allowlist above
```

- Same-line `//nolint:<linter> // <reason>` only. The linter name is mandatory (never bare `//nolint`), and the reason after `//` is mandatory — `nolintlint` rejects both omissions.
- Never file-wide or package-wide suppressions; a directive above a function is the widest acceptable scope, and only when every line in it trips the same false positive.
- If the same false positive recurs across files, the linter is misconfigured for this codebase: fix the setting in the template (see [Changing The Enabled Set](#changing-the-enabled-set)) instead of scattering `//nolint`.

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
