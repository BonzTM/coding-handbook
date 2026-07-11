using System.ComponentModel.DataAnnotations;

namespace Orders.Infrastructure.Data;

/// <summary>
/// Database connection configuration, bound from the standard
/// <c>ConnectionStrings</c> section (env override:
/// <c>ConnectionStrings__Default</c>). Validated at startup so a missing
/// connection string kills the process before listeners open
/// (csharp/foundations/configuration.md).
///
/// The connection string carries all four Npgsql pool limits explicitly
/// (Maximum/Minimum Pool Size, Connection Lifetime, Connection Idle Lifetime) -
/// never ship the defaults unexamined (csharp/services/database.md).
/// </summary>
public sealed class DatabaseOptions
{
    public const string SectionName = "ConnectionStrings";

    [Required(AllowEmptyStrings = false)]
    public string Default { get; init; } = string.Empty;
}
