-- +goose Up
-- +goose StatementBegin
-- idempotency_keys persists the outcome of idempotent unsafe writes
-- (recipes/add-idempotent-write.md). A row is reserved on first use
-- (response_status NULL = in flight) and updated with the captured response when
-- the write completes, so a duplicate request replays it instead of repeating
-- the side effect. The key is (tenant_id, route, idempotency_key) so a
-- client-supplied key is scoped per tenant and endpoint, never global.
CREATE TABLE idempotency_keys (
    tenant_id       TEXT        NOT NULL,
    route           TEXT        NOT NULL,
    idempotency_key TEXT        NOT NULL,
    request_hash    TEXT        NOT NULL,
    response_status INTEGER,
    response_body   BYTEA,
    created_at      TIMESTAMPTZ NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, route, idempotency_key)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Supports the TTL sweep that reclaims expired rows out of band.
CREATE INDEX idempotency_keys_expires_at_idx
    ON idempotency_keys (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idempotency_keys_expires_at_idx;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS idempotency_keys;
-- +goose StatementEnd
