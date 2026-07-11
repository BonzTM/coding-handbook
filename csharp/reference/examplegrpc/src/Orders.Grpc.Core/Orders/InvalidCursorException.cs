namespace Orders.Grpc.Core.Orders;

/// <summary>
/// A pagination page token could not be decoded. Malformed tokens are a caller
/// error, never guessed at; transport maps this to INVALID_ARGUMENT - not
/// INTERNAL (mirrors the Go reference's ErrInvalidCursor).
/// </summary>
public sealed class InvalidCursorException()
    : Exception("The page token is malformed. Pass back next_page_token verbatim.");
