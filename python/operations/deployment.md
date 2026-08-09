# Deployment

Container and rollout defaults for a reproducible, non-root Python service with bounded lifecycle behavior.

## Default Approach

Build one application environment in a pinned multi-stage image, copy only the non-editable virtual environment into a slim runtime, and run one Uvicorn process as a non-root user. The committed [Dockerfile](../templates/Dockerfile), [.dockerignore](../templates/.dockerignore), and [Kubernetes manifest](../templates/k8s-deployment.yaml) are the canonical starting points; exact image and tool pins live there, not in prose.

### Multi-Stage Uv Build

Copy the `uv` binary from its official image pinned by digest into a pinned Python slim builder. Astral's [Docker guidance](https://docs.astral.sh/uv/guides/integration/docker/) recommends digest pins where reproducibility requires them, dependency/source layer separation, a fresh image-owned `.venv`, and `--no-editable` when only the environment crosses stages.

1. Copy `pyproject.toml` and `uv.lock`; run the dependency-only sync in a cacheable layer.
2. Copy the complete project; run `uv sync --frozen --no-dev --no-editable` to install the locked application into `/app/.venv`.
3. Copy `/app/.venv` into a pinned compatible Python slim runtime. Do not copy uv, source checkout, compiler, cache, tests, or development tools unless runtime behavior requires an explicit artifact.

Set `UV_LINK_MODE=copy` when the build cache and target are on different filesystems. Set `UV_PYTHON_DOWNLOADS=0` when both stages intentionally use the image's compatible interpreter. Keep `.venv`, `.git`, secrets, local env files, caches, coverage, and build output out of the context.

The runtime sets `PYTHONDONTWRITEBYTECODE=1`, `PYTHONUNBUFFERED=1`, `PATH=/app/.venv/bin:$PATH`, a read-only working directory where supported, and a numeric non-root `USER`. Writable temp/cache paths are explicit and bounded. Install runtime OS packages only when a locked Python dependency or certificate/timezone requirement proves the need.

### Image Identity And Supply Chain

Pin builder, uv, and runtime images by digest in release configuration. Tag the produced image with the release version and immutable commit SHA; deployment manifests consume the digest, not `latest` or a mutable version tag. Set OCI source, revision, version, and created labels and emit safe version/revision once at startup.

Review `docker history`, package contents, native libraries, and vulnerability results. The build context and layers contain no `.env`, credentials, package-index tokens, or mounted secret values. CI uses BuildKit secret mounts for any authenticated build input and proves it leaves no layer.

### Uvicorn Process Model

Prefer one Uvicorn worker per container and horizontal replicas, matching Uvicorn's [container guidance](https://uvicorn.dev/deployment/docker/). This makes memory, readiness, shutdown, and autoscaling visible to the platform. Multiple `--workers` are allowed only when a single-container platform cannot replicate or measured CPU utilization justifies them; then size database/HTTP pools and memory per worker and prove multiprocess Prometheus behavior.

Never combine `--reload` and production. Configure the import string/app factory, host, port, concurrency limit, backlog, keep-alive, and `--timeout-graceful-shutdown` explicitly from deployment settings. Uvicorn documents worker and graceful-shutdown controls in its [deployment](https://uvicorn.dev/deployment/) and [settings](https://uvicorn.dev/settings/) references.

### Resources And Capacity

Every container declares CPU/memory requests and limits from load-test evidence. Size Uvicorn concurrency, HTTPX/SQLAlchemy pools, semaphores, and queue capacities together across the maximum replica count. Leave memory headroom for the interpreter, native extensions, connection buffers, telemetry, and bursts; a Python heap is not the container's full resident set.

Autoscaling uses a signal correlated with saturation and preserves downstream capacity. A replica increase that exhausts PostgreSQL or an upstream is forbidden. Investigate sustained memory growth with allocation/profiling evidence; recycling workers may contain a known third-party leak temporarily but requires an owner and removal condition.

### Config And Secrets

Environment-specific configuration and secrets arrive at runtime as orchestrator environment variables or mounted files. The image contains only safe defaults. Pydantic settings validates before readiness; missing/invalid required values name the key, never its value. Rotation defaults to a rolling restart under [security](security.md).

### Health Probes

- Liveness calls `/livez`; it is local and never fails because PostgreSQL or another dependency is down.
- Readiness calls `/readyz`; it remains false during startup and drain and fails within a bound when a required dependency is unavailable.
- Startup probes or sufficient initial delay cover initialization without weakening steady-state liveness.

Probe timeouts stay below the orchestrator timeout. Responses expose no DSNs, credentials, internal addresses, or raw exceptions. The [Kubernetes template](../templates/k8s-deployment.yaml) wires these endpoints, resources, security context, and grace period.

### Graceful Shutdown

On SIGTERM, readiness turns false and intake stops before work drains. Uvicorn and the application share a bounded shutdown budget; the configured Uvicorn graceful-shutdown timeout stays below the platform termination grace after accounting for endpoint propagation, any `preStop`, resource closure, and headroom. The platform must not send SIGKILL while accepted work is still inside the advertised application grace.

Workers stop intake, await in-flight tasks, close clients and engines after users stop, flush telemetry last, and then exit. A second signal may force termination according to the runbook.

### Production Migrations

Run `uv run alembic upgrade head` as one explicit pre-deploy job using the release artifact and production configuration. Wait for success before shifting traffic. Normal application startup never migrates, and concurrent replicas never race for migration ownership. Expand/contract compatibility keeps old and new code safe across the rollout; a failed migration stops deployment.

## Common Mistakes And Forbidden Patterns

- Mutable or unpinned builder/runtime/uv images; deployments consuming `latest` instead of a digest.
- Editable install or source checkout copied into the final image when only the installed package is required.
- uv, compiler, build cache, tests, credentials, `.env`, or package-index tokens in runtime layers.
- Root runtime, broadly writable filesystem, unbounded temp space, or secrets baked into `ARG`/`ENV`.
- Multiple workers by reflex, with pools and Prometheus state sized as though only one process exists.
- `--reload` in production or unspecified Uvicorn concurrency/graceful-shutdown controls.
- Resource limits selected without replica/downstream math; autoscaling that overwhelms PostgreSQL.
- Liveness querying dependencies, readiness true during drain, or probe output leaking topology.
- Application grace at or above platform grace; clients/engines closed before users drain.
- Alembic run by every app replica on startup.

## Verification And Proof

```bash
docker build --tag <app>:verify .
docker run --rm --read-only --user <uid>:<gid> <app>:verify --help
uv run alembic upgrade head
make verify
```

Inspect the image digest, labels, history, user, environment, package inventory, and writable mounts. Start the built image with real local dependencies; prove `/livez`, `/readyz`, and `/metrics`, dependency-driven readiness, and exact version identity. Send SIGTERM during slow HTTP and worker work and prove readiness drops, work drains, resources close, and exit completes inside platform grace. Apply migrations from empty before the rollout and show normal app startup never changes schema.

## Related

- [CI and release](ci-and-release.md)
- [observability](observability.md)
- [resilience](resilience.md)
- [configuration](../foundations/configuration.md)
- [concurrency and asyncio](../foundations/concurrency-and-asyncio.md)
