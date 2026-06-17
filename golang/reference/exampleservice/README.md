# exampleservice

A small, complete HTTP service that is the **keystone reference** for the Go
engineering handbook. It manages a "widgets" feature end to end and exists to
prove the handbook's guidance and committed templates work on real code.

It is its own Go module (`github.com/example/exampleservice`) so editor tooling
resolves the internal imports cleanly when you open it in isolation. The
language baseline is **Go 1.24+**.

> **Note:** the module's `go.mod` `go` directive is `go 1.25.0`, not `1.24`.
> Pinning `golangci-lint` as a `go tool` dependency pulls in a tool graph that
> requires Go 1.25, which raises the directive (`go mod tidy` re-asserts it).
> This is a real, instructive consequence of treating linters as versioned tool
> dependencies; the 1.24+ baseline still describes the service code itself.

## What it is

- `net/http` + Go 1.22 `ServeMux` HTTP service, no web framework.
- An in-memory store by default; a `database/sql` reference store that compiles
  with the standard library only (no driver imported) and documents driver
  wiring per the database doc.
- Structured logging with `log/slog`, a readiness flag, and a stdlib-only
  metrics seam (no Prometheus dependency).
- Typed, validated, fail-fast configuration from flags and the environment.
- A thin `main` that wires lifecycle and performs ordered, bounded shutdown.

## Run it

```bash
make run            # go run ./cmd/exampleservice (in-memory store)
# or directly:
go run ./cmd/exampleservice
```

Then:

```bash
curl -s localhost:8080/livez
curl -s localhost:8080/readyz
curl -s -XPOST localhost:8080/widgets -d '{"id":"w1","name":"Widget One"}'
curl -s localhost:8080/widgets/w1
curl -s localhost:8080/widgets
```

Configuration keys are documented in [.env.example](.env.example). Precedence is
flags > environment > defaults.

## Verify

`make verify` is the single ordered safety gate (tidy, fmt-check, lint, vet,
test, race, vuln, build); humans and CI run the same target.

```bash
make verify
```

The `lint`, `fmt`, and `vuln` targets use `go tool` directives
(`go tool golangci-lint`, `go tool govulncheck`); add the tools with
`go get -tool ...` per `golang/foundations/project-setup.md`. Pinning
`golangci-lint` this way is what raises the `go.mod` directive to 1.25 (see the
note above). The core build/test loop needs no extra tools:

```bash
go build ./...
go test ./...
go test -race ./...
```

## Package map

Each package embodies a specific handbook doc:

| Package / file | Responsibility | Governing handbook doc |
|---|---|---|
| `cmd/exampleservice/main.go` | thin main: signal context, config load, slog, wiring, errgroup, ordered bounded shutdown | `foundations/context-and-concurrency.md`, `templates/cmd-app-main.go.txt` |
| `internal/config` | env+flags load, fail-fast `Validate`, no globals/init | `foundations/configuration.md` |
| `internal/core` | widgets domain service; defines the `Store` interface it consumes (interface-at-consumer); injected `Clock` | `foundations/package-design.md`, `foundations/data-modeling.md`, `foundations/time.md` |
| `internal/db/memory.go` | in-memory `Store` test/dev double (default + tests) | `services/database.md` |
| `internal/db/postgres.go` | real `database/sql` repository; stdlib only, no driver; explicit pool sizing | `services/database.md` |
| `internal/api/http` | transport adapter: server hardening, middleware order, decode→validate→core→map→encode, DTOs, error mapping | `services/http-services.md`, `foundations/serialization.md`, `foundations/errors-and-logging.md` |
| `internal/telemetry` | `slog` logger construction, readiness flag, stdlib metrics seam (no-op + expvar) | `operations/observability.md`, `foundations/errors-and-logging.md` |
| `internal/buildinfo` | `Name`/`Version`/`Commit` stamped via `-ldflags` | `foundations/project-setup.md` |
| `Makefile`, `.golangci.yml`, `Dockerfile`, `.dockerignore`, `.env.example` | adapted copies of `golang/templates/*` with the real module path | `quality/linting.md`, `operations/deployment.md` |

## Notes on deliberate stdlib-only choices

- **Metrics**: production wires a Prometheus client per
  `operations/observability.md`. The reference uses an `expvar`-backed
  implementation behind the `telemetry.Metrics` interface so the module carries
  no external metrics dependency. Swap the adapter, not the call sites.
- **Database**: `internal/db/postgres.go` compiles against `database/sql`
  with no driver. To run it, blank-import a driver in `main` and pass its name
  to `db.OpenDB`; see the file's package comment.
- **errgroup**: the only non-stdlib dependency is `golang.org/x/sync/errgroup`,
  matching the canonical `templates/cmd-app-main.go.txt`.
