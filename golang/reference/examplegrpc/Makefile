# Makefile - single canonical verification entrypoint.
#
# Adapted from the exampleservice reference for the examplegrpc gRPC service.
# Humans and CI run the SAME targets here; `make verify` is the ordered safety
# gate.
#
# Linter config and policy:   golang/quality/linting.md
# gRPC conventions:           golang/services/grpc-services.md
#
# Conventions:
#   - Tools are run via `go tool <name>` (added with `go get -tool`):
#     golangci-lint v2, govulncheck, buf, protoc-gen-go, protoc-gen-go-grpc.
#   - buf codegen plugins are local executables, so `gen` installs them into
#     $(TOOLBIN) and runs buf with that dir on PATH. Generated *.pb.go are
#     committed; `gen-check` regenerates and asserts no diff so codegen stays
#     deterministic.
#   - PURE-GO: builds set CGO_ENABLED=0 (the handbook mandates it).
#   - -trimpath belongs to release builds only, never the routine `build` target.

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

# Local dir the codegen plugin binaries are installed into for buf.
TOOLBIN := $(CURDIR)/.tools/bin

.PHONY: help tidy fmt fmt-check lint vet test race cover vuln build run \
        tools gen gen-check buf-lint verify

help: ## Show this help.
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "} {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

tidy: ## Sync go.mod/go.sum and verify the module graph.
	go mod tidy
	go mod verify

fmt: ## Format all Go source in place (gofumpt + gci, per .golangci.yml).
	go tool golangci-lint fmt

fmt-check: ## Fail if any file is not formatted (gofumpt + gci; CI-safe, no writes).
	@out="$$(go tool golangci-lint fmt --diff || true)"; \
	if [ -n "$$out" ]; then \
		echo "not formatted (gofumpt/gci):"; echo "$$out"; \
		echo "run: make fmt"; \
		exit 1; \
	fi

lint: ## Run golangci-lint.
	go tool golangci-lint run

vet: ## Run go vet.
	go vet ./...

test: ## Run all tests.
	go test ./...

race: ## Run all tests under the race detector.
	go test -race ./...

cover: ## Write cover.out and print total coverage (atomic mode is race-safe).
	go test -covermode=atomic -coverprofile=cover.out ./...
	go tool cover -func=cover.out | grep '^total:'

vuln: ## Run govulncheck against the module.
	go tool govulncheck ./...

build: ## Compile all packages (pure Go; the handbook mandates CGO_ENABLED=0).
	CGO_ENABLED=0 go build ./...

run: ## Run the service locally (in-memory store, local/dev auth).
	CGO_ENABLED=0 go run ./cmd/examplegrpc

tools: ## Install the buf codegen plugins (protoc-gen-go*) into .tools/bin.
	GOBIN=$(TOOLBIN) go install \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc

buf-lint: ## Lint the proto contracts.
	go tool buf lint

gen: tools ## Regenerate the gRPC stubs from the proto (commit the result).
	PATH="$(TOOLBIN):$$PATH" go tool buf generate

gen-check: gen ## Fail if regenerating the stubs changes tracked output (deterministic codegen gate).
	@if ! git diff --quiet -- api || [ -n "$$(git ls-files --others --exclude-standard -- api)" ]; then \
		echo "generated code is stale or uncommitted; run: make gen && git add api"; \
		git --no-pager diff -- api; \
		git --no-pager ls-files --others --exclude-standard -- api; \
		exit 1; \
	fi

verify: tidy fmt-check buf-lint gen-check lint vet test race vuln build ## Full ordered safety gate (run before every push and in CI).
	@echo "verify: OK"
