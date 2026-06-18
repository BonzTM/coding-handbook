-- +goose Up
-- +goose StatementBegin
CREATE TABLE widgets (
    id         TEXT        NOT NULL,
    tenant_id  TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    -- Composite key (tenant_id, id): a widget id is unique WITHIN a tenant, so
    -- two tenants may independently use the same id. This is the storage-layer
    -- expression of multi-tenancy; every query also filters on tenant_id.
    PRIMARY KEY (tenant_id, id)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Composite index supporting keyset pagination scoped per tenant: the List
-- query orders by (created_at, id) and resumes after an opaque cursor, so this
-- index makes each page an index range scan rather than a full sort. tenant_id
-- leads so a future tenant-scoped predicate is also covered.
CREATE INDEX widgets_tenant_created_id_idx
    ON widgets (tenant_id, created_at, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS widgets_tenant_created_id_idx;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS widgets;
-- +goose StatementEnd
