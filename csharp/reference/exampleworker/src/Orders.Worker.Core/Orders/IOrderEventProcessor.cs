using Orders.Worker.Core.Events;

namespace Orders.Worker.Core.Orders;

/// <summary>
/// The domain-behavior port the consumer invokes for each decoded message.
/// Intentionally narrow; the in-memory <see cref="OrderProjector"/> and any
/// database-backed implementation satisfy it.
///
/// Processing SHOULD be idempotent at the domain level where feasible, but the
/// consumer ALSO guards it with the inbox keyed by envelope id, so a duplicate
/// delivery never invokes this twice. A thrown
/// <see cref="InvalidEventException"/> is terminal (dead-letter); an
/// <see cref="OperationCanceledException"/> from the caller's token is orderly
/// shutdown (nack, redeliver); anything else is transient and retried with
/// bounded backoff.
/// </summary>
public interface IOrderEventProcessor
{
    Task ProcessAsync(OrderEvent orderEvent, CancellationToken cancellationToken);
}
