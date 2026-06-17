// Package buildinfo exposes build-time metadata stamped into the binary via
// -ldflags and surfaced in logs and a /version-style endpoint.
//
// The defaults below are intentionally generic so an unstamped `go run` still
// works; release builds override them. See the Dockerfile and Makefile for the
// canonical -X linker flags:
//
//	-X github.com/example/exampleworker/internal/buildinfo.Version=v1.2.3
//	-X github.com/example/exampleworker/internal/buildinfo.Commit=$(git rev-parse HEAD)
package buildinfo

// Build metadata. These are var (not const) so the linker can override them
// with -ldflags -X at build time; they are never mutated at runtime.
var (
	// Name is the service name.
	Name = "exampleworker"
	// Version is the release version (semver tag), or "dev" for local builds.
	Version = "dev"
	// Commit is the VCS revision the binary was built from, or "unknown".
	Commit = "unknown"
)
