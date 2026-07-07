# Multi-stage build for a handbook Go service. Contract: golang/operations/deployment.md
# Destination: Dockerfile (repo root).

# ---- build stage ----------------------------------------------------------
# Pin the builder to a specific patch release (latest stable). The language
# baseline is Go 1.24+; the go.mod `go` directive is 1.26 because pinning
# golangci-lint and sqlc as `go tool` dependencies requires it (the sqlc tool
# graph sets the floor; see README/Makefile).
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
ARG APP=exampleservice

# Static binary (CGO_ENABLED=0) so the runtime image can be distroless static.
# -trimpath strips local paths (release builds only); -s -w drop the symbol table.
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w \
        -X github.com/example/exampleservice/internal/buildinfo.Name=${APP} \
        -X github.com/example/exampleservice/internal/buildinfo.Version=${VERSION} \
        -X github.com/example/exampleservice/internal/buildinfo.Commit=${COMMIT}" \
      -o /out/app ./cmd/${APP}

# ---- runtime stage --------------------------------------------------------
# Distroless static + nonroot: CA certs, tzdata, and a nonroot user; no shell, no libc.
# Alternative is scratch (see deployment.md) — distroless is the default.
FROM gcr.io/distroless/static-debian13:nonroot

# OCI labels tie the image back to source and the release tag. Override via --build-arg.
#   --build-arg CREATED=$(date -u +%Y-%m-%dT%H:%M:%SZ)
ARG VERSION=dev
ARG COMMIT=unknown
ARG CREATED=
LABEL org.opencontainers.image.source="https://github.com/example/exampleservice" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${CREATED}"

# Copy only the binary.
COPY --from=build /out/app /app

# nonroot user (UID 65532) is provided by the base image.
USER nonroot:nonroot

# Documentation only; does not publish the port.
EXPOSE 8080

# Binary is PID 1 and handles SIGTERM itself (graceful shutdown). Exec form — no shell.
ENTRYPOINT ["/app"]
