package messaging_test

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain gives the whole package goroutine-leak coverage, per
// golang/quality/testing.md ### Leak Detection: this package OWNS long-lived
// goroutines (the consumer run loop, the broker's subscription relay, the
// outbox relay), so every test must prove shutdown actually stops what Run
// started. After the last test, goleak fails the package if anything is still
// parked on a channel, timer, or blocking read.
//
// There is deliberately no allowlist: every goroutine spawned here is ours.
// Add IgnoreTopFunction entries only for framework goroutines the package does
// not own, and comment why — an open-ended allowlist defeats the check.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
