using System.ComponentModel.DataAnnotations;

namespace Orders.Grpc.Api;

/// <summary>
/// Server shape and hardening knobs, all from config
/// (csharp/foundations/configuration.md). Two listeners, mirroring the Go
/// reference: gRPC on an HTTP/2-only port, probes on a plain HTTP/1.1 port -
/// h2c and HTTP/1.1 cannot share a plaintext listener.
/// </summary>
internal sealed class ServerOptions
{
    public const string SectionName = "Server";

    [Range(1, 65535)]
    public int GrpcPort { get; init; } = 5001;

    [Range(1, 65535)]
    public int ProbesPort { get; init; } = 8080;

    /// <summary>
    /// Server-side ceiling applied only when the CLIENT sends no deadline: a
    /// missing deadline is not permission to run forever
    /// (csharp/services/grpc-services.md). A client deadline that is tighter
    /// is enforced by the framework and left untouched.
    /// </summary>
    [Range(typeof(TimeSpan), "00:00:01", "00:10:00")]
    public TimeSpan MaxRpcDuration { get; init; } = TimeSpan.FromSeconds(30);

    /// <summary>
    /// Server reflection is for internal services; config-gate it OFF at
    /// public edges (csharp/services/grpc-services.md).
    /// </summary>
    public bool EnableReflection { get; init; } = true;
}
