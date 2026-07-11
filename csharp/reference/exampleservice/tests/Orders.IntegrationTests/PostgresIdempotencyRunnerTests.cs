using System.Text;

using Microsoft.Extensions.Options;
using Microsoft.Extensions.Time.Testing;

using Orders.Core.Idempotency;
using Orders.Core.Orders;
using Orders.Infrastructure.Data;

using Xunit;

namespace Orders.IntegrationTests;

/// <summary>
/// The PostgreSQL idempotency runner against the real database
/// (csharp/recipes/add-idempotent-write.md): claim via unique constraint,
/// byte-identical replay, fingerprint mismatch, in-flight duplicates, released
/// claims on error outcomes, and TTL takeover under a FakeTimeProvider.
/// </summary>
public sealed class PostgresIdempotencyRunnerTests
{
    private static readonly TimeSpan _ttl = TimeSpan.FromHours(24);

    private readonly PostgresFixture _postgres;
    private readonly FakeTimeProvider _time = new(new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero));

    public PostgresIdempotencyRunnerTests(PostgresFixture postgres)
    {
        _postgres = postgres;
    }

    private PostgresIdempotencyRunner CreateRunner(OrdersDbContext db)
        => new(db, _time, Options.Create(new IdempotencyOptions { Ttl = _ttl }));

    private static IdempotencyScope FreshScope()
        => new(new TenantId($"i-{Guid.NewGuid():N}"[..20]), "POST /orders", $"key-{Guid.NewGuid():N}");

    private static StoredResponse Response(string body)
        => new(201, "application/json; charset=utf-8", Encoding.UTF8.GetBytes(body), "/orders/1");

    [Fact]
    public async Task FirstUse_ExecutesAndRecordsResponse()
    {
        IdempotencyScope scope = FreshScope();
        await using OrdersDbContext db = _postgres.CreateContext();
        int executions = 0;

        IdempotencyResult result = await CreateRunner(db).RunAsync(
            scope, "fp-1",
            _ =>
            {
                executions++;
                return Task.FromResult<StoredResponse?>(Response("""{"orderId":"1"}"""));
            },
            TestContext.Current.CancellationToken);

        Assert.Equal(1, executions);
        var executed = Assert.IsType<IdempotencyResult.Executed>(result);
        Assert.NotNull(executed.Response);
    }

    [Fact]
    public async Task Duplicate_SameFingerprint_ReplaysWithoutReExecuting()
    {
        IdempotencyScope scope = FreshScope();
        int executions = 0;

        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await CreateRunner(db).RunAsync(
                scope, "fp-1",
                _ =>
                {
                    executions++;
                    return Task.FromResult<StoredResponse?>(Response("""{"orderId":"first"}"""));
                },
                TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext retryDb = _postgres.CreateContext();
        IdempotencyResult result = await CreateRunner(retryDb).RunAsync(
            scope, "fp-1",
            _ =>
            {
                executions++;
                return Task.FromResult<StoredResponse?>(Response("""{"orderId":"second"}"""));
            },
            TestContext.Current.CancellationToken);

        Assert.Equal(1, executions); // the operation never ran again
        var replayed = Assert.IsType<IdempotencyResult.Replayed>(result);
        Assert.Equal(201, replayed.Response.StatusCode);
        Assert.Equal(
            """{"orderId":"first"}""",
            Encoding.UTF8.GetString(replayed.Response.Body.Span)); // byte-identical
        Assert.Equal("/orders/1", replayed.Response.Location);
    }

    [Fact]
    public async Task Duplicate_DifferentFingerprint_ReportsMismatch()
    {
        IdempotencyScope scope = FreshScope();
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await CreateRunner(db).RunAsync(
                scope, "fp-1",
                _ => Task.FromResult<StoredResponse?>(Response("{}")),
                TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext retryDb = _postgres.CreateContext();
        IdempotencyResult result = await CreateRunner(retryDb).RunAsync(
            scope, "fp-DIFFERENT",
            _ => Task.FromResult<StoredResponse?>(Response("{}")),
            TestContext.Current.CancellationToken);

        Assert.IsType<IdempotencyResult.FingerprintMismatch>(result);
    }

    [Fact]
    public async Task Duplicate_WhileFirstStillInFlight_Reports409Case()
    {
        IdempotencyScope scope = FreshScope();

        // Simulate a crashed/still-running first request: a committed InFlight claim.
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            db.IdempotencyRecords.Add(new IdempotencyRecord
            {
                Id = Guid.CreateVersion7(),
                TenantId = scope.TenantId.Value,
                Route = scope.Route,
                Key = scope.Key,
                Fingerprint = "fp-1",
                State = IdempotencyState.InFlight,
                CreatedAt = _time.GetUtcNow(),
            });
            await db.SaveChangesAsync(TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext duplicateDb = _postgres.CreateContext();
        IdempotencyResult result = await CreateRunner(duplicateDb).RunAsync(
            scope, "fp-1",
            _ => Task.FromResult<StoredResponse?>(Response("{}")),
            TestContext.Current.CancellationToken);

        Assert.IsType<IdempotencyResult.InFlight>(result);
    }

    [Fact]
    public async Task ErrorOutcome_ReleasesClaimSoRetryReExecutes()
    {
        IdempotencyScope scope = FreshScope();

        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            IdempotencyResult errorOutcome = await CreateRunner(db).RunAsync(
                scope, "fp-1",
                _ => Task.FromResult<StoredResponse?>(null), // domain error: do not record
                TestContext.Current.CancellationToken);
            var executed = Assert.IsType<IdempotencyResult.Executed>(errorOutcome);
            Assert.Null(executed.Response);
        }

        await using OrdersDbContext retryDb = _postgres.CreateContext();
        IdempotencyResult retry = await CreateRunner(retryDb).RunAsync(
            scope, "fp-1",
            _ => Task.FromResult<StoredResponse?>(Response("{}")),
            TestContext.Current.CancellationToken);

        var retried = Assert.IsType<IdempotencyResult.Executed>(retry);
        Assert.NotNull(retried.Response); // re-executed, not replayed and not 409
    }

    [Fact]
    public async Task ExpiredRecord_IsTreatedAsNeverSeen()
    {
        IdempotencyScope scope = FreshScope();
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await CreateRunner(db).RunAsync(
                scope, "fp-1",
                _ => Task.FromResult<StoredResponse?>(Response("""{"orderId":"old"}""")),
                TestContext.Current.CancellationToken);
        }

        _time.Advance(_ttl + TimeSpan.FromMinutes(1)); // past the dedupe window

        int executions = 0;
        await using OrdersDbContext lateDb = _postgres.CreateContext();
        IdempotencyResult result = await CreateRunner(lateDb).RunAsync(
            scope, "fp-2", // even a different request is fine now
            _ =>
            {
                executions++;
                return Task.FromResult<StoredResponse?>(Response("""{"orderId":"new"}"""));
            },
            TestContext.Current.CancellationToken);

        Assert.Equal(1, executions);
        var executed = Assert.IsType<IdempotencyResult.Executed>(result);
        Assert.Equal(
            """{"orderId":"new"}""",
            Encoding.UTF8.GetString(executed.Response!.Body.Span));
    }
}
