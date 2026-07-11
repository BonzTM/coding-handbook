namespace Orders.Core.Orders;

/// <summary>
/// The tenant already has an order with this external reference. Raised by the
/// repository when the database reports a unique violation (SQLSTATE 23505) -
/// never by a racy pre-check SELECT (csharp/services/database.md). Transport
/// maps it to 409.
/// </summary>
public sealed class DuplicateOrderException(string externalReference)
    : Exception($"An order with external reference '{externalReference}' already exists.")
{
    public string ExternalReference { get; } = externalReference;
}
