using System.Security.Cryptography.X509Certificates;

namespace Orders.Grpc.Api.Tls;

/// <summary>
/// mTLS client-certificate verification against the configured CA bundle and
/// ONLY that bundle: CustomRootTrust ignores the machine's root store, so a
/// certificate from any public CA is rejected unless the operator put that CA
/// in the bundle. Revocation is not checked - the bundle model is a private
/// internal CA without CRL/OCSP endpoints; rotate the bundle to revoke
/// (csharp/services/grpc-services.md, Transport Security).
/// </summary>
internal sealed class ClientCertificateValidator
{
    private readonly X509Certificate2Collection _trustedRoots;

    private ClientCertificateValidator(X509Certificate2Collection trustedRoots)
    {
        _trustedRoots = trustedRoots;
    }

    /// <summary>Fail-fast factory: an unreadable or empty CA bundle aborts startup.</summary>
    public static ClientCertificateValidator FromCaBundle(string caBundlePath)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(caBundlePath);
        var roots = new X509Certificate2Collection();
        roots.ImportFromPemFile(caBundlePath);
        if (roots.Count == 0)
        {
            throw new InvalidOperationException(
                $"The client CA bundle '{caBundlePath}' contains no certificates.");
        }

        return new ClientCertificateValidator(roots);
    }

    public bool Validates(X509Certificate2 candidate)
    {
        ArgumentNullException.ThrowIfNull(candidate);
        using var chain = new X509Chain();
        chain.ChainPolicy.TrustMode = X509ChainTrustMode.CustomRootTrust;
        chain.ChainPolicy.CustomTrustStore.AddRange(_trustedRoots);
        chain.ChainPolicy.RevocationMode = X509RevocationMode.NoCheck;
        return chain.Build(candidate);
    }
}
