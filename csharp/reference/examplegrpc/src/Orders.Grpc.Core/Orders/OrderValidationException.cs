namespace Orders.Grpc.Core.Orders;

/// <summary>
/// Order input failed validation. Carries every offending field structurally
/// so the transport can attach a google.rpc.BadRequest detail listing each
/// {field, description} pair instead of a flat message string
/// (csharp/services/grpc-services.md, Error Details). Transport maps it to
/// INVALID_ARGUMENT.
/// </summary>
public sealed class OrderValidationException : Exception
{
    public OrderValidationException(IReadOnlyList<FieldViolation> violations)
        : base(BuildMessage(violations))
    {
        Violations = violations;
    }

    public IReadOnlyList<FieldViolation> Violations { get; }

    private static string BuildMessage(IReadOnlyList<FieldViolation> violations)
    {
        ArgumentNullException.ThrowIfNull(violations);
        ArgumentOutOfRangeException.ThrowIfZero(violations.Count);
        return "Order validation failed: "
            + string.Join("; ", violations.Select(v => $"{v.Field}: {v.Description}"));
    }
}
