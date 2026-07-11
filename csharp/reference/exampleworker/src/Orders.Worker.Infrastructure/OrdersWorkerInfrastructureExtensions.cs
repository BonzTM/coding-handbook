using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Options;

using Orders.Worker.Core.Events;
using Orders.Worker.Core.Messaging;
using Orders.Worker.Core.Orders;
using Orders.Worker.Infrastructure.Health;
using Orders.Worker.Infrastructure.Messaging;
using Orders.Worker.Infrastructure.Telemetry;

namespace Orders.Worker.Infrastructure;

/// <summary>
/// Composition entry point for the messaging layer. Program.cs calls
/// <see cref="AddOrdersWorkerMessaging"/> once - the only place the host
/// crosses the Infrastructure boundary
/// (csharp/foundations/solution-and-project-design.md). Every registration
/// below binds a Core port to this module's in-memory implementation; swapping
/// in a real broker or SQL stores changes THIS file only.
/// </summary>
public static class OrdersWorkerInfrastructureExtensions
{
    public static IServiceCollection AddOrdersWorkerMessaging(
        this IServiceCollection services, IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(services);
        ArgumentNullException.ThrowIfNull(configuration);

        AddMessagingOptions(services);
        AddMessagingPorts(services);
        AddDeliveryPipeline(services);
        AddTelemetryAndHealth(services);
        return services;
    }

    /// <summary>Fail-fast options: a malformed retry budget or empty topic
    /// kills the process at startup, before the consumer subscribes
    /// (csharp/foundations/configuration.md).</summary>
    private static void AddMessagingOptions(IServiceCollection services)
    {
        services.AddOptions<MessagingOptions>()
            .BindConfiguration(MessagingOptions.SectionName)
            .ValidateDataAnnotations()
            .ValidateOnStart();
        services.AddOptions<ConsumerOptions>()
            .BindConfiguration(ConsumerOptions.SectionName)
            .ValidateDataAnnotations()
            .ValidateOnStart();
        services.AddOptions<OutboxOptions>()
            .BindConfiguration(OutboxOptions.SectionName)
            .ValidateDataAnnotations()
            .ValidateOnStart();
    }

    /// <summary>The Core ports and their in-memory implementations. Concrete
    /// types are registered once and forwarded, so the gauges and tests
    /// observe the same instances the pipeline uses.</summary>
    private static void AddMessagingPorts(IServiceCollection services)
    {
        services.AddSingleton(provider => new InMemoryBroker(
            provider.GetRequiredService<IOptions<MessagingOptions>>().Value.ChannelCapacity));
        services.AddSingleton<IEventPublisher, BrokerEventPublisher>();
        services.AddSingleton<IMessageSource, BrokerMessageSource>();
        services.AddSingleton<InMemoryInboxStore>();
        services.AddSingleton<IInboxStore>(provider => provider.GetRequiredService<InMemoryInboxStore>());
        services.AddSingleton<InMemoryOutboxStore>();
        services.AddSingleton<IOutboxStore>(provider => provider.GetRequiredService<InMemoryOutboxStore>());
        services.AddSingleton<InMemoryDeadLetterStore>();
        services.AddSingleton<IDeadLetterStore>(provider => provider.GetRequiredService<InMemoryDeadLetterStore>());
    }

    /// <summary>The consume/relay pipeline: backoff from validated options,
    /// the TimeProvider-backed delayer, one handler per message scope, and the
    /// two BackgroundServices.</summary>
    private static void AddDeliveryPipeline(IServiceCollection services)
    {
        services.AddSingleton(provider =>
        {
            var consumer = provider.GetRequiredService<IOptions<ConsumerOptions>>().Value;
            return new BackoffPolicy(consumer.BaseBackoff, consumer.MaxBackoff);
        });
        services.AddSingleton<IRetryDelayer, TimeProviderRetryDelayer>();
        services.AddSingleton<IOrderEventProcessor>(provider =>
            new OrderProjector(provider.GetRequiredService<TimeProvider>()));
        // One handler per message: the consumer opens a DI scope per delivery
        // (csharp/services/eventing-and-messaging.md) - a database-backed
        // processor would get its scoped DbContext from that same scope.
        services.AddScoped<OrderEventHandler>();
        services.AddSingleton<OutboxRelay>();
        services.AddHostedService<OrderEventsConsumer>();
        services.AddHostedService<OutboxRelayService>();
    }

    private static void AddTelemetryAndHealth(IServiceCollection services)
    {
        services.AddSingleton<WorkerReadiness>();
        services.AddSingleton(provider =>
        {
            var metrics = new WorkerMetrics(provider.GetRequiredService<System.Diagnostics.Metrics.IMeterFactory>());
            var broker = provider.GetRequiredService<InMemoryBroker>();
            string topic = provider.GetRequiredService<IOptions<MessagingOptions>>().Value.Topic;
            var dlq = provider.GetRequiredService<InMemoryDeadLetterStore>();
            var outbox = provider.GetRequiredService<InMemoryOutboxStore>();
            metrics.RegisterQueueGauges(
                consumerLag: () => broker.Depth(topic),
                dlqDepth: () => dlq.Count,
                outboxPending: () => outbox.PendingCount);
            return metrics;
        });

        // /readyz gates on the consumer being subscribed and the broker
        // reachable; /livez never runs this check
        // (csharp/operations/observability.md).
        services.AddHealthChecks()
            .AddCheck<BrokerHealthCheck>("broker", tags: ["ready"]);
    }
}
