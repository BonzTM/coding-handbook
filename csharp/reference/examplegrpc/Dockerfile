# Multi-stage build for a handbook gRPC service. Contract:
# csharp/operations/deployment.md
# Destination: Dockerfile (repo root).
#
# Adapted from the keystone exampleservice copy for this module's two-project
# shape (no Infrastructure project - no database).

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
COPY src/Orders.Grpc.Api/Orders.Grpc.Api.csproj src/Orders.Grpc.Api/packages.lock.json src/Orders.Grpc.Api/
COPY src/Orders.Grpc.Core/Orders.Grpc.Core.csproj src/Orders.Grpc.Core/packages.lock.json src/Orders.Grpc.Core/

# The proto is a build input: Grpc.Tools generates the stubs during publish,
# so the wire contract must be in the context BEFORE the source build.
COPY api/ api/

# Restore the HOST project, not the solution: project references pull Core
# transitively, and the test project never enters the image build (also
# excluded by .dockerignore). --locked-mode fails the build if any lock file
# has drifted from the csproj - same rule as verify.ps1.
RUN dotnet restore src/Orders.Grpc.Api/Orders.Grpc.Api.csproj --locked-mode

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
RUN dotnet publish src/Orders.Grpc.Api/Orders.Grpc.Api.csproj \
      -c Release \
      --no-restore \
      -o /app/publish \
      -p:InformationalVersion="${VERSION}+${COMMIT}"

# ---- runtime stage --------------------------------------------------------
# Chiseled Ubuntu runtime: no shell, no package manager, and it ALREADY runs as
# the non-root `app` user (UID 1654) - no USER line needed; the k8s manifest
# still enforces runAsNonRoot at the platform layer.
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

# This host configures its own listeners from config (Server:GrpcPort,
# Server:ProbesPort - see appsettings.json), NOT from ASPNETCORE_URLS: gRPC
# needs an HTTP/2-only listener and the probes an HTTP/1.1 one. Override the
# ports at RUNTIME via Server__GrpcPort / Server__ProbesPort. EXPOSE is
# documentation only.
EXPOSE 5001 8080

# Exec form - there is no shell in this image. dotnet is PID 1; the generic
# host translates SIGTERM into graceful shutdown (IHostApplicationLifetime +
# HostOptions.ShutdownTimeout - see csharp/operations/deployment.md).
ENTRYPOINT ["dotnet", "Orders.Grpc.Api.dll"]
