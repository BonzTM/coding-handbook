package testutil_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/testutil"
)

func TestFakeClockSetAndAdvance(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	c := testutil.NewFakeClock(start)

	if got := c.Now(); !got.Equal(start) {
		t.Fatalf("Now() = %s, want %s", got, start)
	}

	c.Advance(90 * time.Second)
	if got := c.Now(); !got.Equal(start.Add(90 * time.Second)) {
		t.Fatalf("after Advance Now() = %s, want %s", got, start.Add(90*time.Second))
	}

	reset := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c.Set(reset)
	if got := c.Now(); !got.Equal(reset) {
		t.Fatalf("after Set Now() = %s, want %s", got, reset)
	}
}

// TestFakeClockConcurrent exercises the mutex under the race detector.
func TestFakeClockConcurrent(t *testing.T) {
	t.Parallel()

	c := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			c.Advance(time.Second)
			_ = c.Now()
		})
	}
	wg.Wait()

	if got := c.Now(); !got.Equal(time.Unix(50, 0).UTC()) {
		t.Fatalf("Now() = %s, want %s", got, time.Unix(50, 0).UTC())
	}
}
