using System.Security.Cryptography.X509Certificates;

namespace Orders.Grpc.Api.Tls;

/// <summary>
/// Loads the PEM server key pair for Kestrel. Throws on any unreadable or
/// mismatched pair - the caller lets that abort startup, because a configured
/// certificate that cannot load must NEVER downgrade to plaintext
/// (csharp/services/grpc-services.md, Transport Security).
/// </summary>
internal static class ServerCertificateLoader
{
    public static X509Certificate2 Load(string certPath, string keyPath)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(certPath);
        ArgumentException.ThrowIfNullOrWhiteSpace(keyPath);
        using var pem = X509Certificate2.CreateFromPemFile(certPath, keyPath);
        // Round-trip through PKCS#12: a PEM-loaded private key is ephemeral on
        // Windows and SChannel cannot use it for TLS; the re-import produces a
        // key both platforms serve with (csharp/foundations/cross-platform.md).
        return X509CertificateLoader.LoadPkcs12(
            pem.Export(X509ContentType.Pkcs12), password: null);
    }
}
