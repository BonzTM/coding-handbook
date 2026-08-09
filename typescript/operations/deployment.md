# Deployment

Build, image, configuration, migration, rollout, and shutdown rules for TypeScript applications.

## Default Approach

Build immutable artifacts once, run them as non-root, and promote the same digest through environments.

### Production Artifacts

Backend production JavaScript comes from `tsc`; frontend assets come from `vite build`. Node native type stripping is development convenience only and never the deployed backend build.

Use multi-stage OCI builds. Install with `npm ci`, run the canonical verification/build stages, and copy only required runtime output and production dependencies into a pinned Debian slim or distroless runtime image.

Record source revision, version, build time, and base image provenance through OCI labels and application build info. Do not bake secrets or environment-specific config into layers.

### Runtime Hardening

Run as a numeric non-root user with a read-only root filesystem where supported, a writable bounded temp area, dropped capabilities, and no package manager or shell unless operationally required.

Set CPU, memory, process, and ephemeral-storage requests/limits from load evidence. Configure Node memory behavior to leave headroom for native allocations and platform termination.

### Configuration And Secrets

Inject validated runtime configuration and secrets through the platform. Frontend build variables are public; runtime frontend config is served through a validated public endpoint when promotion requires environment independence.

Secret rotation uses rolling restart by default. Never use Docker build arguments for secret material; layers and build metadata can retain them.

### Migrations And Rollout

Run `node-pg-migrate` as one explicit, observable deployment job before or between application rollout stages according to expand/migrate/contract compatibility. Application replicas do not race to migrate at startup.

Use rolling or canary rollout with readiness, surge/unavailable bounds, and automatic/manual rollback criteria. New and old versions must coexist safely for APIs, messages, database schema, and caches.

### Health And Shutdown

Wire startup, readiness, and liveness probes to their distinct contracts. Mark readiness false before drain. The orchestrator termination grace period exceeds the application's bounded drain plus telemetry flush budget.

Handle `SIGTERM`, stop intake, abort process-owned work, drain in-flight work, close pools/consumers/workers, flush telemetry, and exit. PID 1 must receive and forward signals correctly.

### Frontend Delivery

Serve HTML with revalidation or no-cache and fingerprinted assets as immutable with a long lifetime. Configure SPA deep-link fallback without rewriting API or asset failures into HTML success.

Set CSP, security headers, compression, and source-map publication according to security and debugging policy. Public source maps require explicit risk review; private upload to an error service is preferred when needed.

### Rollback

Rollback names the artifact digest, configuration compatibility, schema state, message compatibility, and irreversible effects. A down migration is not the default rollback. Prefer forward repair when data transformation cannot be reversed safely.

## Common Mistakes And Forbidden Patterns

- Running TypeScript source or native stripping as the production backend artifact.
- Floating base tags, mutable promoted builds, or `npm install` in image builds.
- Root containers, secrets in layers/build args, or writable application directories without need.
- Every replica running migrations at startup.
- Readiness enabled before initialization or during shutdown drain.
- Grace period shorter than the application's drain budget.
- Long-cache HTML pointing at assets removed by the next rollout.
- Rollback plan that ignores schema and message compatibility.

## Verification And Proof

- A clean multi-stage build uses the committed lockfile and produces the expected artifact.
- Image scan, SBOM/provenance, non-root, filesystem, and secret-layer checks pass.
- Container smoke test starts from emitted JavaScript, serves probes, receives `SIGTERM`, drains, and exits in budget.
- Migration applies on empty and prior schema independently of application startup.
- Mixed-version canary and rollback tests cover database, HTTP, event, and cache compatibility.
- Frontend deep links, asset cache headers, CSP, and no-secret artifact inspection pass.

Related: [../foundations/project-setup.md](../foundations/project-setup.md), [../services/database.md](../services/database.md), and [ci-and-release.md](ci-and-release.md).
