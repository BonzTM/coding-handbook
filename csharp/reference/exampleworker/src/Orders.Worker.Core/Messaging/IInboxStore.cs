namespace Orders.Worker.Core.Messaging;

/// <summary>
/// Durable dedupe port: how the consumer processes each event exactly once
/// under at-least-once delivery (csharp/services/eventing-and-messaging.md,
/// Outbox And Inbox Patterns). Defined at the consumer; the in-memory
/// implementation satisfies it here, and a SQL-backed inbox - an
/// insert-on-conflict-do-nothing table keyed by event id, written in the SAME
/// EF Core transaction as the handler's state change - plugs in unchanged.
/// </summary>
public interface IInboxStore
{
    /// <summary>
    /// Records that <paramref name="eventId"/> has been processed. Returns
    /// <see langword="true"/> if the id was ALREADY recorded (a duplicate
    /// delivery), in which case the consumer drops the message after ack
    /// without re-invoking the domain. Check-and-record must be atomic - the
    /// SQL inbox gets this from the unique constraint on the event id.
    /// </summary>
    Task<bool> MarkProcessedAsync(Guid eventId, CancellationToken cancellationToken);

    /// <summary>Reports whether an id has been recorded, without recording it.
    /// A read-only helper for tests and operator tooling.</summary>
    Task<bool> SeenAsync(Guid eventId, CancellationToken cancellationToken);
}

/// <summary>
/// Compensation seam for NON-transactional inbox stores: the consumer removes
/// the dedupe record when the domain side effect fails after the record was
/// written, so a retry is not mistaken for a duplicate. A SQL inbox does not
/// implement this - it writes the record and the side effect in one
/// transaction that rolls back together, which is the production contract this
/// hook simulates (mirrors the Go reference's in-memory rollback seam).
/// </summary>
public interface IInboxCompensation
{
    /// <summary>Removes a dedupe record after a failed side effect.</summary>
    Task RemoveAsync(Guid eventId, CancellationToken cancellationToken);
}
