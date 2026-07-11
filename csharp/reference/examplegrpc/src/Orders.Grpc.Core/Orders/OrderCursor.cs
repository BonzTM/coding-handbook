using System.Globalization;
using System.Text;

namespace Orders.Grpc.Core.Orders;

/// <summary>
/// Opaque keyset-pagination cursor over the stable (CreatedAt, Id) sort key.
/// Encoded as URL-safe base64 so clients pass it back verbatim; decoding is
/// strict - any malformed token is rejected, never guessed at. On the wire it
/// is the page_token / next_page_token field (same contract as the keystone
/// HTTP module's cursor query parameter).
/// </summary>
public readonly record struct OrderCursor(DateTimeOffset CreatedAt, Guid Id)
{
    private const char Separator = '|';

    public string Encode()
    {
        string plain = string.Create(
            CultureInfo.InvariantCulture,
            $"{CreatedAt:O}{Separator}{Id:N}");
        return Convert.ToBase64String(Encoding.UTF8.GetBytes(plain))
            .TrimEnd('=')
            .Replace('+', '-')
            .Replace('/', '_');
    }

    public static bool TryDecode(string? token, out OrderCursor cursor)
    {
        cursor = default;
        if (string.IsNullOrEmpty(token) || token.Length > 128)
        {
            return false;
        }

        if (!TryFromBase64Url(token, out string plain))
        {
            return false;
        }

        string[] parts = plain.Split(Separator);
        if (parts.Length != 2)
        {
            return false;
        }

        if (!DateTimeOffset.TryParseExact(
                parts[0], "O", CultureInfo.InvariantCulture, DateTimeStyles.RoundtripKind, out var createdAt))
        {
            return false;
        }

        if (!Guid.TryParseExact(parts[1], "N", out var id))
        {
            return false;
        }

        cursor = new OrderCursor(createdAt, id);
        return true;
    }

    private static bool TryFromBase64Url(string token, out string plain)
    {
        plain = string.Empty;
        string padded = token.Replace('-', '+').Replace('_', '/');
        padded = (padded.Length % 4) switch
        {
            2 => padded + "==",
            3 => padded + "=",
            0 => padded,
            _ => string.Empty,
        };
        if (padded.Length == 0)
        {
            return false;
        }

        byte[] buffer = new byte[padded.Length];
        if (!Convert.TryFromBase64String(padded, buffer, out int written))
        {
            return false;
        }

        plain = Encoding.UTF8.GetString(buffer.AsSpan(0, written));
        return true;
    }
}
