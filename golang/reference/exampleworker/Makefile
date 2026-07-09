# Makefile - single canonical verification entrypoint.
#
# Adapted copy of golang/templates/Makefile for the exampleworker reference.
# Humans and CI run the SAME targets here; do not restate raw command lists
# elsewhere. `make verify` is the ordered safety gate.
#
# Linter config and policy:   golang/quality/linting.md
# CI and release wiring:      golang/operations/ci-and-release.md
#
# Conventions:
#   - Language baseline Go 1.24+; tools are run via `go tool <name>` (added with
#     `go get -tool`).
#   - golangci-lint and govulncheck are tool dependencies, never global installs.
#   - -trimpath belongs to release builds only, never the routine `build` target.

# Use bash with strict flags so multi-command recipes fail loudly.
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

.PHONY: help tidy tidy-check fmt fmt-check lint vet test race cover vuln build run verify

help: ## Show this help.
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "} {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

tidy: ## Sync go.mod/go.sum in place and verify the module graph.
	go mod tidy
	go mod verify

tidy-check: ## Fail if go mod tidy would change go.mod/go.sum (CI-safe, no writes).
	go mod tidy -diff
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

build: ## Compile all packages.
	go build ./...

run: ## Run the worker locally (in-memory broker by default).
	go run ./cmd/exampleworker

verify: tidy-check fmt-check lint vet test race vuln build ## Full ordered safety gate (run before every push and in CI).
	@echo "verify: OK"
