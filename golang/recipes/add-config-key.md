# Recipe: Add Config Key

Use this when a feature needs a new runtime setting (a tuning knob, a feature flag, a timeout, or a secret) that the process reads at startup.

## Files To Touch

- `internal/config/config.go` (typed struct field + load in `Load` + check in `Validate`)
- `internal/config/config_test.go` (happy path + a missing/malformed failure case)
- `.env.example` (the key, with a safe example value)
- the project README config-key table (same key, same description)

## Steps

1. Add the typed field to the right struct in `internal/config`, following the governing [foundations/configuration.md](../foundations/configuration.md) and the [reference config package](../reference/exampleservice/internal/config/config.go). Use `time.Duration` for timeouts/intervals, `int`/`int64` for counts/sizes, a named enum type for closed sets, never a bare `string` for something with a fixed range. Group it into the relevant nested struct (`HTTPConfig`, `DatabaseConfig`, etc.) if one fits.
2. Add a named `default<Field>` constant alongside the others so defaults stay a single reviewable source of truth, not inline literals. Skip the default only for genuinely required values (including secrets) that must have no fallback.
3. Load it in ONE place inside `Load`: seed a flag default from the environment via the `envReader` (`env.string` / `env.int` / `env.duration` / `env.bool`), then assign the parsed flag into the `Config` struct. Do not add an `os.Getenv` lookup anywhere else. Malformed values are recorded by the `envReader` and aborted by the existing `env.err()` check, named by key, before flags parse — do not silently coerce to the default.
4. VALIDATE it in `Validate()` following the reference fail-fast pattern (see `Validate` in [config.go](../reference/exampleservice/internal/config/config.go)): required values must be non-empty/non-zero; ranges and ordering invariants (`MaxIdle <= MaxOpen`) and enum membership are checked here. Each failure returns an actionable message naming the key (e.g. `config: WORKER_CONCURRENCY must be positive, got 0`). Never silently correct a bad value.
5. Separate secrets from tuning knobs. A secret has no default, is validated as present-and-non-empty at startup, and its VALUE is never logged or echoed — only its key name on failure. Define its provenance and rotation per [operations/security.md](../operations/security.md) (### Secrets) and route the specific secrets manager to [decisions/framework-selection.md](../decisions/framework-selection.md).
6. Document it in `.env.example` AND the README config-key table in the same change, with a safe, realistic example value. The two must describe the same key identically.

## Invariants To Preserve

- no lazy or scattered `os.Getenv` lookups in handlers, repositories, or workers — config is read once in `Load` and threaded explicitly
- required values (and all secrets) fail fast at startup with a message naming the key; no silent fallback on missing or malformed input
- `.env.example` and the README config-key table stay in sync — same keys, same descriptions
- secrets carry no default and their values never reach a log line, error message, or panic dump
- no config reads in `init()` and no package-level config globals

## Proof

- a config test for the happy path (defaults applied, precedence holds): `go test ./internal/config/ -run TestLoad`
- a failure test asserting a malformed or missing-required value is rejected with the key-named message, mirroring the reference cases `TestLoadInvalidLogLevel` / `TestLoadMalformedEnvRejected` / `TestValidate`: `go test ./internal/config/ -run 'TestValidate|TestLoad.*(Invalid|Malformed)'`
- `make verify`
- startup smoke: launch the service with the required key unset and confirm it aborts before opening listeners with an actionable, key-named message. With the reference layout: `DB_MAX_OPEN_CONNS=0 go run ./cmd/exampleservice` exits non-zero and prints `config: DB_MAX_OPEN_CONNS must be positive, got 0` (substitute your new required key)
