namespace Orders.Worker.Core.Messaging;

/// <summary>
/// One pending publish, written together with the domain state change - the
/// transactional-outbox row from csharp/services/eventing-and-messaging.md
/// (Outbox And Inbox Patterns): "save the domain change and an OutboxMessage
/// row in the same EF Core transaction, then relay asynchronously". The relay
/// reads pending records, publishes them, and marks them sent, so a process
/// that commits state but crashes before publishing still publishes on the
/// next scan - no dual-write loss.
/// </summary>
public sealed class OutboxRecord
{
    /// <summary>Stable record/event id; it becomes the published
    /// <see cref="Events.EventEnvelope.Id"/> so the consumer's inbox can dedupe
    /// a redelivered relay.</summary>
    public required Guid Id { get; init; }

    /// <summary>Event type ("orders.order-placed.v1").</summary>
    public required string Type { get; init; }

    /// <summary>Serialized JSON payload.</summary>
    public required string Payload { get; init; }

    /// <summary>Ordering key for the published envelope's Subject (the order
    /// id). Extends the doc's minimal OutboxMessage shape so the relay can
    /// honor the per-entity ordering contract.</summary>
    public string? Subject { get; init; }

    /// <summary>Producer timestamp (from <see cref="TimeProvider"/>), also the
    /// relay's drain order.</summary>
    public DateTimeOffset OccurredAt { get; init; }

    /// <summary>Set when the relay confirms publication. Null means pending.</summary>
    public DateTimeOffset? SentAt { get; set; }
}

/// <summary>
/// Pending-message store port the relay drains. The in-memory implementation
/// satisfies it here; the production implementation is an ordinary EF Core
/// entity + migration where <see cref="AddAsync"/> is the INSERT that runs in
/// the same transaction as the domain write, and <see cref="GetPendingAsync"/>
/// is <c>WHERE "SentAt" IS NULL ORDER BY "OccurredAt" ... FOR UPDATE SKIP
/// LOCKED</c> (csharp/services/database.md).
/// </summary>
public interface IOutboxStore
{
    /// <summary>Enqueues a pending record. In a database build this runs in the
    /// same transaction as the domain write - that atomicity is the entire
    /// point of the pattern.</summary>
    Task AddAsync(OutboxRecord record, CancellationToken cancellationToken);

    /// <summary>Returns up to <paramref name="limit"/> unsent records in
    /// occurred-at (enqueue) order.</summary>
    Task<IReadOnlyList<OutboxRecord>> GetPendingAsync(int limit, CancellationToken cancellationToken);

    /// <summary>Marks the record published so it is not relayed again. A crash
    /// between publish and mark-sent means one duplicate delivery - which the
    /// consumer's inbox absorbs, by design.</summary>
    Task MarkSentAsync(Guid id, DateTimeOffset sentAt, CancellationToken cancellationToken);
}
