namespace Orders.Worker.Infrastructure.Health;

/// <summary>
/// The worker's readiness flag: false until the consumer is subscribed,
/// flipped back to false the moment shutdown begins (before the drain), so the
/// platform stops routing readiness-gated traffic while in-flight work
/// finishes. Liveness is independent and stays green during the drain - a
/// draining pod must shed traffic, not be killed
/// (csharp/operations/observability.md).
/// </summary>
public sealed class WorkerReadiness
{
    private volatile bool _ready;

    public bool IsReady => _ready;

    public void SetReady(bool ready) => _ready = ready;
}
