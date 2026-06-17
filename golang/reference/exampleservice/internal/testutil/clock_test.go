package testutil_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/testutil"
)

// FakeClock must satisfy the production core.Clock seam.
var _ core.Clock = (*testutil.FakeClock)(nil)

func TestFakeClockSetAndAdvance(t *testing.T) {
	start := time.Unix(1700000000, 0).UTC()
	c := testutil.NewFakeClock(start)

	if got := c.Now(); !got.Equal(start) {
		t.Fatalf("Now() = %v, want %v", got, start)
	}

	c.Advance(90 * time.Second)
	if got, want := c.Now(), start.Add(90*time.Second); !got.Equal(want) {
		t.Errorf("after Advance: Now() = %v, want %v", got, want)
	}

	c.Advance(-30 * time.Second)
	if got, want := c.Now(), start.Add(60*time.Second); !got.Equal(want) {
		t.Errorf("after negative Advance: Now() = %v, want %v", got, want)
	}

	reset := time.Unix(1800000000, 0).UTC()
	c.Set(reset)
	if got := c.Now(); !got.Equal(reset) {
		t.Errorf("after Set: Now() = %v, want %v", got, reset)
	}
}

func TestFakeClockZeroValue(t *testing.T) {
	var c testutil.FakeClock
	if got := c.Now(); !got.IsZero() {
		t.Errorf("zero-value Now() = %v, want zero time", got)
	}
}

// TestFakeClockConcurrent exercises the mutex under -race: concurrent readers
// and writers must not data-race.
func TestFakeClockConcurrent(t *testing.T) {
	c := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(2)
		go func() { defer wg.Done(); c.Advance(time.Second) }()
		go func() { defer wg.Done(); _ = c.Now() }()
	}
	wg.Wait()
	if got, want := c.Now(), time.Unix(8, 0).UTC(); !got.Equal(want) {
		t.Errorf("Now() = %v, want %v", got, want)
	}
}
