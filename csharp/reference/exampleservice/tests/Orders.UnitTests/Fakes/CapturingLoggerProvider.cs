using Microsoft.Extensions.Logging;

namespace Orders.UnitTests.Fakes;

/// <summary>
/// Captures log records per category so tests can assert on the DEDICATED
/// audit stream (category "Orders.Audit") - who/what/result/requestId - the
/// same way the Go reference's audit tests do (csharp/operations/security.md,
/// Audit Logging).
/// </summary>
public sealed class CapturingLoggerProvider : ILoggerProvider
{
    private readonly Lock _gate = new();
    private readonly List<CapturedLogRecord> _records = [];

    public IReadOnlyList<CapturedLogRecord> Snapshot()
    {
        lock (_gate)
        {
            return [.. _records];
        }
    }

    public IReadOnlyList<CapturedLogRecord> AuditRecords()
        => [.. Snapshot().Where(r => string.Equals(r.Category, "Orders.Audit", StringComparison.Ordinal))];

    public ILogger CreateLogger(string categoryName) => new CapturingLogger(this, categoryName);

    public void Dispose() => GC.SuppressFinalize(this);

    private void Add(CapturedLogRecord record)
    {
        lock (_gate)
        {
            _records.Add(record);
        }
    }

    private sealed class CapturingLogger(CapturingLoggerProvider owner, string category) : ILogger
    {
        public IDisposable? BeginScope<TState>(TState state)
            where TState : notnull => null;

        public bool IsEnabled(LogLevel logLevel) => true;

        public void Log<TState>(
            LogLevel logLevel,
            EventId eventId,
            TState state,
            Exception? exception,
            Func<TState, Exception?, string> formatter)
        {
            ArgumentNullException.ThrowIfNull(formatter);
            owner.Add(new CapturedLogRecord(category, logLevel, formatter(state, exception)));
        }
    }
}

public sealed record CapturedLogRecord(string Category, LogLevel Level, string Message);
