using System.ComponentModel.DataAnnotations;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// The consume loop's bounded-retry policy (csharp/services/eventing-and-messaging.md,
/// Retries And Dead-Letter Behavior). Bound from the "Consumer" section with
/// ValidateOnStart; the cross-field invariant (max >= base) is enforced by
/// <see cref="Validate"/> so a bad pair fails at startup, not mid-retry.
/// </summary>
public sealed class ConsumerOptions : IValidatableObject
{
    public const string SectionName = "Consumer";

    /// <summary>Total delivery attempts (including the first) before a message
    /// is dead-lettered.</summary>
    [Range(1, 100)]
    public int MaxAttempts { get; init; } = 5;

    /// <summary>First retry's backoff ceiling; subsequent attempts double it
    /// (capped at <see cref="MaxBackoff"/>) before full jitter is applied.</summary>
    public TimeSpan BaseBackoff { get; init; } = TimeSpan.FromMilliseconds(100);

    /// <summary>Cap on the exponential backoff ceiling.</summary>
    public TimeSpan MaxBackoff { get; init; } = TimeSpan.FromSeconds(30);

    public IEnumerable<ValidationResult> Validate(ValidationContext validationContext)
    {
        if (BaseBackoff <= TimeSpan.Zero)
        {
            yield return new ValidationResult(
                "Consumer:BaseBackoff must be positive.", [nameof(BaseBackoff)]);
        }

        if (MaxBackoff < BaseBackoff)
        {
            yield return new ValidationResult(
                "Consumer:MaxBackoff must be >= Consumer:BaseBackoff.", [nameof(MaxBackoff)]);
        }
    }
}
