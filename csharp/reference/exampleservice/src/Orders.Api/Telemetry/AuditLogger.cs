namespace Orders.Api.Telemetry;

/// <summary>
/// The dedicated audit stream (csharp/operations/security.md, Audit Logging).
///
/// Records security-relevant actions - authentication failures, authorization
/// denials, and successful data-mutating writes - on the DEDICATED
/// `Orders.Audit` logger category, separate from the application logger, so
/// the logging pipeline can route it to its own sink with its own retention
/// and access controls. Every record carries who (actor + tenant), what
/// (action, resource, result), when (the log timestamp, UTC in production
/// formatters), and where (requestId). Never secrets, tokens, or payload PII -
/// resources are referenced by id only.
/// </summary>
internal sealed class AuditLogger(ILoggerFactory loggerFactory)
{
    public const string CategoryName = "Orders.Audit";

    private readonly ILogger _logger = loggerFactory.CreateLogger(CategoryName);

    public void AuthenticationFailed(string reason, string requestId)
        => AuditLog.AuthenticationFailed(_logger, "anonymous", "authenticate", reason, requestId);

    public void AuthorizationDenied(string actor, string tenant, string action, string requestId)
        => AuditLog.AuthorizationDenied(_logger, actor, tenant, action, requestId);

    public void OrderCreated(string actor, string tenant, string orderId, string requestId)
        => AuditLog.OrderMutated(_logger, actor, tenant, "order.create", orderId, "success", requestId);

    public void OrderAmended(string actor, string tenant, string orderId, string requestId)
        => AuditLog.OrderMutated(_logger, actor, tenant, "order.amend", orderId, "success", requestId);

    public void OrderDeleted(string actor, string tenant, string orderId, string requestId)
        => AuditLog.OrderMutated(_logger, actor, tenant, "order.delete", orderId, "success", requestId);
}

/// <summary>
/// Source-generated audit messages with stable structured fields
/// (csharp/foundations/errors-and-logging.md). The audit stream is never
/// sampled or rate-limited (csharp/operations/observability.md).
/// </summary>
internal static partial class AuditLog
{
    [LoggerMessage(
        Level = LogLevel.Warning,
        Message = "audit: {Actor} {Action} result=failure reason={Reason} requestId={RequestId}")]
    public static partial void AuthenticationFailed(
        ILogger logger, string actor, string action, string reason, string requestId);

    [LoggerMessage(
        Level = LogLevel.Warning,
        Message = "audit: {Actor}@{Tenant} {Action} result=denied requestId={RequestId}")]
    public static partial void AuthorizationDenied(
        ILogger logger, string actor, string tenant, string action, string requestId);

    [LoggerMessage(
        Level = LogLevel.Information,
        Message = "audit: {Actor}@{Tenant} {Action} resource={Resource} result={Result} requestId={RequestId}")]
    public static partial void OrderMutated(
        ILogger logger, string actor, string tenant, string action, string resource, string result, string requestId);
}
