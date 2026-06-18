package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/example/exampleworker/internal/core"
	"github.com/example/exampleworker/internal/telemetry"
)

// Decoder turns a raw message body into a domain event. It is a seam so the
// consumer is payload-format-neutral; the reference uses JSON.
type Decoder func(msg Message) (core.WidgetEvent, error)

// JSONDecoder decodes the message body as a JSON WidgetEvent. A malformed body
// is a non-retryable ErrInvalidEvent: replaying unparseable bytes will never
// succeed, so it moves to the DLQ rather than retrying forever.
func JSONDecoder(msg Message) (core.WidgetEvent, error) {
	var e core.WidgetEvent
	if err := json.Unmarshal(msg.Body, &e); err != nil {
		return core.WidgetEvent{}, fmt.Errorf("%w: decode body: %s", core.ErrInvalidEvent, err.Error())
	}
	return e, nil
}

// ConsumerConfig is the consume loop's retry/backoff policy and dependencies.
type ConsumerConfig struct {
	// Topic is the subscription topic.
	Topic string
	// MaxAttempts is the total delivery attempts (including the first) before a
	// message is dead-lettered. Must be >= 1.
	MaxAttempts int
	// BaseBackoff and MaxBackoff bound the exponential backoff ceiling.
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

// Consumer is the consume loop bound to a context. It subscribes to a topic and
// for each message: decodes, dedupes via the inbox (exactly-once), calls the
// domain Processor, and settles. A transient failure retries IN-PLACE with
// bounded exponential backoff + full jitter computed from the injected clock
// (no real sleeps in tests); a non-retryable failure or an exhausted retry
// budget dead-letters the message. Every settled message is Ack'd so the broker
// does not redeliver it after the consumer has made its terminal decision.
//
// Idempotency contract: the inbox is recorded BEFORE the side effect is
// considered durable here, and processing is guarded so a duplicate delivery of
// an already-processed id is dropped without re-invoking the Processor.
type Consumer struct {
	broker    Broker
	processor core.Processor
	inbox     InboxStore
	dlq       DeadLetterStore
	clock     core.Clock
	waiter    Waiter
	decode    Decoder
	metrics   telemetry.Metrics
	logger    *slog.Logger
	cfg       ConsumerConfig
	backoff   backoffPolicy

	// inflight is incremented while a message is being processed and decremented
	// when it settles, so a test can observe the graceful drain finishing
	// in-flight work. It is atomic so a probe/test can read it concurrently with
	// the single consume goroutine that mutates it.
	inflight atomic.Int64
}

// ConsumerDeps bundles the consumer's collaborators so the constructor stays
// stable as the layer grows.
type ConsumerDeps struct {
	Broker    Broker
	Processor core.Processor
	Inbox     InboxStore
	DLQ       DeadLetterStore
	Clock     core.Clock
	// Waiter pauses between retries. Production wires SleepWaiter; tests wire a
	// clock-driven fake so no real time passes.
	Waiter Waiter
	// Decode turns a message into a domain event. nil defaults to JSONDecoder.
	Decode Decoder
	// Metrics records consume outcomes. nil defaults to NopMetrics.
	Metrics telemetry.Metrics
	// Logger is required.
	Logger *slog.Logger
	// Rand is the jitter source. nil uses the package default.
	Rand *rand.Rand
}

// NewConsumer constructs a Consumer. Required deps (Broker, Processor, Inbox,
// DLQ, Clock, Waiter, Logger) must be non-nil.
func NewConsumer(cfg ConsumerConfig, deps ConsumerDeps) *Consumer {
	decode := deps.Decode
	if decode == nil {
		decode = JSONDecoder
	}
	var metrics telemetry.Metrics = telemetry.NopMetrics{}
	if deps.Metrics != nil {
		metrics = deps.Metrics
	}
	return &Consumer{
		broker:    deps.Broker,
		processor: deps.Processor,
		inbox:     deps.Inbox,
		dlq:       deps.DLQ,
		clock:     deps.Clock,
		waiter:    deps.Waiter,
		decode:    decode,
		metrics:   metrics,
		logger:    deps.Logger,
		cfg:       cfg,
		backoff: backoffPolicy{
			base: cfg.BaseBackoff,
			max:  cfg.MaxBackoff,
			rng:  deps.Rand,
		},
	}
}

// Run subscribes and consumes until ctx is cancelled or the broker closes the
// subscription. It is the errgroup member for the consume loop. Cancelling ctx
// stops pulling NEW messages (the subscription channel closes and the range
// ends); a message already pulled is finished and settled before Run returns,
// which is the graceful-drain contract.
func (c *Consumer) Run(ctx context.Context) error {
	msgs, err := c.broker.Subscribe(ctx, c.cfg.Topic)
	if err != nil {
		return fmt.Errorf("subscribe %q: %w", c.cfg.Topic, err)
	}
	c.logger.InfoContext(ctx, "consumer started", "topic", c.cfg.Topic, "max_attempts", c.cfg.MaxAttempts)

	for msg := range msgs {
		// Drain semantics: finish the message we already pulled even if ctx was
		// cancelled mid-range, using a detached context so the terminal settle
		// (ack + inbox + dlq) is not aborted by the cancelled drain context. A
		// real broker bounds this with its visibility timeout; main bounds the
		// whole drain with ShutdownGrace.
		c.handle(ctx, msg)
	}
	c.logger.InfoContext(ctx, "consumer stopped", "topic", c.cfg.Topic)
	return nil
}

// handle processes one delivery to a terminal outcome (ack-after-process,
// ack-after-dedupe-drop, or ack-after-dead-letter). It Acks in every terminal
// case so the broker does not redeliver a message the consumer has already
// decided on.
func (c *Consumer) handle(ctx context.Context, msg Message) {
	c.inflight.Add(1)
	defer c.inflight.Add(-1)

	// Decode first. A decode failure is non-retryable: dead-letter immediately.
	event, err := c.decode(msg)
	if err != nil {
		c.deadLetter(ctx, msg, 1, FailureInvalid, err)
		return
	}

	// Bounded retry loop. Attempt is 1-based. A non-retryable error breaks to the
	// DLQ; a transient error waits (jittered backoff) and retries until the
	// budget is exhausted, then dead-letters.
	var lastErr error
	for attempt := 1; attempt <= c.cfg.MaxAttempts; attempt++ {
		procErr := c.processOnce(ctx, msg, event)
		if procErr == nil {
			c.metrics.IncConsumed(msg.Type, "ack")
			c.ack(ctx, msg)
			return
		}
		// Context cancellation during processing is not a message failure: stop
		// retrying and Nack so the message is redelivered (not lost, not DLQ'd).
		if errors.Is(procErr, context.Canceled) || errors.Is(procErr, context.DeadlineExceeded) {
			c.nack(ctx, msg)
			return
		}
		lastErr = procErr

		// Non-retryable (validation/schema): dead-letter now, do not burn the
		// remaining budget.
		if errors.Is(procErr, core.ErrInvalidEvent) {
			c.deadLetter(ctx, msg, attempt, FailureInvalid, procErr)
			return
		}

		// Transient: if attempts remain, wait the jittered backoff then retry.
		if attempt < c.cfg.MaxAttempts {
			delay := c.backoff.delay(attempt)
			c.metrics.IncConsumed(msg.Type, "retry")
			c.logger.WarnContext(ctx, "retrying message",
				"message_id", msg.ID, "event_type", msg.Type,
				"attempt", attempt, "backoff", delay.String(), "error", procErr.Error(),
			)
			if werr := c.waiter.Wait(ctx, delay); werr != nil {
				// Drain/cancel during backoff: Nack for redelivery, do not DLQ.
				c.nack(ctx, msg)
				return
			}
		}
	}

	// Budget exhausted on transient failures.
	c.deadLetter(ctx, msg, c.cfg.MaxAttempts, FailureExhausted, lastErr)
}

// processOnce runs the inbox dedupe guard and, on first sight, the domain
// Processor. A duplicate delivery (id already recorded) returns nil without
// re-invoking Process, which is the exactly-once guarantee. The dedupe record
// is written before Process so a crash after Process but before ack still
// drops the redelivery — matching the inbox-in-the-same-transaction contract a
// SQL implementation provides.
func (c *Consumer) processOnce(ctx context.Context, msg Message, event core.WidgetEvent) error {
	already, err := c.inbox.MarkProcessed(ctx, msg.ID)
	if err != nil {
		// Inbox failure is transient (the store is a dependency): let the retry
		// loop handle it.
		return fmt.Errorf("inbox mark: %w", err)
	}
	if already {
		c.metrics.IncConsumed(msg.Type, "dropped_duplicate")
		c.logger.InfoContext(ctx, "duplicate delivery dropped",
			"message_id", msg.ID, "event_type", msg.Type)
		return nil
	}
	if err := c.processor.Process(ctx, event); err != nil {
		// The id was recorded but the side effect failed. Roll the dedupe record
		// back so a retry re-attempts the domain logic; an exhausted/invalid
		// message is dead-lettered and never re-seen. (A SQL inbox achieves this
		// by writing the record and the side effect in one transaction that rolls
		// back together.)
		c.rollbackInbox(ctx, msg.ID)
		return err
	}
	return nil
}

// rollbackInbox best-effort removes a dedupe record after a failed side effect
// so a retry is not mistaken for a duplicate. The in-memory inbox exposes this
// via a type assertion; a SQL inbox rolls back the transaction instead.
func (c *Consumer) rollbackInbox(ctx context.Context, id string) {
	type remover interface {
		remove(id string)
	}
	if r, ok := c.inbox.(remover); ok {
		r.remove(id)
		return
	}
	_ = ctx
}

func (c *Consumer) ack(ctx context.Context, msg Message) {
	if err := msg.Ack(); err != nil {
		c.logger.ErrorContext(ctx, "ack failed", "message_id", msg.ID, "error", err.Error())
	}
}

func (c *Consumer) nack(ctx context.Context, msg Message) {
	if err := msg.Nack(); err != nil {
		c.logger.ErrorContext(ctx, "nack failed", "message_id", msg.ID, "error", err.Error())
	}
}

// deadLetter parks the message and Acks it so the broker does not redeliver a
// message the consumer has terminally given up on. A DLQ write failure is
// logged; the message is still Ack'd because re-delivering it would loop.
func (c *Consumer) deadLetter(ctx context.Context, msg Message, attempts int, class string, cause error) {
	reason := ""
	if cause != nil {
		reason = cause.Error()
	}
	dl := DeadLetter{
		Message:        msg,
		Topic:          c.cfg.Topic,
		Attempts:       attempts,
		FailureClass:   class,
		Reason:         reason,
		DeadLetteredAt: c.clock.Now().UTC(),
	}
	if err := c.dlq.Add(ctx, dl); err != nil {
		c.logger.ErrorContext(ctx, "dead-letter write failed",
			"message_id", msg.ID, "error", err.Error())
	}
	c.metrics.IncConsumed(msg.Type, "dead_lettered")
	c.logger.WarnContext(ctx, "message dead-lettered",
		"message_id", msg.ID, "event_type", msg.Type,
		"attempts", attempts, "failure_class", class, "reason", reason)
	c.ack(ctx, msg)
}

// InFlight reports the number of messages currently being processed. It is
// safe to read concurrently (used by tests to observe the drain and could back
// a gauge metric).
func (c *Consumer) InFlight() int { return int(c.inflight.Load()) }
