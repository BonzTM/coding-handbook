# Makefile - single canonical verification entrypoint.
#
# Humans and CI run the SAME targets here; do not restate raw command lists.
# `make verify` is the ordered safety gate defined by the Python handbook.

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

.PHONY: help lock-check sync format format-check fmt fmt-check lint imports types test test-integration audit verify

help: ## Show this help.
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "} {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

lock-check: ## Fail when uv.lock disagrees with pyproject.toml.
	uv lock --check

sync: ## Install the committed graph without resolving or rewriting it.
	uv sync --frozen

format: ## Format Python source in place.
	uv run ruff format .

format-check: ## Fail when Python source is not formatted.
	uv run ruff format --check .

fmt: format ## Alias for format.

fmt-check: format-check ## Alias for format-check.

lint: ## Run curated Ruff correctness and security policy.
	uv run ruff check .

imports: ## Enforce architectural import contracts.
	uv run lint-imports

types: ## Run strict mypy.
	uv run mypy .

test: ## Run the configured ordinary pytest suite.
	uv run pytest

test-integration: ## Run tests that require real external dependencies.
	uv run pytest -m integration

audit: ## Audit the resolved environment without adding a persistent dependency.
	uv run --with pip-audit pip-audit

verify: lock-check sync format-check lint imports types test audit ## Full ordered safety gate.
	@echo "verify: OK"
