using System.Net;
using System.Net.Http.Headers;
using System.Security.Cryptography;
using System.Text;

using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.TestHost;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.IdentityModel.JsonWebTokens;
using Microsoft.IdentityModel.Protocols;
using Microsoft.IdentityModel.Protocols.OpenIdConnect;
using Microsoft.IdentityModel.Tokens;

namespace Orders.UnitTests.Transport;

/// <summary>
/// Production-shaped auth host: `Auth:Enabled=true`, so the REAL JwtBearer
/// handler validates tokens against a JWKS document - served by an in-memory
/// stub of the issuer's discovery + JWKS endpoints, signed by a local RSA key
/// (csharp/operations/security.md). No network, no real identity provider.
/// </summary>
public sealed class JwksApiFactory : OrdersApiFactory
{
    public const string Authority = "https://issuer.test";
    public const string Audience = "orders-api";
    private const string KeyId = "unit-test-signing-key";

    private readonly RSA _rsa = RSA.Create(2048);

    protected override Dictionary<string, string?> Settings { get; } = new(StringComparer.Ordinal)
    {
        ["Auth:Enabled"] = "true",
        ["Auth:Authority"] = Authority,
        ["Auth:Audience"] = Audience,
    };

    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        base.ConfigureWebHost(builder);
        builder.ConfigureTestServices(services =>
            // Registered after AddJwtBearer's own post-configure, so this wins:
            // point metadata retrieval at the stub JWKS transport. Everything
            // else (issuer/audience pinning, algorithm allowlist) stays as
            // Program.cs configured it.
            services.PostConfigure<JwtBearerOptions>(JwtBearerDefaults.AuthenticationScheme, options =>
                options.ConfigurationManager = new ConfigurationManager<OpenIdConnectConfiguration>(
                    $"{Authority}/.well-known/openid-configuration",
                    new OpenIdConnectConfigurationRetriever(),
                    new HttpDocumentRetriever(CreateStubIssuerClient()))));
    }

    /// <summary>Mint a signed JWT the way the issuer would.</summary>
    public string CreateToken(
        string subject,
        string tenant,
        IReadOnlyList<string> roles,
        string? audience = Audience,
        bool sign = true)
    {
        var descriptor = new SecurityTokenDescriptor
        {
            Issuer = Authority,
            Audience = audience,
            Expires = DateTime.UtcNow.AddMinutes(5),
            Claims = new Dictionary<string, object>(StringComparer.Ordinal)
            {
                ["sub"] = subject,
                ["tenant"] = tenant,
                ["roles"] = roles,
            },
            SigningCredentials = sign
                ? new SigningCredentials(SigningKey(), SecurityAlgorithms.RsaSha256)
                : null, // alg=none - the allowlist must reject it
        };
        return new JsonWebTokenHandler().CreateToken(descriptor);
    }

    public static AuthenticationHeaderValue Bearer(string token) => new("Bearer", token);

    protected override void Dispose(bool disposing)
    {
        if (disposing)
        {
            _rsa.Dispose();
        }

        base.Dispose(disposing);
    }

    private RsaSecurityKey SigningKey() => new(_rsa) { KeyId = KeyId };

    private HttpClient CreateStubIssuerClient()
    {
        RSAParameters publicKey = _rsa.ExportParameters(includePrivateParameters: false);
        string modulus = Base64UrlEncoder.Encode(publicKey.Modulus);
        string exponent = Base64UrlEncoder.Encode(publicKey.Exponent);
        string jwks = $$"""
            {"keys":[{"kty":"RSA","use":"sig","kid":"{{KeyId}}","alg":"RS256","n":"{{modulus}}","e":"{{exponent}}"}]}
            """;
        string discovery = $$"""
            {"issuer":"{{Authority}}","jwks_uri":"{{Authority}}/jwks"}
            """;
        // The HttpClient takes ownership of the handler and disposes it with
        // itself (disposeHandler defaults to true); CA2000 cannot see that.
#pragma warning disable CA2000 // Dispose objects before losing scope
        return new HttpClient(new StubIssuerHandler(discovery, jwks));
#pragma warning restore CA2000
    }

    /// <summary>Serves the two issuer documents JWKS validation fetches; nothing else.</summary>
    private sealed class StubIssuerHandler(string discovery, string jwks) : HttpMessageHandler
    {
        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request, CancellationToken cancellationToken)
        {
            string? body = request.RequestUri?.AbsoluteUri switch
            {
                $"{Authority}/.well-known/openid-configuration" => discovery,
                $"{Authority}/jwks" => jwks,
                _ => null,
            };
            HttpResponseMessage response = body is null
                ? new HttpResponseMessage(HttpStatusCode.NotFound)
                : new HttpResponseMessage(HttpStatusCode.OK)
                {
                    Content = new StringContent(body, Encoding.UTF8, "application/json"),
                };
            return Task.FromResult(response);
        }
    }
}
