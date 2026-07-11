using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

using Orders.Core.Orders;

namespace Orders.Api.ErrorHandling;

/// <summary>
/// The ONE place domain exceptions map to wire status/type pairs - never
/// scattered try/catch per endpoint (csharp/foundations/errors-and-logging.md).
/// Anything unmapped falls through to the default exception handler: an opaque
/// 500 ProblemDetails with requestId, detail only in the server log.
/// </summary>
internal sealed class DomainExceptionHandler(IProblemDetailsService problems) : IExceptionHandler
{
    public async ValueTask<bool> TryHandleAsync(
        HttpContext httpContext, Exception exception, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(httpContext);

        var mapped = Map(exception);
        if (mapped is null)
        {
            return false;
        }

        httpContext.Response.StatusCode = mapped.Status ?? StatusCodes.Status500InternalServerError;
        return await problems.TryWriteAsync(new ProblemDetailsContext
        {
            HttpContext = httpContext,
            Exception = exception,
            ProblemDetails = mapped,
        }).ConfigureAwait(false);
    }

    private static ProblemDetails? Map(Exception exception) => exception switch
    {
        OrderNotFoundException ex => new ProblemDetails
        {
            Status = StatusCodes.Status404NotFound,
            Type = ProblemTypes.OrderNotFound,
            Title = "Order not found.",
            Detail = $"Order {ex.OrderId} does not exist.",
        },
        DuplicateOrderException ex => new ProblemDetails
        {
            Status = StatusCodes.Status409Conflict,
            Type = ProblemTypes.DuplicateOrder,
            Title = "Duplicate order.",
            Detail = $"An order with external reference '{ex.ExternalReference}' already exists.",
        },
        OrderVersionConflictException ex => new ProblemDetails
        {
            Status = StatusCodes.Status409Conflict,
            Type = ProblemTypes.VersionConflict,
            Title = "Concurrent modification.",
            Detail = $"Order {ex.OrderId} was modified concurrently; reload and retry with the current version.",
        },
        _ => null,
    };
}
