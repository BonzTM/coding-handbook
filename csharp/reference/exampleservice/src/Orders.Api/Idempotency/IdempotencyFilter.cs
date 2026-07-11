using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

using Microsoft.AspNetCore.Http.HttpResults;

using Orders.Api.Auth;
using Orders.Api.Contracts;
using Orders.Api.ErrorHandling;
using Orders.Api.Telemetry;
using Orders.Core.Idempotency;

namespace Orders.Api.Idempotency;

/// <summary>
/// Endpoint filter enforcing the Idempotency-Key contract on POST /orders
/// (csharp/recipes/add-idempotent-write.md):
///
/// - the header is REQUIRED (this is resource creation): missing/malformed → 400
/// - key scope is (tenant, route, key); fingerprint is a SHA-256 over the
///   canonical request so the same key with a different body is a 422
/// - first use runs the handler inside the runner's single transaction and
///   records the 201 response; error outcomes are not recorded, so a retry
///   after a 4xx/5xx re-executes
/// - a completed duplicate replays the stored bytes; an in-flight duplicate is
///   a 409
/// </summary>
internal sealed class IdempotencyFilter(IIdempotencyRunner runner, OrdersMetrics metrics) : IEndpointFilter
{
    public const string HeaderName = "Idempotency-Key";

    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(next);

        var http = context.HttpContext;
        string key = http.Request.Headers[HeaderName].ToString();
        if (!IsWellFormedKey(key))
        {
            return TypedResults.Problem(
                statusCode: StatusCodes.Status400BadRequest,
                type: ProblemTypes.IdempotencyKeyMissing,
                title: "A well-formed Idempotency-Key header is required.",
                detail: $"Send a unique client-generated key (1-{IdempotencyScope.MaxKeyLength} characters: letters, digits, '-', '_', ':', '.').");
        }

        // Filter applied to an endpoint it does not understand - a wiring
        // bug, not a client error.
        var request = context.Arguments.OfType<CreateOrderRequest>().FirstOrDefault() ?? throw new InvalidOperationException(
            "IdempotencyFilter requires a bound CreateOrderRequest argument.");

        string route = $"{http.Request.Method} {http.Request.Path}";
        var scope = new IdempotencyScope(http.User.RequiredTenant(), route, key);
        string fingerprint = Fingerprint(route, request);

        object? passthrough = null;
        var outcome = await runner.RunAsync(
                scope,
                fingerprint,
                async cancellationToken =>
                {
                    object? result = await next(context).ConfigureAwait(false);
                    var stored = TryCapture(result);
                    if (stored is null)
                    {
                        passthrough = result; // error outcome: return it, do not record it
                    }

                    return stored;
                },
                http.RequestAborted)
            .ConfigureAwait(false);

        return ToResult(outcome, passthrough);
    }

    private object? ToResult(IdempotencyResult outcome, object? passthrough)
    {
        switch (outcome)
        {
            case IdempotencyResult.Executed { Response: { } stored }:
                return new StoredResponseResult(stored, replayed: false);
            case IdempotencyResult.Executed:
                return passthrough;
            case IdempotencyResult.Replayed replayed:
                metrics.IdempotentReplay();
                return new StoredResponseResult(replayed.Response, replayed: true);
            case IdempotencyResult.InFlight:
                return TypedResults.Problem(
                    statusCode: StatusCodes.Status409Conflict,
                    type: ProblemTypes.IdempotencyInFlight,
                    title: "A request with this Idempotency-Key is still in flight.",
                    detail: "Retry after the original request completes.");
            case IdempotencyResult.FingerprintMismatch:
                return TypedResults.Problem(
                    statusCode: StatusCodes.Status422UnprocessableEntity,
                    type: ProblemTypes.IdempotencyKeyReuse,
                    title: "This Idempotency-Key was already used for a different request.",
                    detail: "Generate a fresh key for each distinct request.");
            default:
                throw new InvalidOperationException($"Unhandled idempotency outcome {outcome.GetType().Name}.");
        }
    }

    /// <summary>Only the success shape is recorded; anything else is passed through unrecorded.</summary>
    private static StoredResponse? TryCapture(object? result)
    {
        if (result is not Created<OrderResponse> { Value: { } body } created)
        {
            return null;
        }

        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(body, OrdersJsonContext.Default.OrderResponse);
        return new StoredResponse(
            created.StatusCode,
            "application/json; charset=utf-8",
            bytes,
            created.Location);
    }

    /// <summary>
    /// SHA-256 over the canonical request (route + the bound DTO re-serialized
    /// through the wire context), lowercase hex. Canonicalizing from the bound
    /// DTO means insignificant whitespace differences do not defeat replay.
    /// </summary>
    internal static string Fingerprint(string route, CreateOrderRequest request)
    {
        byte[] payload = JsonSerializer.SerializeToUtf8Bytes(request, OrdersJsonContext.Default.CreateOrderRequest);
        byte[] buffer = new byte[Encoding.UTF8.GetByteCount(route) + 1 + payload.Length];
        int written = Encoding.UTF8.GetBytes(route, buffer);
        buffer[written] = (byte)'\n';
        payload.CopyTo(buffer.AsSpan(written + 1));
        return Convert.ToHexStringLower(SHA256.HashData(buffer));
    }

    internal static bool IsWellFormedKey(string key)
    {
        if (key.Length is 0 or > IdempotencyScope.MaxKeyLength)
        {
            return false;
        }

        foreach (char c in key)
        {
            bool ok = char.IsAsciiLetterOrDigit(c) || c is '-' or '_' or ':' or '.';
            if (!ok)
            {
                return false;
            }
        }

        return true;
    }
}
