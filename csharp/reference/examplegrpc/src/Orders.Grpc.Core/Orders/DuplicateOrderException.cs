namespace Orders.Grpc.Core.Orders;

/// <summary>
/// The tenant already has an order with this external reference. Raised by the
/// store on a uniqueness conflict - never by a racy pre-check read. Transport
/// maps it to ALREADY_EXISTS.
/// </summary>
public sealed class DuplicateOrderException(string externalReference)
    : Exception($"An order with external reference '{externalReference}' already exists.")
{
    public string ExternalReference { get; } = externalReference;
}
