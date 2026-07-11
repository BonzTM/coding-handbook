namespace Orders.Core.Idempotency;

/// <summary>
/// The response captured for an idempotent write: exact status, content type,
/// body bytes, and Location header. A replay returns these bytes verbatim -
/// re-deriving the response from live state is forbidden
/// (csharp/recipes/add-idempotent-write.md, Byte-identical replay).
/// </summary>
public sealed record StoredResponse(
    int StatusCode,
    string ContentType,
    ReadOnlyMemory<byte> Body,
    string? Location);
