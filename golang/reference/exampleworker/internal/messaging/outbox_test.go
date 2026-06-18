package messaging_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/messaging"
	"github.com/example/exampleworker/internal/testutil"
)

// TestOutboxRelayPublishesThenMarksSent proves the relay drains pending records,
// publishes them to the broker, and marks them sent (so a second flush is a
// no-op).
func TestOutboxRelayPublishesThenMarksSent(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker(messaging.WithBuffer(8))
	outbox := messaging.NewMemoryOutbox()
	clk := testutil.NewFakeClock(time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))

	for _, id := range []string{"o1", "o2", "o3"} {
		if err := outbox.Add(context.Background(), messaging.OutboxRecord{
			ID: id, Topic: "widget.events", Type: "widget.created", Body: []byte(`{}`), CreatedAt: clk.Now(),
		}); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	relay := messaging.NewRelay(time.Second, 100, messaging.RelayDeps{
		Store: outbox, Broker: broker, Clock: clk, Logger: discardLogger(),
	})

	n, err := relay.Flush(context.Background())
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if n != 3 {
		t.Fatalf("published %d, want 3", n)
	}
	if outbox.PendingCount() != 0 {
		t.Errorf("pending = %d, want 0 after flush", outbox.PendingCount())
	}

	// Messages are on the broker, in order.
	sub, err := broker.Subscribe(context.Background(), "widget.events")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	for _, want := range []string{"o1", "o2", "o3"} {
		select {
		case msg := <-sub:
			if msg.ID != want {
				t.Errorf("got message %q, want %q", msg.ID, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for message %q", want)
		}
	}

	// A second flush publishes nothing (all sent).
	n2, err := relay.Flush(context.Background())
	if err != nil {
		t.Fatalf("second flush: %v", err)
	}
	if n2 != 0 {
		t.Errorf("second flush published %d, want 0", n2)
	}
}

// TestOutboxRelayLeavesPendingOnPublishFailure proves a publish failure leaves
// the record pending (no mark-sent) so the next scan retries it — reliable
// publish, no loss.
func TestOutboxRelayLeavesPendingOnPublishFailure(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker()
	if err := broker.Close(); err != nil { // closed broker -> Publish fails
		t.Fatalf("close: %v", err)
	}
	outbox := messaging.NewMemoryOutbox()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	if err := outbox.Add(context.Background(), messaging.OutboxRecord{ID: "o1", Topic: "t", Type: "e", Body: []byte("{}")}); err != nil {
		t.Fatalf("add: %v", err)
	}

	relay := messaging.NewRelay(time.Second, 10, messaging.RelayDeps{
		Store: outbox, Broker: broker, Clock: clk, Logger: discardLogger(),
	})

	n, err := relay.Flush(context.Background())
	if err == nil {
		t.Fatal("expected publish error against a closed broker")
	}
	if !errors.Is(err, messaging.ErrBrokerClosed) {
		t.Errorf("err = %v, want ErrBrokerClosed", err)
	}
	if n != 0 {
		t.Errorf("published %d, want 0", n)
	}
	if outbox.PendingCount() != 1 {
		t.Errorf("pending = %d, want 1 (left for retry)", outbox.PendingCount())
	}
}

// TestOutboxRelayRunFinalFlush proves Run performs a final flush on context
// cancellation so a record enqueued just before shutdown is still published.
func TestOutboxRelayRunFinalFlush(t *testing.T) {
	t.Parallel()

	broker := messaging.NewMemoryBroker(messaging.WithBuffer(4))
	outbox := messaging.NewMemoryOutbox()
	clk := testutil.NewFakeClock(time.Unix(0, 0).UTC())
	if err := outbox.Add(context.Background(), messaging.OutboxRecord{ID: "late", Topic: "widget.events", Type: "e", Body: []byte("{}")}); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Long poll interval so the only flush is the final one on cancel.
	relay := messaging.NewRelay(time.Hour, 10, messaging.RelayDeps{
		Store: outbox, Broker: broker, Clock: clk, Logger: discardLogger(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- relay.Run(ctx) }()

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("relay Run: %v", err)
	}
	if outbox.PendingCount() != 0 {
		t.Errorf("pending = %d, want 0 after final flush", outbox.PendingCount())
	}
}
