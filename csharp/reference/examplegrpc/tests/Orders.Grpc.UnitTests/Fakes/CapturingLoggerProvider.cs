using Microsoft.Extensions.Logging;

namespace Orders.Grpc.UnitTests.Fakes;

/// <summary>
/// Captures log records per category so tests can assert the access-log and
/// exception-shielding contracts (one line per RPC, unexpected exceptions
/// logged once server-side) without scraping console output.
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
