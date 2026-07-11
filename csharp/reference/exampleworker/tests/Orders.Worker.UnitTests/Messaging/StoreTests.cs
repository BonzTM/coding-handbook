using Orders.Worker.Core.Messaging;
using Orders.Worker.Infrastructure.Messaging;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Messaging;

/// <summary>
/// Contracts of the in-memory inbox/outbox/DLQ stores - the same behavior the
/// documented SQL implementations provide through unique constraints and
/// transactions.
/// </summary>
public sealed class StoreTests
{
    [Fact]
    public async Task Inbox_FirstMarkRecordsSecondMarkReportsDuplicate()
    {
        var inbox = new InMemoryInboxStore();
        var id = Guid.NewGuid();

        Assert.False(await inbox.MarkProcessedAsync(id, TestContext.Current.CancellationToken));
        Assert.True(await inbox.MarkProcessedAsync(id, TestContext.Current.CancellationToken));
        Assert.True(await inbox.SeenAsync(id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task Inbox_RemoveCompensatesSoRetryIsNotADuplicate()
    {
        var inbox = new InMemoryInboxStore();
        var id = Guid.NewGuid();
        _ = await inbox.MarkProcessedAsync(id, TestContext.Current.CancellationToken);

        await inbox.RemoveAsync(id, TestContext.Current.CancellationToken);

        Assert.False(await inbox.SeenAsync(id, TestContext.Current.CancellationToken));
        Assert.False(await inbox.MarkProcessedAsync(id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task Outbox_PendingPreservesOrderAndHonorsLimit()
    {
        var outbox = new InMemoryOutboxStore();
        var ids = new[] { Guid.NewGuid(), Guid.NewGuid(), Guid.NewGuid() };
        foreach (var id in ids)
        {
            await outbox.AddAsync(NewRecord(id), TestContext.Current.CancellationToken);
        }

        var firstTwo = await outbox.GetPendingAsync(2, TestContext.Current.CancellationToken);

        Assert.Equal(ids.Take(2), firstTwo.Select(record => record.Id));
        Assert.Equal(3, outbox.PendingCount);
    }

    [Fact]
    public async Task Outbox_MarkSentExcludesFromPending()
    {
        var outbox = new InMemoryOutboxStore();
        var id = Guid.NewGuid();
        await outbox.AddAsync(NewRecord(id), TestContext.Current.CancellationToken);

        await outbox.MarkSentAsync(id, TestEvents.Start, TestContext.Current.CancellationToken);

        Assert.Equal(0, outbox.PendingCount);
        Assert.Empty(await outbox.GetPendingAsync(10, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task Outbox_MarkSentUnknownIdThrows()
    {
        var outbox = new InMemoryOutboxStore();

        await Assert.ThrowsAsync<InvalidOperationException>(() =>
            outbox.MarkSentAsync(Guid.NewGuid(), TestEvents.Start, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task DeadLetterStore_RetainsEntriesAndCount()
    {
        var store = new InMemoryDeadLetterStore();
        var deadLetter = new DeadLetter(
            TestEvents.Envelope(TestEvents.Placed()),
            Attempts: 3,
            DeadLetterFailureClass.Exhausted,
            Reason: "boom",
            DeadLetteredAt: TestEvents.Start);

        await store.AddAsync(deadLetter, TestContext.Current.CancellationToken);

        Assert.Equal(1, store.Count);
        Assert.Equal(deadLetter, Assert.Single(store.Snapshot()));
    }

    private static OutboxRecord NewRecord(Guid id) => new()
    {
        Id = id,
        Type = "orders.order-placed.v1",
        Payload = "{}",
        Subject = "order-1",
        OccurredAt = TestEvents.Start,
    };
}
