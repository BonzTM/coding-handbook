using System.Security.Authentication;

using Microsoft.AspNetCore.Server.Kestrel.Core;
using Microsoft.AspNetCore.Server.Kestrel.Https;

namespace Orders.Grpc.Api.Tls;

/// <summary>
/// Kestrel listener wiring for the two-port shape (gRPC + probes) with
/// config-gated TLS/mTLS (csharp/services/grpc-services.md, Transport
/// Security). Called once from Program.cs. Any certificate problem throws
/// while the listener binds, which aborts startup - the fail-fast contract.
/// </summary>
internal static class KestrelTlsExtensions
{
    public static void ConfigureOrdersListeners(
        this KestrelServerOptions kestrel, ServerOptions server, TlsOptions tls)
    {
        ArgumentNullException.ThrowIfNull(kestrel);
        ArgumentNullException.ThrowIfNull(server);
        ArgumentNullException.ThrowIfNull(tls);

        // gRPC listener: HTTP/2 only. Plaintext h2c and HTTP/1.1 cannot share
        // a listener (no ALPN without TLS), hence the separate probes port.
        kestrel.ListenAnyIP(server.GrpcPort, listen =>
        {
            listen.Protocols = HttpProtocols.Http2;
            if (tls.HasServerTls)
            {
                listen.UseHttps(https => ConfigureHttps(https, tls));
            }
        });

        // Probes listener: plain HTTP/1.1 for /livez and /readyz - the
        // HTTP sidecar shape of the Go reference. Never TLS: the kubelet
        // probes it in-cluster.
        kestrel.ListenAnyIP(server.ProbesPort, listen => listen.Protocols = HttpProtocols.Http1);
    }

    private static void ConfigureHttps(HttpsConnectionAdapterOptions https, TlsOptions tls)
    {
        https.ServerCertificate = ServerCertificateLoader.Load(tls.CertPath, tls.KeyPath);
        // TLS 1.2 minimum, per csharp/services/grpc-services.md ("Enforce a
        // TLS minimum via HttpsConnectionAdapterOptions.SslProtocols (TLS 1.2
        // or later)"). CA5398 prefers None (OS default) so future TLS
        // versions are picked up automatically; the doc's explicit floor wins
        // here, reviewed and suppressed narrowly per csharp/quality/linting.md.
#pragma warning disable CA5398 // Avoid hardcoded SslProtocols values - the handbook mandates an explicit TLS 1.2 floor
        https.SslProtocols = SslProtocols.Tls12 | SslProtocols.Tls13;
#pragma warning restore CA5398
        if (tls.HasClientCa)
        {
            var validator = ClientCertificateValidator.FromCaBundle(tls.ClientCaPath);
            https.ClientCertificateMode = ClientCertificateMode.RequireCertificate;
            https.ClientCertificateValidation = (certificate, _, _) => validator.Validates(certificate);
        }
    }
}
