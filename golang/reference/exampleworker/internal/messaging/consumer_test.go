package messaging_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/core"
	"github.com/example/exampleworker/internal/messaging"
	"github.com/example/exampleworker/internal/testutil"
)

// discardLogger returns a logger that drops output so tests stay quiet.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// recordingWaiter is a clock-driven Waiter that NEVER sleeps. It records each
// requested backoff so a test can assert the bounds, advances the supplied fake
// clock by the requested delay (so time-dependent assertions hold), and returns
// immediately. This is how the retry/DLQ test runs without any real sleep.
type recordingWaiter struct {
	clock *testutil.FakeClock

	mu     sync.Mutex
	delays []time.Duration
}

func (w *recordingWaiter) Wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	w.mu.Lock()
	w.delays = append(w.delays, d)
	w.mu.Unlock()
	w.clock.Advance(d)
	return nil
}

func (w *recordingWaiter) recorded() []time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]time.Duration, len(w.delays))
	copy(out, w.delays)
	return out
}

// fakeProcessor lets a test script per-call outcomes.
type fakeProcessor struct {
	mu      sync.Mutex
	calls   int
	results []error // result for call i; last result repeats
	seen    []core.WidgetEvent
}

func (p *fakeProcessor) Process(_ context.Context, e core.WidgetEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seen = append(p.seen, e)
	idx := p.calls
	p.calls++
	if idx >= len(p.results) {
		if len(p.results) == 0 {
			return nil
		}
		return p.results[len(p.results)-1]
	}
	return p.results[idx]
}

func (p *fakeProcessor) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func mustBody(t *testing.T, e core.WidgetEvent) []byte {
	t.Helper()
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	return b
}

func newConsumer(t *testing.T, cfg messaging.ConsumerConfig, deps messaging.ConsumerDeps) *messaging.Consumer {
	t.Helper()
	if deps.Logger == nil {
		deps.Logger = discardLogger()
	}
	return messaging.NewConsumer(cfg, deps)
}

func baseConfig() messaging.ConsumerConfig {
	return messaging.ConsumerConfig{
		Topic:       "widget.events",
		MaxAttempts: 4,
		BaseBackoff: 100 * time.Millisecond,
		MaxBackoff:  time.Second,
	}
}

// TestConsumeProcessAck is the happy path: consume -> process -> ack, projection
// updated, no DLQ.
func TestConsumeProcessAck(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	projector := core.NewWidgetProjector(clk)
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()

	c := newConsumer(t, baseConfig(), messaging.ConsumerDeps{
		Broker: broker, Processor: projector, Inbox: inbox, DLQ: dlq,
		Clock: clk, Waiter: &recordingWaiter{clock: clk},
	})

	ev := core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "gadget"}
	ctx, cancel := context.WithCancel(context.Background())
	if err := broker.Publish(ctx, "widget.events", messaging.Message{ID: "m1", Type: "widget.created", Body: mustBody(t, ev)}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	done := runConsumer(ctx, t, c)
	waitFor(t, func() bool { _, ok := projector.Get("t1", "w1"); return ok })
	cancel()
	<-done

	if dlq.Len() != 0 {
		t.Errorf("DLQ should be empty, got %d", dlq.Len())
	}
	w, ok := projector.Get("t1", "w1")
	if !ok || w.Name != "gadget" {
		t.Errorf("projection = %+v ok=%v", w, ok)
	}
	seen, err := inbox.Seen(context.Background(), "m1")
	if err != nil {
		t.Fatalf("inbox Seen: %v", err)
	}
	if !seen {
		t.Error("inbox should have recorded m1")
	}
}

// TestDuplicateProcessedOnce proves the inbox dedupe: the same message id
// delivered twice invokes the domain Processor exactly once.
func TestDuplicateProcessedOnce(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	proc := &fakeProcessor{}
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()

	c := newConsumer(t, baseConfig(), messaging.ConsumerDeps{
		Broker: broker, Processor: proc, Inbox: inbox, DLQ: dlq,
		Clock: clk, Waiter: &recordingWaiter{clock: clk},
	})

	ev := core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "n"}
	body := mustBody(t, ev)
	ctx, cancel := context.WithCancel(context.Background())
	// Same ID twice (at-least-once duplicate delivery).
	for range 2 {
		if err := broker.Publish(ctx, "widget.events", messaging.Message{ID: "dup", Type: "widget.created", Body: body}); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}

	done := runConsumer(ctx, t, c)
	waitFor(t, func() bool { return proc.callCount() >= 1 })
	// Give the second delivery time to be dropped.
	waitFor(t, func() bool { return c.InFlight() == 0 })
	cancel()
	<-done

	if got := proc.callCount(); got != 1 {
		t.Errorf("Process called %d times, want exactly 1 (dedupe)", got)
	}
}

// TestRetryThenDLQ proves bounded retry with jittered backoff and dead-lettering
// after the budget is exhausted, using a fake clock and NO real sleeps. It also
// asserts each backoff is bounded by the exponential ceiling.
func TestRetryThenDLQ(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	// Always-failing transient processor.
	proc := &fakeProcessor{results: []error{core.ErrTransient}}
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()
	waiter := &recordingWaiter{clock: clk}

	cfg := messaging.ConsumerConfig{Topic: "widget.events", MaxAttempts: 4, BaseBackoff: 100 * time.Millisecond, MaxBackoff: time.Second}
	c := newConsumer(t, cfg, messaging.ConsumerDeps{
		Broker: broker, Processor: proc, Inbox: inbox, DLQ: dlq,
		Clock: clk, Waiter: waiter,
		// Pin jitter to its MAX (ceiling) so the bound assertion is exact and
		// deterministic: full jitter draws in [0, ceiling], so ceiling is the bound.
		Rand: rand.New(rand.NewPCG(1, 2)),
	})

	ev := core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "n"}
	ctx, cancel := context.WithCancel(context.Background())
	if err := broker.Publish(ctx, "widget.events", messaging.Message{ID: "m1", Type: "widget.created", Body: mustBody(t, ev)}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	done := runConsumer(ctx, t, c)
	waitFor(t, func() bool { return dlq.Len() == 1 })
	cancel()
	<-done

	// MaxAttempts attempts means MaxAttempts-1 backoff waits.
	delays := waiter.recorded()
	if len(delays) != cfg.MaxAttempts-1 {
		t.Fatalf("recorded %d backoffs, want %d", len(delays), cfg.MaxAttempts-1)
	}
	// Each backoff bounded by base*2^(attempt-1) capped at max (full jitter).
	ceiling := cfg.BaseBackoff
	for i, d := range delays {
		if d < 0 || d > ceiling {
			t.Errorf("backoff[%d] = %s, want in [0, %s]", i, d, ceiling)
		}
		ceiling *= 2
		if ceiling > cfg.MaxBackoff {
			ceiling = cfg.MaxBackoff
		}
	}

	entries := dlq.Entries()
	if len(entries) != 1 {
		t.Fatalf("DLQ len = %d, want 1", len(entries))
	}
	dl := entries[0]
	if dl.FailureClass != messaging.FailureExhausted {
		t.Errorf("failure class = %q, want %q", dl.FailureClass, messaging.FailureExhausted)
	}
	if dl.Attempts != cfg.MaxAttempts {
		t.Errorf("attempts = %d, want %d", dl.Attempts, cfg.MaxAttempts)
	}
	if dl.Message.ID != "m1" {
		t.Errorf("DLQ message id = %q, want m1", dl.Message.ID)
	}
	if proc.callCount() != cfg.MaxAttempts {
		t.Errorf("Process called %d times, want %d", proc.callCount(), cfg.MaxAttempts)
	}
}

// TestInvalidGoesToDLQImmediately proves a non-retryable validation failure is
// dead-lettered without burning the retry budget.
func TestInvalidGoesToDLQImmediately(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	proc := &fakeProcessor{}
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()
	waiter := &recordingWaiter{clock: clk}

	c := newConsumer(t, baseConfig(), messaging.ConsumerDeps{
		Broker: broker, Processor: proc, Inbox: inbox, DLQ: dlq,
		Clock: clk, Waiter: waiter,
	})

	ctx, cancel := context.WithCancel(context.Background())
	// Malformed JSON body -> decode failure -> non-retryable.
	if err := broker.Publish(ctx, "widget.events", messaging.Message{ID: "bad", Type: "widget.created", Body: []byte("{not json")}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	done := runConsumer(ctx, t, c)
	waitFor(t, func() bool { return dlq.Len() == 1 })
	cancel()
	<-done

	if proc.callCount() != 0 {
		t.Errorf("Process should not be called for a decode failure, got %d", proc.callCount())
	}
	if len(waiter.recorded()) != 0 {
		t.Errorf("no backoff should occur for a non-retryable failure, got %v", waiter.recorded())
	}
	if dlq.Entries()[0].FailureClass != messaging.FailureInvalid {
		t.Errorf("failure class = %q, want %q", dlq.Entries()[0].FailureClass, messaging.FailureInvalid)
	}
}

// TestReplayAfterProcessedIsDropped proves idempotency under replay: replaying a
// message whose id is already in the inbox does not re-invoke the domain.
func TestReplayAfterProcessedIsDropped(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	proc := &fakeProcessor{}
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()

	// Pre-seed the inbox: this id was processed in a prior run (replay scenario).
	if _, err := inbox.MarkProcessed(context.Background(), "replayed"); err != nil {
		t.Fatalf("seed inbox: %v", err)
	}

	c := newConsumer(t, baseConfig(), messaging.ConsumerDeps{
		Broker: broker, Processor: proc, Inbox: inbox, DLQ: dlq,
		Clock: clk, Waiter: &recordingWaiter{clock: clk},
	})

	ev := core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "n"}
	ctx, cancel := context.WithCancel(context.Background())
	if err := broker.Publish(ctx, "widget.events", messaging.Message{ID: "replayed", Type: "widget.created", Body: mustBody(t, ev)}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	done := runConsumer(ctx, t, c)
	waitFor(t, func() bool { return c.InFlight() == 0 })
	cancel()
	<-done

	if proc.callCount() != 0 {
		t.Errorf("replayed message must not re-invoke Process, got %d calls", proc.callCount())
	}
	if dlq.Len() != 0 {
		t.Errorf("replay drop must not dead-letter, got %d", dlq.Len())
	}
}

// runConsumer starts c.Run on a goroutine and returns a channel closed when it
// returns. The consumer's Run exits when the subscription channel closes (ctx
// cancelled), which is the drain trigger.
func runConsumer(ctx context.Context, t *testing.T, c *messaging.Consumer) <-chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := c.Run(ctx); err != nil {
			t.Errorf("consumer Run: %v", err)
		}
	}()
	return done
}

// waitFor polls cond up to a generous deadline. It uses a tiny real poll
// interval only to yield the scheduler; the consumer's backoff itself uses the
// fake clock and never sleeps.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}
