# exampleworker

A compiling reference **event worker** for the Go engineering handbook. It is
the eventing counterpart to `exampleservice`: where that exemplar shows an HTTP
service, this one shows a broker-neutral consumer/relay worker with the rigor
the handbook expects at a messaging boundary —
[golang/services/eventing-and-messaging.md](../../services/eventing-and-messaging.md),
plus the [add-event-consumer](../../recipes/add-event-consumer.md) and
[add-event-publisher](../../recipes/add-event-publisher.md) recipes.

It mirrors `exampleservice`'s structure, conventions, and infra exactly: a thin
`cmd/exampleworker/main.go` (signal context + `x/sync/errgroup` + ordered,
bounded shutdown), `internal/{config,core,telemetry,buildinfo,testutil}`, and
the same Makefile / linter / Docker / editor config (with the module path
substituted).

```text
cmd/exampleworker/main.go     thin lifecycle wiring; errgroup runs consumer +
                              outbox relay + probe/metrics sidecar; ordered drain
internal/messaging/
  broker.go                   broker-neutral Broker interface + Message (Ack/Nack)
  memory.go                   in-memory, channel-backed Broker (offline-testable)
  inbox.go                    durable dedupe store keyed by message id (exactly-once)
  consumer.go                 consume loop: decode -> dedupe -> process -> retry/DLQ
  backoff.go                  exponential backoff + full jitter; Waiter seam
  dlq.go                      dead-letter store (retains attempts, class, reason)
  outbox.go                   transactional-outbox store + relay (publish -> mark sent)
internal/core/                domain: apply a widget event behind a Processor seam
internal/health/              probe/metrics HTTP sidecar (/livez /readyz /metrics)
internal/config/              env+flag load, fail-fast Validate
internal/telemetry/           slog logger, readiness, metrics seam (Nop + Prometheus),
                              config-gated OTel tracing
internal/buildinfo/           build metadata stamped via -ldflags
internal/testutil/            FakeClock (no real sleeps in tests)
```

## Design

A specific broker (Kafka, NATS, RabbitMQ, SQS, ...) is an ADR/framework
selection decision, so application code stays **broker-neutral**: the consumer
and relay depend only on the small `messaging.Broker` interface. The reference
wires an in-memory, channel-backed broker so the whole flow is offline-testable
under `-race`; a real client plugs into the same interface with a one-line change
in `main`.

The consumer follows the handbook's consumer rules: decode, validate, **dedupe**
(an inbox keyed by message id, so an at-least-once duplicate is processed exactly
once), call core logic, then ack. Transient failures **retry with bounded
exponential backoff + full jitter** computed from the injectable clock; a
non-retryable validation/schema failure or an exhausted retry budget moves the
message to the **DLQ** with its attempt count and failure class. The
**transactional outbox** relay drains pending rows, publishes them, and marks
them sent only after a successful publish, so a crash between the DB commit and
the publish is recovered on the next scan.

Time is injected via `core.Clock`; tests use `testutil.FakeClock` and a
clock-driven `Waiter`, so retry/backoff tests are deterministic with **no real
sleeps**.

## Shutdown (ordered graceful drain)

On SIGINT/SIGTERM the root context is cancelled and the worker drains in order:
stop pulling new messages (the subscription channel closes, the consume loop
finishes/acks in-flight work within the grace), the outbox relay performs a final
flush, the probe sidecar drains, the broker is closed, and telemetry is flushed
last. Readiness flips to unready first; liveness stays green so the platform does
not kill the pod mid-drain.

## Commands

`make verify` is the single ordered safety gate (gofmt + golangci-lint + vet +
test + race + govulncheck + build). Tools (golangci-lint v2, govulncheck) are
pinned as module `go tool` dependencies, never global installs.

```sh
make verify   # full gate
make run      # run locally with the in-memory broker
make test     # tests only
```

## Configuration

Every key is documented in [.env.example](./.env.example). Precedence is flags >
environment > defaults; a malformed value is a fail-fast error naming the
offending key.
