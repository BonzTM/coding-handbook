package messaging

import (
	"context"
	"math/rand/v2"
	"time"
)

// backoffPolicy computes the retry wait for an attempt using exponential
// backoff with FULL jitter, per golang/services/eventing-and-messaging.md
// ### Retries And Dead-Letter Behavior ("bounded exponential backoff with
// jitter"). Full jitter draws uniformly in [0, ceiling], where the ceiling
// doubles each attempt up to max. The randomness source is injected so tests
// are deterministic.
type backoffPolicy struct {
	base time.Duration
	max  time.Duration
	// rng is the jitter source; injected so tests can pin it. nil uses a
	// per-call default in production wiring.
	rng *rand.Rand
}

// ceiling returns the un-jittered exponential ceiling for a 1-based attempt
// number: base * 2^(attempt-1), capped at max. attempt is clamped to >= 1.
func (p backoffPolicy) ceiling(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := p.base
	for range attempt - 1 {
		d *= 2
		if d >= p.max {
			return p.max
		}
	}
	if d > p.max {
		return p.max
	}
	return d
}

// delay returns the jittered wait for a 1-based attempt number: a uniform draw
// in [0, ceiling(attempt)]. The result is therefore always bounded by the
// exponential ceiling, which the retry-then-DLQ test asserts.
func (p backoffPolicy) delay(attempt int) time.Duration {
	ceil := p.ceiling(attempt)
	if ceil <= 0 {
		return 0
	}
	if p.rng != nil {
		return time.Duration(p.rng.Int64N(int64(ceil) + 1))
	}
	return time.Duration(rand.Int64N(int64(ceil) + 1)) //nolint:gosec // jitter, not security-sensitive
}

// Waiter pauses for a duration or until the context is cancelled. It is the
// seam that keeps real sleeps out of tests: production wires SleepWaiter, tests
// wire a clock-driven fake that returns immediately while recording the
// requested delay so backoff bounds can be asserted without wall-clock waits.
type Waiter interface {
	// Wait blocks for d or until ctx is cancelled, whichever comes first. It
	// returns ctx.Err() if the context is cancelled before d elapses.
	Wait(ctx context.Context, d time.Duration) error
}

// SleepWaiter is the production Waiter backed by a real timer. It uses a
// context-aware sleep so a cancelled drain stops waiting immediately.
type SleepWaiter struct{}

// Wait sleeps for d or returns early if ctx is cancelled.
func (SleepWaiter) Wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
