package buildinfo_test

import (
	"testing"

	"github.com/example/examplegrpc/internal/buildinfo"
)

func TestDefaultsAreSet(t *testing.T) {
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
