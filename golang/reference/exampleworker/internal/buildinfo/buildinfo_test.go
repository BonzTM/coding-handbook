package buildinfo_test

import (
	"testing"

	"github.com/example/exampleworker/internal/buildinfo"
)

// TestDefaults asserts the unstamped defaults so an unstamped `go run` is
// observable and a release build's -ldflags override is the only way these
// change.
func TestDefaults(t *testing.T) {
	t.Parallel()

	if buildinfo.Name == "" {
		t.Error("Name must not be empty")
	}
	if buildinfo.Version == "" {
		t.Error("Version must not be empty")
	}
	if buildinfo.Commit == "" {
		t.Error("Commit must not be empty")
	}
}
