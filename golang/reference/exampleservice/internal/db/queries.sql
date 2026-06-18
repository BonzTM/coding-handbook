-- queries.sql is the sqlc source of truth for the widgets store. sqlc generates
-- the typed Go in internal/db/sqlcgen from these annotated queries against the
-- goose migration schema (see sqlc.yaml). Regenerate with: go tool sqlc generate.
--
-- Every read and write is scoped by tenant_id so one tenant can never observe
-- another's rows: multi-tenancy is enforced in the WHERE clause, not only at the
-- transport edge.

-- name: CreateWidget :exec
INSERT INTO widgets (id, tenant_id, name, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetWidget :one
SELECT id, tenant_id, name, created_at
FROM widgets
WHERE tenant_id = $1 AND id = $2;

-- name: ListWidgets :many
-- Keyset pagination over the stable (created_at, id) order WITHIN a tenant. The
-- tenant_id predicate scopes every page; when @from_start is true the
-- row-comparison guard is bypassed and the scan starts at the tenant's first
-- row; otherwise it resumes strictly after the cursor (@cursor_created_at,
-- @cursor_id) using a tuple comparison so id breaks created_at ties and pages
-- neither overlap nor skip rows. The caller asks for one more than the page size
-- to detect whether another page exists.
SELECT id, tenant_id, name, created_at
FROM widgets
WHERE
    tenant_id = sqlc.arg(tenant_id)
    AND (
        sqlc.arg(from_start)::boolean
        OR (created_at, id) > (sqlc.arg(cursor_created_at)::timestamptz, sqlc.arg(cursor_id)::text)
    )
ORDER BY created_at, id
LIMIT sqlc.arg(page_limit);

-- Idempotency-Key store (recipes/add-idempotent-write.md). Keyed by
-- (tenant_id, route, idempotency_key) so a client key is scoped per tenant and
-- endpoint. A row is reserved (response_status NULL) on first use and updated
-- with the response when the write completes, so a duplicate replays it. The
-- TTL is enforced with expires_at.

-- name: InsertIdempotency :execrows
-- Reserve the key on first use; if a row already exists it is reclaimed ONLY
-- when expired (expires_at <= now). The affected-row count tells the caller
-- whether it won the lease (1) or a live row already exists (0); on 0 the caller
-- reads the existing row to decide replay vs. in-flight vs. mismatch. This is the
-- atomic compare-and-set that closes the concurrent-duplicate race.
INSERT INTO idempotency_keys (tenant_id, route, idempotency_key, request_hash, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id, route, idempotency_key) DO UPDATE
    SET request_hash = EXCLUDED.request_hash,
        response_status = NULL,
        response_body = NULL,
        created_at = EXCLUDED.created_at,
        expires_at = EXCLUDED.expires_at
    WHERE idempotency_keys.expires_at <= EXCLUDED.created_at;

-- name: GetIdempotency :one
SELECT request_hash, response_status, response_body
FROM idempotency_keys
WHERE tenant_id = $1 AND route = $2 AND idempotency_key = $3 AND expires_at > $4;

-- name: CompleteIdempotency :exec
UPDATE idempotency_keys
SET response_status = $4, response_body = $5
WHERE tenant_id = $1 AND route = $2 AND idempotency_key = $3;

-- name: ReleaseIdempotency :exec
DELETE FROM idempotency_keys
WHERE tenant_id = $1 AND route = $2 AND idempotency_key = $3 AND response_status IS NULL;
