package messaging_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/messaging"
)

func TestMemoryInboxMarkProcessed(t *testing.T) {
	t.Parallel()

	in := messaging.NewMemoryInbox()
	ctx := context.Background()

	already, err := in.MarkProcessed(ctx, "id1")
	if err != nil {
		t.Fatalf("first mark: %v", err)
	}
	if already {
		t.Error("first MarkProcessed should report not-already-processed")
	}

	already, err = in.MarkProcessed(ctx, "id1")
	if err != nil {
		t.Fatalf("second mark: %v", err)
	}
	if !already {
		t.Error("second MarkProcessed should report already-processed")
	}

	seen, err := in.Seen(ctx, "id1")
	if err != nil || !seen {
		t.Errorf("Seen(id1) = %v, %v; want true, nil", seen, err)
	}
}

func TestMemoryDLQAdd(t *testing.T) {
	t.Parallel()

	q := messaging.NewMemoryDLQ()
	dl := messaging.DeadLetter{
		Message:        messaging.Message{ID: "m1", Type: "e"},
		Topic:          "t",
		Attempts:       3,
		FailureClass:   messaging.FailureExhausted,
		Reason:         "boom",
		DeadLetteredAt: time.Unix(0, 0).UTC(),
	}
	if err := q.Add(context.Background(), dl); err != nil {
		t.Fatalf("add: %v", err)
	}
	if q.Len() != 1 {
		t.Fatalf("len = %d, want 1", q.Len())
	}
	entries := q.Entries()
	if entries[0].Message.ID != "m1" || entries[0].FailureClass != messaging.FailureExhausted {
		t.Errorf("entry = %+v", entries[0])
	}

	// Entries returns a copy: mutating it must not affect the store.
	entries[0].Reason = "mutated"
	if q.Entries()[0].Reason == "mutated" {
		t.Error("Entries must return a copy")
	}
}

func TestMemoryBrokerNackRedelivers(t *testing.T) {
	t.Parallel()

	b := messaging.NewMemoryBroker(messaging.WithBuffer(4))
	// t.Context() is cancelled when the test ends, so the broker's relay
	// goroutine stops with the test; the package's goleak TestMain fails
	// anything still running afterwards.
	ctx := t.Context()
	sub, err := b.Subscribe(ctx, "t")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := b.Publish(ctx, "t", messaging.Message{ID: "m1"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	first := <-sub
	if first.ID != "m1" {
		t.Fatalf("first delivery = %q, want m1", first.ID)
	}
	// Nack re-enqueues for redelivery (at-least-once).
	if err := first.Nack(); err != nil {
		t.Fatalf("nack: %v", err)
	}
	select {
	case second := <-sub:
		if second.ID != "m1" {
			t.Errorf("redelivery = %q, want m1", second.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("nack did not redeliver")
	}
}

func TestMemoryBrokerHealthAndClose(t *testing.T) {
	t.Parallel()

	b := messaging.NewMemoryBroker()
	if !b.Healthy() {
		t.Error("new broker should be healthy")
	}
	b.SetHealthy(false)
	if b.Healthy() {
		t.Error("broker should report unhealthy after SetHealthy(false)")
	}
	b.SetHealthy(true)
	if err := b.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if b.Healthy() {
		t.Error("closed broker should report unhealthy")
	}
	if err := b.Publish(context.Background(), "t", messaging.Message{ID: "x"}); err == nil {
		t.Error("publish to closed broker should fail")
	}
	// Close is idempotent.
	if err := b.Close(); err != nil {
		t.Errorf("second close: %v", err)
	}
}

func TestMemoryOutboxOrderAndMarkSent(t *testing.T) {
	t.Parallel()

	o := messaging.NewMemoryOutbox()
	ctx := context.Background()
	for _, id := range []string{"a", "b", "c"} {
		if err := o.Add(ctx, messaging.OutboxRecord{ID: id, Topic: "t", Type: "e"}); err != nil {
			t.Fatalf("add %s: %v", id, err)
		}
	}
	pending, err := o.Pending(ctx, 2)
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(pending) != 2 || pending[0].ID != "a" || pending[1].ID != "b" {
		t.Fatalf("pending = %+v, want [a b]", pending)
	}
	if err := o.MarkSent(ctx, "a", time.Unix(1, 0).UTC()); err != nil {
		t.Fatalf("mark sent: %v", err)
	}
	if o.PendingCount() != 2 {
		t.Errorf("pending count = %d, want 2", o.PendingCount())
	}
	if err := o.MarkSent(ctx, "missing", time.Unix(1, 0).UTC()); err == nil {
		t.Error("mark sent on unknown id should error")
	}
}

func TestBackoffSleepWaiterZero(t *testing.T) {
	t.Parallel()
	// A zero/negative delay returns immediately with the context error (nil).
	if err := (messaging.SleepWaiter{}).Wait(context.Background(), 0); err != nil {
		t.Errorf("Wait(0) = %v, want nil", err)
	}
}

func TestBackoffSleepWaiterCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (messaging.SleepWaiter{}).Wait(ctx, time.Hour); err == nil {
		t.Error("Wait with cancelled context should return ctx.Err()")
	}
}
