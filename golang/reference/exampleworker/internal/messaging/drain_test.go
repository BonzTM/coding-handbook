package messaging_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/core"
	"github.com/example/exampleworker/internal/messaging"
	"github.com/example/exampleworker/internal/testutil"
)

// blockingProcessor blocks the first Process call until released, so a test can
// hold a message in-flight, trigger drain (cancel the consume context), and
// assert the in-flight message still finishes and is acked — no message lost.
type blockingProcessor struct {
	started chan struct{}
	release chan struct{}

	mu        sync.Mutex
	processed []string
}

func (p *blockingProcessor) Process(_ context.Context, e core.WidgetEvent) error {
	select {
	case p.started <- struct{}{}:
	default:
	}
	<-p.release
	p.mu.Lock()
	p.processed = append(p.processed, e.WidgetID)
	p.mu.Unlock()
	return nil
}

func (p *blockingProcessor) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.processed)
}

// TestGracefulDrainFinishesInFlight proves the ordered drain: cancelling the
// consume context stops pulling NEW messages, but a message already pulled is
// finished and acked before Run returns. No message is lost.
func TestGracefulDrainFinishesInFlight(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker(messaging.WithBuffer(4))
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	proc := &blockingProcessor{started: make(chan struct{}, 1), release: make(chan struct{})}
	inbox := messaging.NewMemoryInbox()
	dlq := messaging.NewMemoryDLQ()

	c := newConsumer(t, baseConfig(), messaging.ConsumerDeps{
		Broker: broker, Processor: proc, Inbox: inbox, DLQ: dlq,
		Clock: clk, Waiter: &recordingWaiter{clock: clk},
	})

	ev := core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "inflight", TenantID: "t1", Name: "n"}
	ctx, cancel := context.WithCancel(context.Background())
	if err := broker.Publish(ctx, "widget.events", messaging.Message{ID: "m1", Type: "widget.created", Body: mustBody(t, ev)}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	done := runConsumer(ctx, t, c)

	// Wait until the message is being processed (in-flight), THEN trigger drain.
	<-proc.started
	cancel()

	// The consumer must not return while the in-flight message is unfinished.
	select {
	case <-done:
		t.Fatal("consumer returned before in-flight message finished")
	case <-time.After(50 * time.Millisecond):
	}

	// Release the in-flight message; the consumer finishes it and then drains.
	close(proc.release)
	<-done

	if proc.count() != 1 {
		t.Errorf("processed %d messages, want 1 (in-flight finished)", proc.count())
	}
	if c.InFlight() != 0 {
		t.Errorf("InFlight = %d after drain, want 0", c.InFlight())
	}
	if dlq.Len() != 0 {
		t.Errorf("DLQ = %d, want 0 (no message lost or failed)", dlq.Len())
	}
	seen, err := inbox.Seen(context.Background(), "m1")
	if err != nil {
		t.Fatalf("inbox Seen: %v", err)
	}
	if !seen {
		t.Error("in-flight message should be recorded in the inbox after drain")
	}
}
