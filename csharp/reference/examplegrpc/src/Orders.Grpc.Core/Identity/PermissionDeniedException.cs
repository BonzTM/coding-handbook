namespace Orders.Grpc.Core.Identity;

/// <summary>
/// The authenticated principal lacks the role a domain operation requires.
/// Raised by the domain service, mapped once at the transport boundary to
/// PERMISSION_DENIED (csharp/foundations/errors-and-logging.md; the analyzer
/// override for context-requiring exception constructors is CA1032 in
/// .editorconfig).
/// </summary>
public sealed class PermissionDeniedException(string subject, string requiredRole)
    : Exception($"Principal '{subject}' requires role '{requiredRole}'.")
{
    public string Subject { get; } = subject;

    public string RequiredRole { get; } = requiredRole;
}
