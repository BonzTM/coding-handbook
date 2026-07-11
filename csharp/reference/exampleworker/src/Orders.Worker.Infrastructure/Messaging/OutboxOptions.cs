using System.ComponentModel.DataAnnotations;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// Transactional-outbox relay settings, bound from the "Outbox" section with
/// ValidateOnStart.
/// </summary>
public sealed class OutboxOptions : IValidatableObject
{
    public const string SectionName = "Outbox";

    /// <summary>How often the relay scans the outbox store for pending
    /// records. Also the budget for the final flush on shutdown.</summary>
    public TimeSpan PollInterval { get; init; } = TimeSpan.FromSeconds(1);

    /// <summary>Max pending records the relay claims per scan.</summary>
    [Range(1, 10_000)]
    public int BatchSize { get; init; } = 100;

    public IEnumerable<ValidationResult> Validate(ValidationContext validationContext)
    {
        if (PollInterval <= TimeSpan.Zero)
        {
            yield return new ValidationResult(
                "Outbox:PollInterval must be positive.", [nameof(PollInterval)]);
        }
    }
}
