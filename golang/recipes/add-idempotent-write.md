# Recipe: Add Idempotent Write

Use this when a non-GET endpoint must be safe to retry — a payment, a charge, a resource creation — so that a client retry (or a retry injected by [../operations/resilience.md](../operations/resilience.md)) does not double-apply the side effect. The mechanism is server-side HTTP write idempotency keyed off a client-supplied `Idempotency-Key` header: the first request does the work and records its response; a replay returns the recorded response without re-running the work.

This is the REST-write counterpart to consumer/message idempotency, which lives with the inbox/dedupe model in [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md) and [../services/database.md](../services/database.md) (Outbox And Inbox). Use that for broker-delivered duplicates; use this recipe for HTTP retries.

## Files To Touch

- the idempotency middleware or store under `internal/api/http` (transport-owned interception of unsafe methods), or a `core`-seam helper if the dedupe must wrap a multi-step core operation
- a key-store table migration under `internal/db/migrations` (see [../recipes/add-migration.md](add-migration.md))
- the query/repository code that reads, claims, and completes a key under `internal/db`
- the write-path handler/core method, so the response is persisted in the **same transaction** as the write (the production-grade target — see the honest note below for the capture-and-replay form the keystone reference ships)
- middleware/store tests plus an integration test against the real schema

## Steps

1. **Require or accept the header on unsafe methods.** For `POST`/`PATCH`/`DELETE` that mutate, read `Idempotency-Key`. Decide per route whether it is required (recommended for payments/creation: reject a missing key with `400`) or optional (best-effort). Validate the key shape — bound the length and charset (e.g. an opaque UUID/ULID), do not accept arbitrary unbounded strings.
2. **Scope the key.** The lookup key is `(principal_or_tenant, route, idempotency_key)`, never the bare header value. Scoping by tenant/principal means two tenants reusing the same key string cannot collide or read each other's stored response.
3. **Compute a request fingerprint.** Hash the canonical request (method + route + relevant headers + a stable digest of the body) into a `request_fingerprint`. This is what detects a key being reused for a *different* request.
4. **Claim the key, then process in one transaction (production-grade target).** On first use, in a single DB transaction: insert the key row in an `in-flight` state, run the write, and persist the response (status code + body bytes + content type) against that key row, then commit. The idempotency record and the domain write commit atomically — see [../services/database.md](../services/database.md) (Transaction Rules). Use the insert's unique-constraint conflict as the concurrency gate so two simultaneous first-uses race on the row, not on application logic. This single-transaction claim-write-complete seam is the production-grade target; see the honest note below for the simpler form the keystone reference ships.
5. **Replay a completed duplicate.** On a duplicate key whose stored record is `COMPLETED` and whose `request_fingerprint` matches, return the stored response **byte-identical** (same status, same body, same content type). Do not re-run core. Optionally add a header (e.g. `Idempotent-Replayed: true`) for observability, but the response body must not change.
6. **Reject an in-flight duplicate.** On a duplicate key still `IN-FLIGHT` (the original request has not committed yet), return `409 Conflict` (or `425 Too Early`) so the client retries later rather than racing a second execution. Pick one status and document it as part of the contract per [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md).
7. **Reject a fingerprint mismatch.** On a duplicate key whose `request_fingerprint` differs from the stored one, return `422 Unprocessable Entity` — the same key was reused for a different request, which is a client bug, and replaying the old response would be wrong.
8. **Expire keys.** Stamp each record with a creation time and document a TTL (e.g. 24h). A reaper job or a partial index/`DELETE` bounds the table; an expired key is treated as never-seen. Document the TTL where clients can find it — a key replayed after expiry re-executes, so the TTL must exceed the client's realistic retry window.

## What The Keystone Reference Ships (Honest Note)

The reference module (`golang/reference/exampleservice`) ships the simpler **capture-and-replay** form, not single-transaction atomicity, and says so in code and README. Its idempotency middleware lets the handler commit the domain write, then calls `store.Complete` to persist the captured status+body **after** the write has already committed. Both the in-memory store and the SQL store's `Complete` are a **separate** write, not part of the domain write's transaction. The consequence is explicit: a crash between the committed write and the persisted record re-executes the side effect on the next retry — capture-and-replay narrows the double-apply window but does not eliminate it.

The single-transaction claim-write-complete seam in Step 4 is therefore the **production-grade target**, not a description of what the reference already does. To reach it on the SQL path, run claim + write + complete in one `*sql.Tx` (the generated query layer exposes `Queries.WithTx(*sql.Tx)` for exactly this) so the response row commits atomically with the write. The in-memory store has no transaction and stays documented as not single-transaction. Treat the invariant below as the bar a production implementation must clear, and the reference's capture-and-replay as the documented, lower-guarantee starting point.

## Invariants To Preserve

- **Atomic commit (production target).** The domain write and the idempotency record commit in **one** transaction. Never commit the write and then write the key row separately — a crash between them either double-applies on retry or strands a completed write with no record. The keystone reference does **not** yet meet this bar (see the honest note above); it ships capture-and-replay and documents the gap. A production implementation MUST close it with the single-`*sql.Tx` seam.
- **Byte-identical replay.** A completed-key replay returns the exact stored status, body, and content type. Re-deriving the response (re-serializing live state) is forbidden; persist the bytes that were sent.
- **Tenant-scoped keys.** Keys are scoped to `(principal/tenant + route)` so they cannot collide or leak responses across tenants.
- **Fingerprint enforcement.** Same key + different request is rejected (`422`), never silently replayed and never silently re-executed.
- **In-flight is not a second execution.** A concurrent duplicate while the first is still open returns `409`/`425`; exactly one execution wins via the unique constraint.
- **Bounded table.** A documented TTL plus reaping keeps the key store from growing unbounded.
- **Standard error envelope.** The `400` (missing key), `409`/`425` (in-flight), and `422` (fingerprint mismatch) responses use the repo's structured error envelope with a machine-readable `code` (see [../foundations/serialization.md](../foundations/serialization.md#error-responses)).

## Proof

- a test issuing the **same key twice** (sequentially) asserts exactly **one** side effect (one row created / one charge) and that both responses are byte-identical (status, body, content type)
- an **in-flight-duplicate** test (first request holds the transaction open, or two concurrent requests) asserts the second returns `409` (or the documented `425`) and that still only one execution commits — run under `make race` to prove the claim path is race-free
- a **different-body-same-key** test asserts `422` and that no second side effect occurred
- a **missing-key** test on a required route asserts `400`; an **expired-key** test asserts the key re-executes after TTL
- an integration test runs the migration and exercises the real `(tenant, route, key)` unique constraint and the same-transaction commit (see [../services/database.md](../services/database.md) Verification And Proof)
- handler tests use `httptest` per [../services/http-services.md](../services/http-services.md); the response envelope and error statuses are part of the wire contract per [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md) and reviewed accordingly
- `make verify`
