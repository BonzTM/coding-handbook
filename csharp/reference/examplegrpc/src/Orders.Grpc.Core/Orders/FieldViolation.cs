namespace Orders.Grpc.Core.Orders;

/// <summary>
/// One request field that failed validation, with an actionable reason. This
/// is the structured, transport-agnostic carrier the gRPC boundary renders as
/// a google.rpc.BadRequest field violation - Core stays free of any wire or
/// proto dependency (csharp/services/grpc-services.md, Error Details).
/// </summary>
public readonly record struct FieldViolation(string Field, string Description);
