using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;

using Orders.Grpc.Api.Tls;

using Xunit;

namespace Orders.Grpc.UnitTests.Tls;

/// <summary>
/// The transport-security invariants that do not need a live TLS handshake
/// (csharp/services/grpc-services.md, Verification And Proof): a configured
/// key pair that cannot load FAILS (never a silent plaintext downgrade), a
/// good PEM pair loads with its private key, and the mTLS validator accepts
/// only certificates chaining to the configured CA bundle.
/// </summary>
public sealed class TlsGatingTests : IDisposable
{
    private readonly string _dir = Directory.CreateTempSubdirectory("orders-grpc-tls-").FullName;

    public void Dispose() => Directory.Delete(_dir, recursive: true);

    [Fact]
    public void ServerCertificateLoader_MissingFiles_Throws()
    {
        string cert = Path.Combine(_dir, "missing.crt");
        string key = Path.Combine(_dir, "missing.key");

        Assert.ThrowsAny<Exception>(() => ServerCertificateLoader.Load(cert, key));
    }

    [Fact]
    public void ServerCertificateLoader_MismatchedKeyPair_Throws()
    {
        using var rightKey = RSA.Create(2048);
        using var wrongKey = RSA.Create(2048);
        using var certificate = CreateSelfSigned("CN=server", rightKey);
        string certPath = WriteFile("server.crt", certificate.ExportCertificatePem());
        string keyPath = WriteFile("server.key", wrongKey.ExportPkcs8PrivateKeyPem());

        Assert.ThrowsAny<Exception>(() => ServerCertificateLoader.Load(certPath, keyPath));
    }

    [Fact]
    public void ServerCertificateLoader_ValidPemPair_LoadsWithPrivateKey()
    {
        using var key = RSA.Create(2048);
        using var certificate = CreateSelfSigned("CN=server", key);
        string certPath = WriteFile("server.crt", certificate.ExportCertificatePem());
        string keyPath = WriteFile("server.key", key.ExportPkcs8PrivateKeyPem());

        using var loaded = ServerCertificateLoader.Load(certPath, keyPath);

        Assert.True(loaded.HasPrivateKey);
        Assert.Equal(certificate.Thumbprint, loaded.Thumbprint);
    }

    [Fact]
    public void ClientCertificateValidator_EmptyBundle_FailsFast()
    {
        string bundle = WriteFile("empty-ca.pem", "\n");

        Assert.ThrowsAny<Exception>(() => ClientCertificateValidator.FromCaBundle(bundle));
    }

    [Fact]
    public void ClientCertificateValidator_AcceptsCertSignedByTheBundleCa_AndNothingElse()
    {
        using var caKey = RSA.Create(2048);
        using var ca = CreateCa("CN=Test Internal CA", caKey);
        string bundle = WriteFile("ca.pem", ca.ExportCertificatePem());
        var validator = ClientCertificateValidator.FromCaBundle(bundle);

        using var clientKey = RSA.Create(2048);
        using var signedClient = IssueClientCertificate(ca, "CN=good-client", clientKey);
        Assert.True(validator.Validates(signedClient));

        using var strangerKey = RSA.Create(2048);
        using var selfSignedStranger = CreateSelfSigned("CN=stranger", strangerKey);
        Assert.False(validator.Validates(selfSignedStranger));
    }

    private string WriteFile(string name, string contents)
    {
        string path = Path.Combine(_dir, name);
        File.WriteAllText(path, contents);
        return path;
    }

    private static X509Certificate2 CreateSelfSigned(string subject, RSA key)
    {
        var request = new CertificateRequest(
            subject, key, HashAlgorithmName.SHA256, RSASignaturePadding.Pkcs1);
        return request.CreateSelfSigned(
            DateTimeOffset.UtcNow.AddDays(-1), DateTimeOffset.UtcNow.AddDays(30));
    }

    private static X509Certificate2 CreateCa(string subject, RSA key)
    {
        var request = new CertificateRequest(
            subject, key, HashAlgorithmName.SHA256, RSASignaturePadding.Pkcs1);
        request.CertificateExtensions.Add(
            new X509BasicConstraintsExtension(
                certificateAuthority: true, hasPathLengthConstraint: false, pathLengthConstraint: 0, critical: true));
        return request.CreateSelfSigned(
            DateTimeOffset.UtcNow.AddDays(-1), DateTimeOffset.UtcNow.AddDays(30));
    }

    private static X509Certificate2 IssueClientCertificate(X509Certificate2 ca, string subject, RSA key)
    {
        var request = new CertificateRequest(
            subject, key, HashAlgorithmName.SHA256, RSASignaturePadding.Pkcs1);
        byte[] serial = new byte[8];
        RandomNumberGenerator.Fill(serial);
        return request.Create(
            ca, DateTimeOffset.UtcNow.AddDays(-1), DateTimeOffset.UtcNow.AddDays(7), serial);
    }
}
