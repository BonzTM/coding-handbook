# exampleworker

A compiling reference **event worker** for the C# engineering handbook. It is
the eventing counterpart to [`exampleservice`](../exampleservice/): where that
keystone shows an HTTP service, this one shows a broker-neutral
consumer/relay worker with the rigor the handbook expects at a messaging
boundary — [csharp/services/eventing-and-messaging.md](../../services/eventing-and-messaging.md),
plus the [add-event-consumer](../../recipes/add-event-consumer.md),
[add-event-publisher](../../recipes/add-event-publisher.md), and
[add-background-worker](../../recipes/add-background-worker.md) recipes. The
language baseline is **.NET 10 / C# 14** (`global.json` pins the SDK).

It mirrors `exampleservice`'s structure, conventions, and root scaffolding
exactly (the same template copies, including that module's flagged template
fixes), with the project names substituted: a thin `Orders.Worker` host
(composition root + probe endpoints + bounded shutdown), `Orders.Worker.Core`
(domain + ports), `Orders.Worker.Infrastructure` (broker/store adapters +
delivery pipeline), and one offline test project.

```text
src/Orders.Worker/                  composition root; /livez /readyz on Kestrel;
                                    HostOptions.ShutdownTimeout bounds the drain
src/Orders.Worker.Core/
  Events/                           EventEnvelope (CloudEvents-style), the
                                    IEventPublisher/IMessageSource ports,
                                    InboundMessage (Ack/Nack), OrderEvent +
                                    source-generated JSON context, the
                                    invalid/transient failure classification
  Messaging/                        IInboxStore (durable dedupe), IOutboxStore
                                    (transactional outbox), IDeadLetterStore
                                    (attempts, class, reason), BackoffPolicy
                                    (exponential + full jitter), IRetryDelayer
  Orders/                           domain: project an order event behind the
                                    IOrderEventProcessor seam
src/Orders.Worker.Infrastructure/
  Messaging/InMemoryBroker.cs       channel-backed broker (offline-testable,
                                    at-least-once: nack redelivers)
  Messaging/BrokerAdapters.cs       the ONLY seam a real broker replaces
  Messaging/InMemoryStores.cs       inbox/outbox/DLQ stores (documented SQL seams)
  Messaging/OrderEventHandler.cs    consume pipeline: decode -> dedupe ->
                                    process -> retry/DLQ -> settle
  Messaging/OrderEventsConsumer.cs  consumer BackgroundService (scope per message)
  Messaging/OutboxRelay*.cs         outbox relay (publish -> mark sent) +
                                    its BackgroundService with final flush
  Telemetry/WorkerMetrics.cs        Meter: consumed/published counters,
                                    consumer-lag / DLQ-depth / outbox gauges
  Health/                           readiness flag + "ready" broker check
tests/Orders.Worker.UnitTests/      domain, pipeline, broker, store, relay,
                                    probe, drain, and metrics tests - all
                                    offline, no real sleeps
```

## Design

A specific broker (Kafka, NATS, RabbitMQ, SQS, ...) is an ADR/framework
selection decision, so application code stays **broker-neutral**: the consumer
and relay depend only on the small Core-owned `IEventPublisher` /
`IMessageSource` ports. The reference wires an in-memory, channel-backed
broker so the whole flow is offline-testable; a real client plugs into the
same ports by replacing the two adapters in
`Infrastructure/Messaging/BrokerAdapters.cs` — one registration change in the
composition root.

The consumer follows the handbook's consumer rules: decode, validate,
**dedupe** (an inbox keyed by envelope id, so an at-least-once duplicate is
processed exactly once), call core logic in a DI scope per message, then ack.
Transient failures **retry in-place with bounded exponential backoff + full
jitter**; a non-retryable validation/schema failure or an exhausted retry
budget moves the message to the **DLQ** with its attempt count, failure class,
and reason. The **transactional outbox** relay drains pending records,
publishes them, and marks them sent only after a successful publish, so a
crash between the state commit and the publish is recovered on the next scan.

**Honesty about storage:** like its Go counterpart, this module's inbox,
outbox, and DLQ are in-memory implementations of Core ports, NOT a database —
that keeps the whole module runnable and testable offline. Each port documents
its production seam precisely: a SQL inbox is an insert-on-conflict table
written in the same EF Core transaction as the handler's state change (making
the in-memory compensation hook unnecessary), the outbox is an EF Core entity
whose `AddAsync` is the INSERT inside the domain-write transaction with a
`SentAt IS NULL ... FOR UPDATE SKIP LOCKED` relay query, and the DLQ is a
topic or table with an operator-owned replay runbook. The same-transaction
atomicity itself is therefore *documented at the seam, not executed here* —
the real-database version of that proof lives in `exampleservice`'s
Testcontainers suite (its idempotency runner commits the claim, domain write,
and response in one transaction).

Time is injected as `TimeProvider` everywhere; retry waits go through an
`IRetryDelayer` seam (production: the TimeProvider-aware `Task.Delay`), so the
retry/backoff tests use `FakeTimeProvider` plus a recording delayer and run
with **no real sleeps**.

## Shutdown (bounded graceful drain)

On SIGTERM the generic host cancels the consume loop's `stoppingToken`: the
consumer stops pulling new messages (the subscription stream ends), finishes
and settles the message already in flight, the outbox relay performs a final
bounded flush, and the whole drain is capped by `HostOptions.ShutdownTimeout`
(15s, below Kubernetes' 30s default grace). Readiness flips to unready the
moment shutdown begins; liveness stays green so the platform does not kill the
pod mid-drain. The drain tests prove both halves: an in-flight message
finishes and is settled (never lost, never dead-lettered), and a wedged
handler cannot hold shutdown past the budget.

## Run it

```bash
export DOTNET_ROOT="$HOME/.dotnet"; export PATH="$HOME/.dotnet:$PATH"
dotnet run --project src/Orders.Worker    # consume from the in-memory broker

curl -s localhost:8080/livez
curl -s localhost:8080/readyz
```

The worker serves probes only — no application routes, no auth stack (the
counterpart of the Go reference's bare probe sidecar). With the in-memory
broker it runs idle until something publishes; the tests are where the flow is
exercised end to end. There is no `--migrate` mode because there is no
database — the seam documentation above says exactly where migrations enter
when the stores become real.

## Verify

`pwsh ./verify.ps1` is the single ordered gate — restore (locked),
format-check, build (warnings-as-errors), test, audit. Humans, the Makefile
shim, and CI run the same script.

```bash
pwsh ./verify.ps1                # the full offline gate
pwsh ./verify.ps1 -Integration   # documented no-op here (see below)
```

**`-Integration` is a no-op in this module by design:** every dependency is
in-memory behind a Core port, so there is no `tests/*IntegrationTests*`
project and nothing that needs Docker — the switch simply runs the same unit
suite. That matches the Go `exampleworker` (offline-only) and the handbook
rule that broker-semantics tests belong with a real broker once one is chosen
via ADR ([eventing-and-messaging.md](../../services/eventing-and-messaging.md),
Testing And Proof).

## Observability

Telemetry is wired once in `AddServiceTelemetry()`; all three signals export
over OTLP (`OTEL_EXPORTER_OTLP_*` environment variables). Like the keystone —
and unlike the Go worker's Prometheus sidecar — there is **no `/metrics`
endpoint by design**: OTLP push is the handbook default, and the OTel
Prometheus exporter is added only when the org scrapes.

- `orders.worker.messages.consumed` (counter; `event.type`, `outcome` =
  ack | retry | dropped_duplicate | dead_lettered) and
  `orders.worker.messages.published` (counter; `event.type`).
- `orders.worker.consumer.lag`, `orders.worker.dlq.depth`, and
  `orders.worker.outbox.pending` (observable gauges). Consumer lag is the
  earliest signal a consumer is losing — alert on it, not on the eventual
  timeout storm. The gauge *names* are the contract; their callbacks read the
  in-memory broker/stores here, where production reads broker metrics and
  monitored queries.
- Every delivery-state transition (retry, duplicate drop, dead-letter, settle
  failure) is a source-generated `[LoggerMessage]` log with stable fields;
  message ids live in logs and traces (`Orders.Worker` ActivitySource), never
  in metric labels.

## Intentionally out of scope

- No real broker client and no Testcontainers broker suite — broker selection
  is an ADR ([decisions/framework-selection.md](../../decisions/framework-selection.md));
  the ports and adapters mark the exact seam.
- No database, no EF Core, no `--migrate` — the storage seams are documented
  at the ports; the real-database proof lives in
  [`exampleservice`](../exampleservice/).
- No HTTP API surface — see `exampleservice`; no gRPC — see `examplegrpc`.
- No scheduled jobs — the relay's `PeriodicTimer` loop is the shape;
  [recipes/add-scheduled-job.md](../../recipes/add-scheduled-job.md) governs.
- No Kubernetes manifests — the committed
  [templates/k8s-deployment.yaml](../../templates/k8s-deployment.yaml) is the
  rollout shape.
