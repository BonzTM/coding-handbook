# Multi-stage build for the examplegrpc handbook service. Contract: golang/operations/deployment.md

# ---- build stage ----------------------------------------------------------
# Pin the builder to a specific patch release. The go.mod `go` directive is
# raised by pinning golangci-lint as a `go tool` dependency (see README/Makefile).
FROM golang:1.26.4 AS build

WORKDIR /src

# Cache dependency layers independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

# Now the source.
COPY . .

# Version metadata stamped into internal/buildinfo. Override at build time:
#   docker build --build-arg VERSION=v1.2.3 --build-arg COMMIT=$(git rev-parse HEAD) ...
ARG VERSION=dev
ARG COMMIT=unknown
ARG APP=examplegrpc

# Static binary (CGO_ENABLED=0) so the runtime image can be distroless static.
# -trimpath strips local paths (release builds only); -s -w drop the symbol table.
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w \
        -X github.com/example/examplegrpc/internal/buildinfo.Name=${APP} \
        -X github.com/example/examplegrpc/internal/buildinfo.Version=${VERSION} \
        -X github.com/example/examplegrpc/internal/buildinfo.Commit=${COMMIT}" \
      -o /out/app ./cmd/${APP}

# ---- runtime stage --------------------------------------------------------
# Distroless static + nonroot: CA certs, tzdata, and a nonroot user; no shell, no libc.
FROM gcr.io/distroless/static-debian13:nonroot

ARG VERSION=dev
ARG COMMIT=unknown
ARG CREATED=
LABEL org.opencontainers.image.source="https://github.com/example/examplegrpc" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${CREATED}"

# Copy only the binary.
COPY --from=build /out/app /app

# nonroot user (UID 65532) is provided by the base image.
USER nonroot:nonroot

# Documentation only; does not publish the ports. 9090 gRPC, 8080 metrics/probes.
EXPOSE 9090
EXPOSE 8080

# Binary is PID 1 and handles SIGTERM itself (graceful shutdown). Exec form — no shell.
ENTRYPOINT ["/app"]
