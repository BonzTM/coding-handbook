package testutil_test

import (
	"testing"
	"time"

	"github.com/example/examplegrpc/internal/testutil"
)

func TestFakeClockSetAndAdvance(t *testing.T) {
	base := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	c := testutil.NewFakeClock(base)

	if got := c.Now(); !got.Equal(base) {
		t.Fatalf("Now() = %v, want %v", got, base)
	}

	c.Advance(90 * time.Minute)
	if got := c.Now(); !got.Equal(base.Add(90 * time.Minute)) {
		t.Fatalf("after Advance Now() = %v", got)
	}

	next := base.Add(24 * time.Hour)
	c.Set(next)
	if got := c.Now(); !got.Equal(next) {
		t.Fatalf("after Set Now() = %v, want %v", got, next)
	}
}
