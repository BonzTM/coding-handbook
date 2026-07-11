# Recipe: Add External Client

Use this when the repo needs to call another HTTP or gRPC service.

Governing docs: [`csharp/operations/resilience.md`](../operations/resilience.md) and [`csharp/foundations/shared-constructs.md`](../foundations/shared-constructs.md) (Clients).

## Files To Touch

- `src/Orders.Core/Ports/I<Name>.cs` — the port, defined from the consumer's perspective
- `src/Orders.Infrastructure/Clients/<Name>Client.cs` — the typed-client adapter, plus its wire DTOs (they live here only)
- `src/Orders.Infrastructure/Clients/<Name>Options.cs` + `appsettings.json` — base address, credentials source, timeouts (follow [add-config-key.md](add-config-key.md))
- `src/Orders.Api/Program.cs` — `AddHttpClient` registration in the composition root
- client tests in `tests/Orders.UnitTests` using a stub `HttpMessageHandler`

## Steps

1. Define the dependency seam as an interface in `Orders.Core`, named for what the consumer needs (`IPaymentGateway.AuthorizeAsync(...)`), not for the vendor's API surface. Core references only this port.
2. Implement the adapter in `Orders.Infrastructure/Clients` as a typed client: a class taking `HttpClient` in its constructor, registered through `IHttpClientFactory`. Base address, auth, and timeouts come from the validated options class — never hardcoded, never from request input.
3. Keep request and response mapping in the client class only: wire DTOs (source-generated `System.Text.Json` per [../foundations/serialization.md](../foundations/serialization.md)) map to Core types at this boundary; nothing upstream sees the vendor's shapes.
4. Instrument requests at the client boundary: the factory's logging plus OpenTelemetry `HttpClient` instrumentation cover the basics; add an explicit metric per [add-metric.md](add-metric.md) when the dashboard needs a domain view.
5. Decide what is retryable and what is terminal before enabling automatic retries. Retries on non-idempotent calls need an idempotency contract with the upstream (send an `Idempotency-Key`; see [add-idempotent-write.md](add-idempotent-write.md) for the server-side model) — otherwise restrict retry to idempotent methods.

### Outbound HTTP Transport

Never construct `new HttpClient()` per call site: a per-request client burns sockets faster than the OS reclaims them, and a single cached `HttpClient` held forever never observes DNS changes. `IHttpClientFactory` solves both — it pools and rotates message handlers on a bounded lifetime — so every outbound client in the repo goes through `AddHttpClient`. Register exactly one named/typed client per upstream and let DI hand the adapter its `HttpClient`:

```csharp
builder.Services
    .AddHttpClient<IPaymentGateway, PaymentGatewayClient>((sp, client) =>
    {
        var opts = sp.GetRequiredService<IOptions<PaymentGatewayOptions>>().Value;
        client.BaseAddress = opts.BaseAddress;
        client.Timeout = TimeSpan.FromSeconds(60); // outermost backstop, above the pipeline's total timeout
    })
    .AddStandardResilienceHandler(o =>
    {
        o.AttemptTimeout.Timeout = TimeSpan.FromSeconds(5);       // cap a single try
        o.TotalRequestTimeout.Timeout = TimeSpan.FromSeconds(30); // cap the whole operation, retries included
        o.Retry.MaxRetryAttempts = 3;
    });
```

Set both layers of deadline. `AddStandardResilienceHandler` (Microsoft.Extensions.Http.Resilience) supplies the per-attempt timeout, total-request timeout, bounded retry with backoff, circuit breaker, and rate limiter in the documented composition order — tune it per upstream instead of hand-rolling policy. `HttpClient.Timeout` stays as the outermost backstop above the pipeline's total timeout. Every call also passes the caller's `CancellationToken` so cancellation propagates end to end.

Dispose every `HttpResponseMessage` (`using var response = ...`) on success and failure paths so the connection returns to the pool. For large bodies, pass `HttpCompletionOption.ResponseHeadersRead` and stream, rather than buffering. Never capture the injected `HttpClient` in a singleton field outside the factory's management — the typed client itself is registered transient and receives a fresh, correctly-rotated client. The transport handles connection hygiene; the resilience pipeline handles policy — keep the two concerns separate per [../operations/resilience.md](../operations/resilience.md).

## Invariants To Preserve

- caller `CancellationToken` and deadlines flow into every outbound request
- secrets stay out of logs; credentials come from config/secret sources per [../operations/security.md](../operations/security.md)
- SSRF or arbitrary-destination risk is constrained: the base address is a validated absolute URI from options, never composed from request input
- retries are bounded and idempotent-safe; the circuit breaker and total timeout cap worst-case latency
- no `new HttpClient()` outside `IHttpClientFactory`; one registered typed client per upstream with explicit attempt and total timeouts
- every response is disposed exactly once; wire DTOs never leak past the client class

## Proof

- client tests driving the adapter through a hand-rolled stub `HttpMessageHandler` (override `SendAsync`, return canned responses, capture the request for assertions) — no network:

  ```csharp
  private sealed class StubHandler(Func<HttpRequestMessage, HttpResponseMessage> respond) : HttpMessageHandler
  {
      protected override Task<HttpResponseMessage> SendAsync(
          HttpRequestMessage request, CancellationToken cancellationToken)
          => Task.FromResult(respond(request));
  }
  // new PaymentGatewayClient(new HttpClient(new StubHandler(...)) { BaseAddress = new("https://stub") })
  ```

- timeout and cancellation tests: a stub handler that delays past the deadline (honoring its `CancellationToken`) proves the attempt timeout and caller cancellation fire instead of hanging
- negative tests for auth failures (`401` from the stub maps to the port's terminal error) and malformed upstream responses (invalid JSON surfaces as a typed client error, not an unhandled exception)
- config validation proving required client settings (base address, credential key) fail fast at startup per [add-config-key.md](add-config-key.md)
- run `pwsh ./verify.ps1`

If the external system interaction is asynchronous publish or subscribe rather than request/response, use [add-event-publisher.md](add-event-publisher.md) or [add-event-consumer.md](add-event-consumer.md) instead.
