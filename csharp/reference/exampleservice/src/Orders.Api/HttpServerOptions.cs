using System.ComponentModel.DataAnnotations;

namespace Orders.Api;

/// <summary>
/// Server hardening knobs (csharp/services/http-services.md): a default
/// request timeout and a body-size cap sized to the real payloads, not the
/// 30 MB Kestrel default. Both from config, per
/// csharp/foundations/configuration.md.
/// </summary>
internal sealed class HttpServerOptions
{
    public const string SectionName = "Server";

    [Range(typeof(TimeSpan), "00:00:01", "00:05:00")]
    public TimeSpan RequestTimeout { get; init; } = TimeSpan.FromSeconds(10);

    [Range(4_096, 32 * 1024 * 1024)]
    public long MaxRequestBodyBytes { get; init; } = 1 * 1024 * 1024;
}
