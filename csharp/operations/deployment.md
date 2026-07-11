# Deployment

Containerization and rollout defaults for a handbook service: build a small, hardened, reproducible image and run it so it survives the platform's lifecycle without OOM kills, CPU throttling, or truncated shutdowns.

## Default Approach

A service ships as a framework-dependent `dotnet publish` output running on the chiseled ASP.NET runtime image, non-root, one process per container. The build is multi-stage: the fat SDK image compiles, the tiny runtime image runs. Image identity is tied to the release tag, runtime resource behavior is driven by the container limits the .NET runtime already reads, and probes are wired to the same `/livez` and `/readyz` endpoints the service already exposes.

The committed [`csharp/templates/Dockerfile`](../templates/Dockerfile) and [`csharp/templates/.dockerignore`](../templates/.dockerignore) are the canonical starting point; this doc is the contract they implement.

### Multi-Stage Build

- Build stage uses the official .NET SDK image; runtime stage uses the chiseled ASP.NET runtime image. Copy the verified tags from the committed [templates/Dockerfile](../templates/Dockerfile) rather than a version written in prose; the SDK the stage uses must satisfy the repo's committed [global.json](../templates/global.json).
- Publish posture: `dotnet publish -c Release --no-restore -o /app` — framework-dependent, because the runtime already lives in the base image. Self-contained publish onto an `aspnet` base ships the runtime twice. Native AOT and ReadyToRun are not defaults; they change trimming, reflection, and diagnostic behavior and require an [ADR](../decisions/architecture-decision-records.md).
- Layer caching: copy `global.json`, `nuget.config`, `Directory.Build.props`, `Directory.Packages.props`, the `.csproj` files, and `packages.lock.json` first and run `dotnet restore --locked-mode`, then copy source and publish with `--no-restore`. Dependency layers cache independently of code changes, and locked-mode restore proves the lockfile in the image build too ([ci-and-release.md](ci-and-release.md)).
- Version stamping: the Docker build context has no `.git`, so MinVer cannot run inside the image build. CI computes the version from the release tag and passes it as a build arg into `-p:Version=...` (see [ci-and-release.md](ci-and-release.md)); the assembly's informational version, the OCI labels, and the VCS tag must all agree. The service logs `service`/`version`/`commit` at startup from that stamped metadata.

### Runtime Image

- Default runtime base: chiseled ASP.NET (Ubuntu-based, distroless-style). It ships CA certificates and a non-root `app` user, but no shell, no package manager, and no coreutils — minimal attack surface while still able to make TLS calls.
- **Document the globalization choice.** The plain chiseled image contains no ICU and no tzdata, so it only runs apps built with `<InvariantGlobalization>true</InvariantGlobalization>`. The handbook default for services that handle user data is invariant globalization OFF, which means the `-extra` chiseled variant (ICU + tzdata included) — the committed template uses it and pins the tag. Machine-facing services may use the plain chiseled image, but only with `InvariantGlobalization` set deliberately in the csproj, never by letting culture behavior degrade silently (see [../foundations/cross-platform.md](../foundations/cross-platform.md)).
- Copy only the publish output from the builder. No source, no NuGet cache, no SDK.
- Non-root is the default in chiseled images; keep it that way. `USER $APP_UID` makes the intent explicit in non-chiseled variants. Never run as root.
- `ENTRYPOINT ["dotnet", "Orders.Api.dll"]` — exec form, no shell wrapper, no init script. The `dotnet` process is PID 1 and receives `SIGTERM` directly (see Graceful Shutdown below). Shell form (`ENTRYPOINT dotnet Orders.Api.dll`) puts a shell at PID 1 and orphans signal handling.
- `EXPOSE 8080` for documentation; it does not publish anything by itself.

### Port And URL Contract

- The ASP.NET runtime images set `ASPNETCORE_HTTP_PORTS=8080` — a non-privileged port, which is what lets the non-root user bind it. Kestrel listens on all interfaces on that port; do not bind `localhost` inside a container.
- `ASPNETCORE_URLS`, when set, **replaces** `ASPNETCORE_HTTP_PORTS` entirely (the runtime warns and ignores the ports variable). Pick one mechanism per service — the default is to leave the image's port default alone and let the manifest own any override.
- The port is a contract with the platform: `containerPort`, the liveness/readiness probe ports, and the Service `targetPort` in [templates/k8s-deployment.yaml](../templates/k8s-deployment.yaml) must all agree with what Kestrel binds. TLS terminates at the platform edge by default; in-cluster TLS is a [framework-selection](../decisions/framework-selection.md) choice.

### Container Runtime Limits

The .NET runtime is container-aware: at startup it reads the cgroup (v1 and v2) memory and CPU limits and sizes the GC and thread pool from them. Unlike Go — where `GOMEMLIMIT`/`GOMAXPROCS` must be set by hand — the defaults are usually right, but only if the container *has* limits. Always set them; a limitless container makes the runtime size itself for the whole node.

- **Memory.** With a container memory limit set, the GC heap hard limit defaults to `max(20 MB, 75% of the limit)`, leaving 25% headroom for stacks, native allocations, and mapped files. As the heap approaches the limit the GC collects more aggressively, and an allocation that would exceed it throws `OutOfMemoryException` instead of waiting for the kernel OOM killer. Override only with a measured reason via `DOTNET_GCHeapHardLimitPercent` or `DOTNET_GCHeapHardLimit` — and note the footgun: these env vars take **hexadecimal** values.
- **GC mode.** ASP.NET Core services default to Server GC, and current runtimes enable DATAS (dynamic adaptation to application sizes) with it, so the heap tracks live data size instead of ballooning to core count — the right default for containers. When the container's CPU limit yields a single logical processor, the runtime falls back to Workstation GC on its own. Keep these defaults; disabling DATAS or forcing a GC mode requires a measured justification in an [ADR](../decisions/architecture-decision-records.md).
- **CPU.** `Environment.ProcessorCount` derives from the cgroup CPU quota (fractional quotas round up), and the thread pool and GC heap count follow it — no `automaxprocs` equivalent needed. `DOTNET_PROCESSOR_COUNT` overrides it when a fractional quota needs an explicit answer. Avoid sub-1-CPU limits for latency-sensitive services; they surface as throttling-induced tail latency, not as errors.

### Secrets And Config Injection

Secrets and environment-specific config reach the container at RUNTIME from the platform; they are never baked into the image.

- Inject secrets as environment variables or files mounted from the orchestrator's secret store (a Kubernetes Secret, a cloud secrets-manager CSI driver, or a Vault agent/sidecar); the image, its layers, and its build args contain NO secret material (see [security.md](security.md)). The specific manager is a [framework-selection](../decisions/framework-selection.md) choice.
- The process reads injected material once at startup through the options pattern with `ValidateOnStart` — fail fast on missing or malformed config, per [../foundations/configuration.md](../foundations/configuration.md). `dotnet user-secrets` is local-dev only and never reaches a container.
- Rotation: the default is a rolling restart that re-runs startup validation; a secret that must rotate without a restart is mounted as a file and read through the configuration reload path. Decide per secret and record it in the [runbook](../templates/runbook.md).
- Per-environment values (staging vs production) come from the manifest and secret store, not from env-specific `appsettings.{Environment}.json` defaults committed with secrets in them.

### Health Probes

- Liveness probe -> `GET /livez`: process-up only. Failing it restarts the container, so it must not depend on downstreams (see [operations/observability.md](observability.md)).
- Readiness probe -> `GET /readyz`: dependency-aware. The platform must gate traffic on readiness, so a pod whose database is unreachable is pulled from rotation instead of serving errors.
- Keep liveness and readiness distinct — separate health-check tags, separate endpoints. A flaky dependency should fail readiness (drain traffic) without failing liveness (restart loop).

### Graceful Shutdown And Termination

- On `SIGTERM` the host stops accepting new connections, Kestrel drains in-flight requests, and hosted services get their `stoppingToken` cancelled — all bounded by `HostOptions.ShutdownTimeout`. See [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md) for the in-process contract.
- **The ordering is strict: `terminationGracePeriodSeconds` > `HostOptions.ShutdownTimeout` > the longest in-flight request (or worker unit of work).** The platform grace must exceed the app grace — plus any `preStop` delay — or the platform sends `SIGKILL` mid-drain; the app grace must exceed the slowest request or the host abandons it at the deadline.
- Do not rely on the defaults lining up: `HostOptions.ShutdownTimeout` defaults to 30 seconds and Kubernetes' `terminationGracePeriodSeconds` also defaults to 30 seconds — zero headroom. Set both explicitly; the committed [templates/k8s-deployment.yaml](../templates/k8s-deployment.yaml) keeps the platform grace above the app grace.
- Budget: readiness flips to not-ready, the load balancer stops routing, in-flight requests finish, hosted services stop, *then* the process exits — all inside the platform grace. Pair shutdown with retry/timeout discipline at clients so a rolling deploy does not surface as user-visible errors (see [operations/resilience.md](resilience.md)).

### Windows Dev, Linux Containers

- Production containers are Linux — the chiseled images have no Windows variant, and this handbook does not target Windows containers (an org that mandates them writes an [ADR](../decisions/architecture-decision-records.md)).
- Windows and macOS developers build and run the same Linux images via Docker Desktop (WSL2 on Windows) or an equivalent Linux-VM runtime. That is the dev/prod parity story: `pwsh ./verify.ps1` proves the code on all three OSes, but the container is the production truth, so path casing, line endings, and file-locking differences must be caught before the image — see [../foundations/cross-platform.md](../foundations/cross-platform.md).
- Anything that only reproduces "in the container" is debugged by running the image locally; the local compose stack (below) exists so that path is one command.

### Orchestration

The probes, runtime limits, and grace-period budget above are not just prose — they are committed, copyable manifests under [`csharp/templates/`](../templates/README.md):

- [`docker-compose.yml`](../templates/docker-compose.yml) — the local stack: the service plus a version-pinned Postgres (the pin lives in the template) with a healthcheck and `depends_on: service_healthy`, wiring the connection string and config keys so the SQL path runs the same way CI's integration job does ([ci-and-release.md](ci-and-release.md)).
- [`k8s-deployment.yaml`](../templates/k8s-deployment.yaml) — the production rollout: a non-root Deployment with resource requests/limits, `/livez` liveness + `/readyz` readiness probes, `terminationGracePeriodSeconds` above the app shutdown grace, a connection-string Secret, a Service, an HPA, and a migration Job (below).

Copy them and adjust the image, namespace, secret name, and resource numbers; this doc is the contract they implement, so do not restate it in the manifests.

### Production Migrations

The "apply before traffic" default from [services/database.md](../services/database.md) has a concrete mechanism: the deploy pipeline runs a **migration Job** — the same app image invoked with the `--migrate` flag, which applies the committed EF Core migrations and exits — and waits for it to complete before rolling the Deployment. The k8s template ships this Job as its second document. One writer applies the schema change; the new pods only start against a schema that is already in place. Auto-migrating on normal startup is never the production path — concurrent replicas would race to apply — and the locked default is an explicit migration step only (see [../recipes/add-migration.md](../recipes/add-migration.md)).

## Common Mistakes And Forbidden Patterns

- Running as root, or undoing the base image's non-root user in the final stage.
- A shell or package manager in the runtime image; a shell-form `ENTRYPOINT` that orphans signal handling and breaks graceful shutdown.
- Self-contained or AOT publish onto an `aspnet` base image — the runtime twice, or an ADR-gated posture adopted by accident.
- Using the plain chiseled image without setting `InvariantGlobalization`, so the service throws `CultureNotFoundException` (or silently mis-sorts user data) at runtime instead of failing the build decision.
- Shipping the whole build context or SDK into the final image; not using a `.dockerignore`.
- Unpinned builder image or `latest` runtime tags, making builds and rollbacks irreproducible.
- No container memory limit, so the GC sizes itself for the node and the pod is the OOM killer's first pick; or overriding `DOTNET_GCHeapHardLimit*` without a measurement (and getting the hex value wrong).
- Setting both `ASPNETCORE_URLS` and `ASPNETCORE_HTTP_PORTS` and being surprised which one wins; probes and Service pointing at a port Kestrel no longer binds.
- Liveness probe that checks downstream dependencies, turning a dependency blip into a restart loop.
- `HostOptions.ShutdownTimeout` at or above `terminationGracePeriodSeconds` (no headroom for the preStop delay and exit), so the platform cuts the drain off with `SIGKILL` — or both left at their identical 30-second defaults.
- Auto-migrating on normal startup in production, so concurrent replicas race to apply the schema.

## Verification And Proof

- `docker build` succeeds and `docker run` starts the service, which logs its `service`/`version`/`commit` at startup.
- `docker inspect` shows a non-root user and the expected `ENTRYPOINT`; image inspection shows no shell, no SDK — publish output plus base only.
- The stamped informational version matches the image's OCI labels and the VCS tag.
- In the running container, `Environment.ProcessorCount` and the GC heap limit reflect the cgroup limits (the startup log or `dotnet-counters` confirms), not host defaults.
- Liveness and readiness probes return the expected codes; killing a critical dependency flips `/readyz` to not-ready while `/livez` stays up, and traffic drains.
- `docker stop` (SIGTERM) drains in-flight requests and exits cleanly inside `HostOptions.ShutdownTimeout`, itself inside the platform termination grace.
- The migration Job runs the image with `--migrate`, exits 0, and the Deployment only rolls afterwards.

## Related

- [operations/ci-and-release.md](ci-and-release.md)
- [operations/observability.md](observability.md)
- [operations/resilience.md](resilience.md)
- [operations/operability.md](operability.md)
- [../foundations/configuration.md](../foundations/configuration.md)
- [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md)
- [../foundations/cross-platform.md](../foundations/cross-platform.md)
