using Microsoft.Extensions.Options;

namespace Orders.Grpc.Api.Tls;

/// <summary>
/// Config-gated transport security (csharp/services/grpc-services.md):
/// cert + key set = TLS; additionally a client CA bundle = mutual TLS
/// (the default posture for internal service-to-service traffic unless a
/// mesh terminates TLS); neither = the insecure plaintext listener for
/// local/dev ONLY, logged loudly at startup. A configured-but-unloadable
/// key pair fails startup - never a silent downgrade to plaintext.
/// </summary>
internal sealed class TlsOptions
{
    public const string SectionName = "Tls";

    /// <summary>PEM server certificate path. Set together with <see cref="KeyPath"/>.</summary>
    public string CertPath { get; init; } = string.Empty;

    /// <summary>PEM private key path.</summary>
    public string KeyPath { get; init; } = string.Empty;

    /// <summary>PEM CA bundle; when set, clients MUST present a certificate this CA signed.</summary>
    public string ClientCaPath { get; init; } = string.Empty;

    public bool HasServerTls => !string.IsNullOrWhiteSpace(CertPath);

    public bool HasClientCa => !string.IsNullOrWhiteSpace(ClientCaPath);
}

/// <summary>
/// Cross-field invariants (csharp/foundations/configuration.md): cert and key
/// travel together, and mTLS requires server TLS. Enforced at startup via
/// ValidateOnStart - a half-configured key pair is a config error, not a
/// warning.
/// </summary>
internal sealed class TlsOptionsValidator : IValidateOptions<TlsOptions>
{
    public ValidateOptionsResult Validate(string? name, TlsOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        var failures = new List<string>(capacity: 2);
        if (string.IsNullOrWhiteSpace(options.CertPath) != string.IsNullOrWhiteSpace(options.KeyPath))
        {
            failures.Add("Tls:CertPath and Tls:KeyPath must be set together.");
        }

        if (options.HasClientCa && !options.HasServerTls)
        {
            failures.Add("Tls:ClientCaPath requires Tls:CertPath and Tls:KeyPath (mTLS requires server TLS).");
        }

        return failures.Count == 0
            ? ValidateOptionsResult.Success
            : ValidateOptionsResult.Fail(failures);
    }
}
