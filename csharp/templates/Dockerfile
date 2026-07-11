# Multi-stage build for a handbook ASP.NET Core service. Contract:
# csharp/operations/deployment.md
# Destination: Dockerfile (repo root).
#
# Replace "Orders" with your app name throughout (projects, entrypoint DLL).

# ---- build stage ----------------------------------------------------------
# Full SDK image; never ships. The SDK inside must satisfy the committed
# global.json (rollForward: latestFeature), so the image tag tracks the same
# 10.0 line the repo pins.
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build

WORKDIR /src

# Cache dependency layers independently of source changes: copy ONLY the
# restore inputs first (solution, central package versions, feeds, per-project
# manifests + lock files), so day-to-day source edits reuse the restore layer.
COPY *.slnx nuget.config Directory.Build.props Directory.Packages.props ./
COPY src/Orders.Api/Orders.Api.csproj src/Orders.Api/packages.lock.json src/Orders.Api/
COPY src/Orders.Core/Orders.Core.csproj src/Orders.Core/packages.lock.json src/Orders.Core/
COPY src/Orders.Infrastructure/Orders.Infrastructure.csproj src/Orders.Infrastructure/packages.lock.json src/Orders.Infrastructure/

# Restore the HOST project, not the solution: project references pull Core and
# Infrastructure transitively, and the test projects never enter the image
# build (they are also excluded by .dockerignore). --locked-mode fails the
# build if any lock file has drifted from the csproj — same rule as verify.ps1.
RUN dotnet restore src/Orders.Api/Orders.Api.csproj --locked-mode

# Now the source.
COPY src/ src/

# Version metadata stamped into the published assembly. Override at build time:
#   docker build --build-arg VERSION=v1.2.3 --build-arg COMMIT=$(git rev-parse HEAD) ...
# (the committed release workflow passes both). Surfaces at runtime via
# AssemblyInformationalVersionAttribute — the app exposes it in logs/telemetry.
ARG VERSION=dev
ARG COMMIT=unknown

# Publish framework-dependent (the runtime image below supplies the runtime).
# --no-restore: the locked restore above is the only restore this build runs.
RUN dotnet publish src/Orders.Api/Orders.Api.csproj \
      -c Release \
      --no-restore \
      -o /app/publish \
      -p:InformationalVersion="${VERSION}+${COMMIT}"

# ---- runtime stage --------------------------------------------------------
# Chiseled Ubuntu runtime: no shell, no package manager, and it ALREADY runs as
# the non-root `app` user (UID 1654) — no USER line needed; the k8s manifest
# still enforces runAsNonRoot at the platform layer.
#
# Globalization: the -extra chiseled tag ships icu + tzdata, matching the
# handbook default of invariant globalization OFF for services that handle
# user data (csharp/foundations/cross-platform.md). A machine-facing service
# that deliberately sets <InvariantGlobalization>true</InvariantGlobalization>
# may switch to the leaner mcr.microsoft.com/dotnet/aspnet:10.0-noble-chiseled
# (no ICU, no tzdata) — record that choice in the README.
FROM mcr.microsoft.com/dotnet/aspnet:10.0-noble-chiseled-extra

# OCI labels tie the image back to source and the release tag. The release
# workflow's metadata step supplies richer labels; these are the offline
# fallback. Override via --build-arg CREATED=$(date -u +%Y-%m-%dT%H:%M:%SZ).
ARG VERSION=dev
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
# baked into the image — non-root users cannot bind 80). EXPOSE is
# documentation only; it does not publish the port. To change the bind, set
# ASPNETCORE_URLS (e.g. ASPNETCORE_URLS=http://+:8080) at RUNTIME in compose or
# the k8s manifest, not here.
EXPOSE 8080

# Exec form — there is no shell in this image. dotnet is PID 1; the generic
# host translates SIGTERM into graceful shutdown (IHostApplicationLifetime +
# HostOptions.ShutdownTimeout — see csharp/operations/deployment.md).
ENTRYPOINT ["dotnet", "Orders.Api.dll"]
