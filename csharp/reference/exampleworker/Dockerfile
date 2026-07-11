# Multi-stage build for the handbook event worker. Contract:
# csharp/operations/deployment.md
# Adapted from the keystone exampleservice Dockerfile with the project names
# substituted (Orders.Api -> Orders.Worker).

# ---- build stage ----------------------------------------------------------
# Full SDK image; never ships. The SDK inside must satisfy the committed
# global.json (rollForward: latestFeature), so the image tag tracks the same
# 10.0 line the repo pins.
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build

WORKDIR /src

# Cache dependency layers independently of source changes: copy ONLY the
# restore inputs first (solution, central package versions, feeds, per-project
# manifests + lock files), so day-to-day source edits reuse the restore layer.
# .editorconfig rides along because it is the single source of analyzer
# severities - without it the in-image build re-enables rules (e.g. CS1591)
# that Directory.Build.props' warnings-as-errors then turns fatal.
COPY *.slnx nuget.config Directory.Build.props Directory.Packages.props .editorconfig ./
COPY src/Orders.Worker/Orders.Worker.csproj src/Orders.Worker/packages.lock.json src/Orders.Worker/
COPY src/Orders.Worker.Core/Orders.Worker.Core.csproj src/Orders.Worker.Core/packages.lock.json src/Orders.Worker.Core/
COPY src/Orders.Worker.Infrastructure/Orders.Worker.Infrastructure.csproj src/Orders.Worker.Infrastructure/packages.lock.json src/Orders.Worker.Infrastructure/

# Restore the HOST project, not the solution: project references pull Core and
# Infrastructure transitively, and the test projects never enter the image
# build (they are also excluded by .dockerignore). --locked-mode fails the
# build if any lock file has drifted from the csproj - same rule as verify.ps1.
RUN dotnet restore src/Orders.Worker/Orders.Worker.csproj --locked-mode

# Now the source.
COPY src/ src/

# Version metadata stamped into the published assembly. Override at build time:
#   docker build --build-arg VERSION=1.2.3 --build-arg COMMIT=$(git rev-parse HEAD) ...
# (the committed release workflow passes both). Surfaces at runtime via
# AssemblyInformationalVersionAttribute - the app exposes it in logs/telemetry.
# The default MUST be a valid semver prefix: the SDK validates the part before
# '+' as a version (NETSDK1018 rejects a bare "dev").
ARG VERSION=0.0.0-dev
ARG COMMIT=unknown

# Publish framework-dependent (the runtime image below supplies the runtime).
# --no-restore: the locked restore above is the only restore this build runs.
RUN dotnet publish src/Orders.Worker/Orders.Worker.csproj \
      -c Release \
      --no-restore \
      -o /app/publish \
      -p:InformationalVersion="${VERSION}+${COMMIT}"

# ---- runtime stage --------------------------------------------------------
# The worker hosts its probe endpoints on Kestrel, so it runs on the aspnet
# runtime image like the keystone service. Chiseled Ubuntu: no shell, no
# package manager, already non-root (`app`, UID 1654).
#
# Globalization: the -extra chiseled tag ships icu + tzdata, matching the
# handbook default of invariant globalization OFF
# (csharp/foundations/cross-platform.md).
FROM mcr.microsoft.com/dotnet/aspnet:10.0-noble-chiseled-extra

# OCI labels tie the image back to source and the release tag. The release
# workflow's metadata step supplies richer labels; these are the offline
# fallback. Override via --build-arg CREATED=$(date -u +%Y-%m-%dT%H:%M:%SZ).
ARG VERSION=0.0.0-dev
ARG COMMIT=unknown
ARG CREATED=
LABEL org.opencontainers.image.source="https://github.com/<org>/<repo>" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${CREATED}"

WORKDIR /app

# Copy only the publish output.
COPY --from=build /app/publish .

# The aspnet images listen on 8080 by default (ASPNETCORE_HTTP_PORTS=8080 is
# baked into the image - non-root users cannot bind 80). EXPOSE is
# documentation only; it does not publish the port.
EXPOSE 8080

# Exec form - there is no shell in this image. dotnet is PID 1; the generic
# host translates SIGTERM into graceful shutdown (IHostApplicationLifetime +
# HostOptions.ShutdownTimeout - see csharp/operations/deployment.md).
ENTRYPOINT ["dotnet", "Orders.Worker.dll"]
