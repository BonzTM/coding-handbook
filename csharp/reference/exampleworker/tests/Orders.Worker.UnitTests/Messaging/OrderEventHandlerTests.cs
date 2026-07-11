using Orders.Worker.Core.Events;
using Orders.Worker.Core.Messaging;
using Orders.Worker.Core.Orders;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Messaging;

/// <summary>
/// The consume pipeline's terminal decisions - ack, duplicate drop, bounded
/// retry then DLQ, immediate DLQ for invalid payloads, nack on drain - all
/// deterministic via FakeTimeProvider + the recording delayer: no real sleeps
/// anywhere (csharp/services/eventing-and-messaging.md, Testing And Proof).
/// </summary>
public sealed class OrderEventHandlerTests
{
    [Fact]
    public async Task ValidEvent_ProcessesUpdatesProjectionAndAcks()
    {
        var harness = new HandlerHarness();
        var projector = new OrderProjector(harness.Time);
        var handler = harness.CreateHandler(projector);
        var envelope = TestEvents.Envelope(TestEvents.Placed(orderId: "order-7", reference: "ref-7"));
        var delivery = new TestDelivery(envelope);

        await handler.HandleAsync(delivery.Message, TestContext.Current.CancellationToken);

        Assert.Equal(1, delivery.Acks);
        Assert.Equal(0, delivery.Nacks);
        var projection = projector.GetProjection("tenant-a", "order-7");
        Assert.NotNull(projection);
        Assert.Equal("ref-7", projection.ExternalReference);
        Assert.True(await harness.Inbox.SeenAsync(envelope.Id, TestContext.Current.CancellationToken));
        Assert.Equal(0, harness.DeadLetters.Count);
    }

    [Fact]
    public async Task DuplicateDelivery_InvokesProcessorExactlyOnce()
    {
        var harness = new HandlerHarness();
        var processor = new ScriptedProcessor();
        var handler = harness.CreateHandler(processor);
        var envelope = TestEvents.Envelope(TestEvents.Placed());
        var first = new TestDelivery(envelope);
        var second = new TestDelivery(envelope);

        await handler.HandleAsync(first.Message, TestContext.Current.CancellationToken);
        await handler.HandleAsync(second.Message, TestContext.Current.CancellationToken);

        Assert.Equal(1, processor.Calls);
        // Both deliveries settle with an ack: the duplicate is dropped, not
        // redelivered forever.
        Assert.Equal(1, first.Acks);
        Assert.Equal(1, second.Acks);
        Assert.Equal(0, harness.DeadLetters.Count);
    }

    [Fact]
    public async Task TransientFailures_RetryWithBoundedBackoff_ThenDeadLetter()
    {
        // Jitter pinned to 1.0: full jitter draws in [0, ceiling], so every
        // recorded delay must equal its exponential ceiling exactly.
        var harness = new HandlerHarness();
        var processor = new ScriptedProcessor(new TransientEventException("store unavailable"));
        var handler = harness.CreateHandler(processor);
        var envelope = TestEvents.Envelope(TestEvents.Placed());
        var delivery = new TestDelivery(envelope);

        await handler.HandleAsync(delivery.Message, TestContext.Current.CancellationToken);

        // MaxAttempts attempts -> MaxAttempts - 1 backoff waits, doubling from
        // BaseBackoff and capped at MaxBackoff.
        Assert.Equal(harness.Options.MaxAttempts, processor.Calls);
        Assert.Equal(
            [TimeSpan.FromMilliseconds(100), TimeSpan.FromMilliseconds(200), TimeSpan.FromMilliseconds(400)],
            harness.Delayer.Delays);

        var deadLetter = Assert.Single(harness.DeadLetters.Snapshot());
        Assert.Equal(envelope.Id, deadLetter.Envelope.Id);
        Assert.Equal(DeadLetterFailureClass.Exhausted, deadLetter.FailureClass);
        Assert.Equal(harness.Options.MaxAttempts, deadLetter.Attempts);
        Assert.Contains("store unavailable", deadLetter.Reason, StringComparison.Ordinal);
        // Dead-lettered AFTER the backoff waits: the timestamp comes from the
        // fake clock the delayer advanced (100 + 200 + 400 ms).
        Assert.Equal(TestEvents.Start + TimeSpan.FromMilliseconds(700), deadLetter.DeadLetteredAt);
        // Terminally settled: acked so the broker never redelivers it.
        Assert.Equal(1, delivery.Acks);
        Assert.Equal(0, delivery.Nacks);
    }

    [Fact]
    public async Task MalformedPayload_DeadLettersImmediatelyWithoutProcessorOrRetry()
    {
        var harness = new HandlerHarness();
        var processor = new ScriptedProcessor();
        var handler = harness.CreateHandler(processor);
        var envelope = TestEvents.Malformed();
        var delivery = new TestDelivery(envelope);

        await handler.HandleAsync(delivery.Message, TestContext.Current.CancellationToken);

        Assert.Equal(0, processor.Calls);
        Assert.Empty(harness.Delayer.Delays);
        var deadLetter = Assert.Single(harness.DeadLetters.Snapshot());
        Assert.Equal(DeadLetterFailureClass.Invalid, deadLetter.FailureClass);
        Assert.Equal(1, deadLetter.Attempts);
        Assert.Equal(1, delivery.Acks);
    }

    [Fact]
    public async Task InvalidEvent_DeadLettersImmediatelyAndCompensatesInbox()
    {
        var harness = new HandlerHarness();
        // Real domain validation: a placed event without an external
        // reference is structurally invalid.
        var projector = new OrderProjector(harness.Time);
        var handler = harness.CreateHandler(projector);
        var envelope = TestEvents.Envelope(TestEvents.Placed(reference: ""));
        var delivery = new TestDelivery(envelope);

        await handler.HandleAsync(delivery.Message, TestContext.Current.CancellationToken);

        Assert.Empty(harness.Delayer.Delays);
        var deadLetter = Assert.Single(harness.DeadLetters.Snapshot());
        Assert.Equal(DeadLetterFailureClass.Invalid, deadLetter.FailureClass);
        Assert.Contains("external reference", deadLetter.Reason, StringComparison.Ordinal);
        Assert.Equal(1, delivery.Acks);
        // The failed attempt's dedupe record was compensated away - the SQL
        // inbox's rollback semantics.
        Assert.False(await harness.Inbox.SeenAsync(envelope.Id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task ReplayedMessage_IsDroppedWithoutReinvokingTheDomain()
    {
        var harness = new HandlerHarness();
        var processor = new ScriptedProcessor();
        var handler = harness.CreateHandler(processor);
        var envelope = TestEvents.Envelope(TestEvents.Placed());
        // This id was processed in a prior run (operator replay scenario).
        await harness.Inbox.MarkProcessedAsync(envelope.Id, TestContext.Current.CancellationToken);
        var delivery = new TestDelivery(envelope);

        await handler.HandleAsync(delivery.Message, TestContext.Current.CancellationToken);

        Assert.Equal(0, processor.Calls);
        Assert.Equal(0, harness.DeadLetters.Count);
        Assert.Equal(1, delivery.Acks);
    }

    [Fact]
    public async Task CancellationDuringProcessing_NacksForRedelivery_NeverDeadLetters()
    {
        var harness = new HandlerHarness();
        using var cts = new CancellationTokenSource();
        var processor = new CallbackProcessor(async (_, _) =>
        {
            await cts.CancelAsync();
            throw new OperationCanceledException("processing observed the drain token");
        });
        var handler = harness.CreateHandler(processor);
        var envelope = TestEvents.Envelope(TestEvents.Placed());
        var delivery = new TestDelivery(envelope);

        await handler.HandleAsync(delivery.Message, cts.Token);

        Assert.Equal(1, processor.Calls);
        Assert.Equal(0, delivery.Acks);
        Assert.Equal(1, delivery.Nacks);
        Assert.Equal(0, harness.DeadLetters.Count);
        // The dedupe record was compensated: the redelivery must process.
        Assert.False(await harness.Inbox.SeenAsync(envelope.Id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task CancellationDuringBackoff_NacksForRedelivery()
    {
        var harness = new HandlerHarness();
        using var cts = new CancellationTokenSource();
        // First attempt fails transiently AND triggers the drain, so the
        // backoff wait observes a cancelled token.
        var processor = new CallbackProcessor(async (_, _) =>
        {
            await cts.CancelAsync();
            throw new TransientEventException("flaky");
        });
        var handler = harness.CreateHandler(processor);
        var delivery = new TestDelivery(TestEvents.Envelope(TestEvents.Placed()));

        await handler.HandleAsync(delivery.Message, cts.Token);

        Assert.Equal(1, processor.Calls);
        Assert.Equal(0, delivery.Acks);
        Assert.Equal(1, delivery.Nacks);
        Assert.Equal(0, harness.DeadLetters.Count);
    }
}
