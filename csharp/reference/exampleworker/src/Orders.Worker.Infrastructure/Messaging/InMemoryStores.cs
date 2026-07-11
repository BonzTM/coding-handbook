using Orders.Worker.Core.Messaging;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// In-memory <see cref="IInboxStore"/> for offline tests and local dev.
/// Check-and-record is atomic under the lock, mirroring the SQL inbox's
/// insert-on-conflict-do-nothing keyed by event id, so two concurrent
/// deliveries of the same id cannot both be treated as first-seen. It also
/// implements <see cref="IInboxCompensation"/> because it is NOT transactional
/// with the domain side effect - the production SQL inbox writes both in one
/// transaction and needs no compensation hook.
/// </summary>
internal sealed class InMemoryInboxStore : IInboxStore, IInboxCompensation
{
    private readonly Lock _gate = new();
    private readonly HashSet<Guid> _seen = [];

    public Task<bool> MarkProcessedAsync(Guid eventId, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            return Task.FromResult(!_seen.Add(eventId));
        }
    }

    public Task<bool> SeenAsync(Guid eventId, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            return Task.FromResult(_seen.Contains(eventId));
        }
    }

    public Task RemoveAsync(Guid eventId, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            _seen.Remove(eventId);
        }

        return Task.CompletedTask;
    }
}

/// <summary>
/// In-memory <see cref="IOutboxStore"/> preserving occurred-at (enqueue)
/// order. The production implementation is an EF Core entity + migration; see
/// the port's documentation for the exact SQL seam.
/// </summary>
internal sealed class InMemoryOutboxStore : IOutboxStore
{
    private readonly Lock _gate = new();
    private readonly List<Guid> _order = [];
    private readonly Dictionary<Guid, OutboxRecord> _records = [];

    /// <summary>Unsent records still waiting for the relay - the outbox
    /// backlog gauge and test assertions read this.</summary>
    public int PendingCount
    {
        get
        {
            lock (_gate)
            {
                return _order.Count(id => _records[id].SentAt is null);
            }
        }
    }

    public Task AddAsync(OutboxRecord record, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(record);
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            if (!_records.ContainsKey(record.Id))
            {
                _order.Add(record.Id);
            }

            _records[record.Id] = record;
        }

        return Task.CompletedTask;
    }

    public Task<IReadOnlyList<OutboxRecord>> GetPendingAsync(int limit, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            var pending = new List<OutboxRecord>();
            foreach (var id in _order)
            {
                var record = _records[id];
                if (record.SentAt is not null)
                {
                    continue;
                }

                pending.Add(record);
                if (pending.Count == limit)
                {
                    break;
                }
            }

            return Task.FromResult<IReadOnlyList<OutboxRecord>>(pending);
        }
    }

    public Task MarkSentAsync(Guid id, DateTimeOffset sentAt, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            if (!_records.TryGetValue(id, out var record))
            {
                throw new InvalidOperationException($"Outbox record '{id}' does not exist.");
            }

            record.SentAt = sentAt;
        }

        return Task.CompletedTask;
    }
}

/// <summary>
/// In-memory <see cref="IDeadLetterStore"/>. Production parks to a DLQ topic
/// or table behind the same port; the runbook owns the replay procedure.
/// </summary>
internal sealed class InMemoryDeadLetterStore : IDeadLetterStore
{
    private readonly Lock _gate = new();
    private readonly List<DeadLetter> _entries = [];

    /// <summary>Parked-message count - the DLQ-depth gauge reads this.</summary>
    public int Count
    {
        get
        {
            lock (_gate)
            {
                return _entries.Count;
            }
        }
    }

    public Task AddAsync(DeadLetter deadLetter, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(deadLetter);
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            _entries.Add(deadLetter);
        }

        return Task.CompletedTask;
    }

    /// <summary>A copy of the parked messages for assertions and operator tooling.</summary>
    public IReadOnlyList<DeadLetter> Snapshot()
    {
        lock (_gate)
        {
            return [.. _entries];
        }
    }
}
