# Makefile - convenience shim ONLY. The canonical gate is verify.ps1, run
# with PowerShell 7 (pwsh) so the same script works on Linux, macOS, and
# Windows. Windows users without make call it directly:
#
#     pwsh ./verify.ps1
#
# Do NOT add real logic here; every target is one line delegating to the
# script. Stages, ordering, and policy live in verify.ps1 and
# csharp/operations/ci-and-release.md.

.DEFAULT_GOAL := verify

.PHONY: verify verify-integration format build test

verify: ## Full ordered gate: restore(locked), format-check, build(warnaserror), test, audit.
	pwsh ./verify.ps1

verify-integration: ## Full gate plus tests/*IntegrationTests* (needs Docker).
	pwsh ./verify.ps1 -Integration

format: ## Fix formatting in place (dotnet format).
	pwsh ./verify.ps1 -Stage format -Fix

build: ## Restore (locked) and build Release with warnings as errors.
	pwsh ./verify.ps1 -Stage build

test: ## Restore, build, and run unit tests.
	pwsh ./verify.ps1 -Stage test
