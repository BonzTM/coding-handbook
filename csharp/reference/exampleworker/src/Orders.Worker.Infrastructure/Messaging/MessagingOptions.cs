using System.ComponentModel.DataAnnotations;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// Broker/topic settings. Bound from the "Messaging" section with
/// ValidateOnStart, so a malformed value kills the process before the consumer
/// subscribes (csharp/foundations/configuration.md).
/// </summary>
public sealed class MessagingOptions
{
    public const string SectionName = "Messaging";

    /// <summary>The topic/subject the consumer subscribes to and the outbox
    /// relay publishes to. Broker selection is an ADR decision; for the
    /// in-memory broker this is a logical queue name.</summary>
    [Required(AllowEmptyStrings = false)]
    public string Topic { get; init; } = "orders.events";

    /// <summary>Per-topic channel capacity of the in-memory broker. Bounded on
    /// purpose: an unbounded queue hides a losing consumer until memory runs
    /// out. Publishers wait when it is full.</summary>
    [Range(1, 1_000_000)]
    public int ChannelCapacity { get; init; } = 256;
}
