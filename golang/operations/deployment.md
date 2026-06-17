# Deployment

Containerization and rollout defaults for a handbook service: build a small, hardened, reproducible image and run it so it survives the platform's lifecycle without OOM kills or CPU throttling.

## Default Approach

A service ships as a single static binary in a minimal, nonroot container. The build is multi-stage: a fat builder image compiles, a tiny runtime image runs. Image identity is tied to the release tag, runtime resource hints match the container limits, and probes are wired to the same `/livez` and `/readyz` endpoints the service already exposes.

The committed [`golang/templates/Dockerfile`](../templates/Dockerfile) and [`golang/templates/.dockerignore`](../templates/.dockerignore) are the canonical starting point; this doc is the contract they implement.

### Multi-Stage Build

- Build stage uses the official `golang` image pinned to a specific patch (e.g. `golang:1.26.4`); the module's `go.mod` baseline stays `go 1.24` with the toolchain pinned to latest stable.
- Compile a static binary: `CGO_ENABLED=0`. No cgo means no glibc dependency, which is what lets the runtime image be `static`/`scratch`.
- Build with `-trimpath` (strips local filesystem paths; release builds only, never the routine `make build`) and stamp version metadata through `-ldflags` into the `internal/buildinfo` package, matching the variables the binary already logs at startup (`Name`, `Version`, `Commit`).
- The version stamped in is the release tag in canonical `v1.2.3` form (see [operations/ci-and-release.md](ci-and-release.md)).
- Layer caching: copy `go.mod`/`go.sum` and run `go mod download` before copying source, so dependency layers cache independently of code changes.

### Runtime Image

- Default runtime base: `gcr.io/distroless/static-debian13:nonroot`. It ships CA certificates, `/etc/passwd` with a `nonroot` user, and tzdata, but no shell, no package manager, and no libc — minimal attack surface while still able to make TLS calls and resolve time zones.
- Alternative: `scratch`. Smaller still, but you must `COPY` CA certs and tzdata yourself and create the nonroot UID manually; choose it only when you have a measured reason. Distroless static is the default because it removes those footguns at a negligible size cost.
- Copy only the binary from the builder. No source, no build cache, no toolchain.
- `USER nonroot` (or numeric `USER 65532` on scratch). Never run as root.
- `ENTRYPOINT` is the binary directly. No shell wrapper, no init script — the process is PID 1 and must handle signals itself (see Graceful Shutdown below).
- `EXPOSE` the service port for documentation; it does not publish anything by itself.

### Image Identity And Versioning

- Tag images with the release tag (`v1.2.3`) plus the immutable commit SHA; avoid relying on `latest` for anything an operator must reason about.
- Set OCI labels so a running image is traceable back to source: `org.opencontainers.image.version`, `org.opencontainers.image.revision`, `org.opencontainers.image.source`, `org.opencontainers.image.created`.
- The label version, the `-ldflags` stamped version, and the VCS tag must all agree. `go version -m <binary>` and the service's own startup log are the proof.

### Container Runtime Limits

Go's runtime defaults assume it sees the whole host. In a container with cgroup limits, those defaults cause OOM kills and CPU throttling unless corrected.

- **Memory — set `GOMEMLIMIT`.** Set it to a soft ceiling *below* the container memory limit (a common starting point is ~90% of the limit, e.g. `GOMEMLIMIT=900MiB` under a 1 GiB limit). `GOMEMLIMIT` makes the GC work harder as the heap approaches the ceiling instead of letting the kernel OOM-kill the process. It is a soft limit, so leave headroom for non-heap memory (stacks, mmap, off-heap). Drive it from the deployment manifest so it tracks the limit.
- **CPU — right-size `GOMAXPROCS` to the CPU quota.** A pod with a 2-core quota on a 64-core node still defaults `GOMAXPROCS` to 64, producing oversubscription and scheduler-induced throttling. Set `GOMAXPROCS` to the integer CPU quota, or adopt the `automaxprocs` convention to derive it from the cgroup at startup. The specific mechanism (env var driven by the manifest vs. an `automaxprocs`-style library import) is a library choice routed to [decisions/framework-selection.md](../decisions/framework-selection.md); the *requirement* — `GOMAXPROCS` matches the quota — is not optional.

### Health Probes

- Liveness probe -> `GET /livez`: process-up only. Failing it restarts the container, so it must not depend on downstreams (see [operations/observability.md](observability.md)).
- Readiness probe -> `GET /readyz`: dependency-aware. The platform must gate traffic on readiness, so a pod whose database is unreachable is pulled from rotation instead of serving errors.
- Keep liveness and readiness distinct. A flaky dependency should fail readiness (drain traffic) without failing liveness (restart loop).

### Graceful Shutdown And Termination

- On `SIGTERM` the process drains in-flight work within a bounded grace period — see [foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md#graceful-shutdown-and-draining).
- The application's shutdown grace period must **exceed** the platform's termination grace period (e.g. Kubernetes `terminationGracePeriodSeconds`), or the platform sends `SIGKILL` mid-drain. Budget: readiness flips to not-ready, the load balancer stops routing, in-flight requests finish, *then* the process exits — all inside the platform grace.
- Pair shutdown with retry/timeout discipline at clients so a rolling deploy does not surface as user-visible errors (see [operations/resilience.md](resilience.md)).

## Common Mistakes And Forbidden Patterns

- Running as root, or `USER root` left in the final stage.
- A shell or package manager in the runtime image; a shell-form `ENTRYPOINT` that orphans signal handling and breaks graceful shutdown.
- Building with cgo enabled (`CGO_ENABLED` unset) and then trying to run on `static`/`scratch`, producing a binary that fails on a missing dynamic loader.
- Shipping the whole build context or `go` toolchain into the final image; not using a `.dockerignore`.
- Unpinned builder image (`golang:latest`) or `latest` runtime tags, making builds and rollbacks irreproducible.
- No `GOMEMLIMIT`, so the process OOM-kills under load instead of GC'ing harder; or `GOMEMLIMIT` set at or above the container limit, which defeats the purpose.
- Leaving `GOMAXPROCS` at the host core count under a fractional/limited CPU quota, causing throttling and tail-latency spikes.
- Liveness probe that checks downstream dependencies, turning a dependency blip into a restart loop.
- Application shutdown grace shorter than the platform termination grace, so drains are cut off by `SIGKILL`.

## Verification And Proof

- `docker build` succeeds and `docker run` starts the service, which logs its `service`/`version`/`commit` at startup.
- `docker history` / image inspection shows no shell, no toolchain, and a nonroot user; image is the binary plus base only.
- `go version -m <binary>` shows the expected module metadata and the version matches the image label and VCS tag.
- `GOMEMLIMIT` and `GOMAXPROCS` in the running container reflect the configured limits (inspect env / process state), not host defaults.
- Liveness and readiness probes return the expected codes; killing a critical dependency flips `/readyz` to not-ready while `/livez` stays up, and traffic drains.
- A `SIGTERM` to the running container drains in-flight requests and exits cleanly within the platform termination grace period.

## Related

- [operations/ci-and-release.md](ci-and-release.md)
- [operations/observability.md](observability.md)
- [operations/resilience.md](resilience.md)
- [operations/operability.md](operability.md)
- [foundations/configuration.md](../foundations/configuration.md)
- [foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md)
