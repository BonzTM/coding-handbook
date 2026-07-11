using Orders.Grpc.Api.Auth;
using Orders.Grpc.Api.Tls;

using Xunit;

namespace Orders.Grpc.UnitTests.Tls;

/// <summary>
/// Startup invariants ValidateOnStart enforces
/// (csharp/foundations/configuration.md): half a TLS key pair, mTLS without
/// server TLS, or enabled auth without a real token are config ERRORS that
/// kill the process at deploy time - not warnings.
/// </summary>
public sealed class OptionsValidationTests
{
    private readonly TlsOptionsValidator _tls = new();
    private readonly AuthOptionsValidator _auth = new();

    [Fact]
    public void Tls_AllEmpty_IsValidPlaintextLocalDev()
    {
        Assert.True(_tls.Validate(null, new TlsOptions()).Succeeded);
    }

    [Fact]
    public void Tls_CertWithoutKey_Fails()
    {
        var result = _tls.Validate(null, new TlsOptions { CertPath = "/certs/tls.crt" });

        Assert.True(result.Failed);
    }

    [Fact]
    public void Tls_KeyWithoutCert_Fails()
    {
        var result = _tls.Validate(null, new TlsOptions { KeyPath = "/certs/tls.key" });

        Assert.True(result.Failed);
    }

    [Fact]
    public void Tls_ClientCaWithoutServerTls_Fails()
    {
        var result = _tls.Validate(null, new TlsOptions { ClientCaPath = "/certs/ca.pem" });

        Assert.True(result.Failed);
    }

    [Fact]
    public void Tls_FullMutualTlsTrio_Succeeds()
    {
        var result = _tls.Validate(null, new TlsOptions
        {
            CertPath = "/certs/tls.crt",
            KeyPath = "/certs/tls.key",
            ClientCaPath = "/certs/ca.pem",
        });

        Assert.True(result.Succeeded);
    }

    [Fact]
    public void Auth_EnabledWithoutToken_Fails()
    {
        var result = _auth.Validate(null, new AuthOptions { Enabled = true, BearerToken = "" });

        Assert.True(result.Failed);
    }

    [Fact]
    public void Auth_EnabledWithShortToken_Fails()
    {
        var result = _auth.Validate(null, new AuthOptions { Enabled = true, BearerToken = "short" });

        Assert.True(result.Failed);
    }

    [Fact]
    public void Auth_Disabled_NeedsNoToken()
    {
        Assert.True(_auth.Validate(null, new AuthOptions { Enabled = false }).Succeeded);
    }
}
