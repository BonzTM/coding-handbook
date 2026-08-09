# Configuration

Typed startup configuration and feature-flag rules for predictable TypeScript processes.

## Default Approach

Read external configuration once, parse it with Zod, and inject one immutable typed object from composition.

### Sources And Precedence

Define precedence explicitly, normally command-line override where supported, process environment, local development file, then safe code default. Production secrets come from the deployment platform, not committed files.

Keep environment variable names uppercase and service-scoped where collision is possible. Commit `.env.example` with descriptions and safe placeholders. Never commit `.env` or production-like credentials.

### Parsing And Normalization

Treat `process.env` as `unknown` string input. Parse booleans, integers, durations, URLs, lists, and enums explicitly. Reject partial parses, empty required values, out-of-range numbers, unsafe destinations, and ambiguous booleans.

Apply defaults in the schema or one mapper. Report all invalid key names at startup when practical, but never print values for secret keys. Freeze or expose the parsed object as readonly.

Parse strings deliberately and return one typed value:

```ts
import { z } from "zod";

const portSchema = z
  .string()
  .regex(/^\d+$/)
  .transform((value) => Number(value))
  .pipe(z.int().min(1).max(65_535));

const envSchema = z.strictObject({
  NODE_ENV: z.enum(["development", "test", "production"]),
  PORT: portSchema.default("3000"),
  DATABASE_URL: z.url(),
  LOG_LEVEL: z.enum(["debug", "info", "warn", "error"]).default("info"),
  FEATURE_WIDGET_V2: z
    .enum(["true", "false"])
    .default("false")
    .transform((value) => value === "true"),
});

export type Config = Readonly<z.infer<typeof envSchema>>;

export function loadConfig(env: NodeJS.ProcessEnv): Config {
  const ownedInput = {
    NODE_ENV: env.NODE_ENV,
    PORT: env.PORT,
    DATABASE_URL: env.DATABASE_URL,
    LOG_LEVEL: env.LOG_LEVEL,
    FEATURE_WIDGET_V2: env.FEATURE_WIDGET_V2,
  };
  return Object.freeze(envSchema.parse(ownedInput));
}
```

`z.strictObject` also detects misspelled keys when the input is a purpose-built config object. If the deployment platform injects unrelated environment keys, select only the owned names before strict parsing rather than switching to an unbounded loose schema.

### Ownership And Access

Only `src/config` reads environment state. Core modules never read `process.env`. Composition passes narrow configuration slices to the adapters that own them.

Configuration is startup state by default. A value that must change without restart needs an explicit dynamic-configuration design covering validation, atomic replacement, failure fallback, telemetry, and operator control.

Create configuration once in composition and pass narrow slices inward:

```ts
type ServerConfig = Pick<Config, "PORT">;

async function main(env: NodeJS.ProcessEnv): Promise<void> {
  const config = loadConfig(env);
  await startServer({ config: { PORT: config.PORT } });
}

void main(process.env).catch(() => {
  console.error("startup configuration failed");
  process.exitCode = 1;
});
```

Production startup uses the configured redacting logger and reports Zod issue paths, not raw input. The example keeps the parse call inside `main`, not at module import.

### Secrets

Secrets have provenance, access policy, and rotation procedure. Prefer mounted files or runtime injection from the platform. Validate presence without logging values. Rolling restart is the default refresh mechanism unless live reload is an accepted requirement.

### Feature Flags

Feature flags are temporary operational controls, not permanent architecture.

- Every flag has an owner, purpose, default, creation date, expiry or removal condition, and removal issue.
- Use safe, deterministic defaults; absence must not accidentally enable risk.
- Evaluate at one owned boundary and pass the resulting decision inward.
- Define targeting inputs and privacy constraints; do not send sensitive attributes to a flag vendor without approval.
- Test enabled, disabled, missing-provider, and stale-configuration behavior.
- Emit low-cardinality evaluation or rollout telemetry; never put user IDs in metric attributes.
- Remove the losing branch, config, tests, and telemetry when rollout is complete.

Flags do not bypass authorization, validation, schema compatibility, or database migration safety. Kill switches fail toward the documented safe behavior when the provider is unavailable.

### Frontend Configuration

Vite-exposed variables are public and embedded at build time. Never place secrets in `VITE_*` values. Validate the public configuration before application render and distinguish build-time values from runtime-served configuration.

## Common Mistakes And Forbidden Patterns

- Reading `process.env` throughout the codebase or during module import.
- JavaScript truthiness used to parse boolean strings.
- Secret values in startup errors, logs, snapshots, source maps, or frontend bundles.
- Environment-specific behavior encoded as scattered `NODE_ENV` branches.
- Mutable configuration singleton changed by tests.
- A permanent feature flag with no owner or removal condition.
- A flag used to hide an incompatible database or message change.

## Verification And Proof

- Tests cover valid, missing, empty, malformed, defaulted, and out-of-range inputs.
- Startup fails before accepting work and identifies invalid key names safely.
- Search confirms only the configuration module reads environment state.
- `.env.example`, deployment values, runbook, and schema use the same names.
- Both feature-flag branches and provider-failure behavior are tested.
- A secret rotation exercise proves the documented restart or reload path.
- Frontend artifact inspection shows no server secret or unintended environment value.

Related: [project-setup.md](project-setup.md), [../operations/deployment.md](../operations/deployment.md), and [../operations/security.md](../operations/security.md).
