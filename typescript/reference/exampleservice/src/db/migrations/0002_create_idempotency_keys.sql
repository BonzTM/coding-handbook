-- Up Migration
CREATE TABLE idempotency_keys (
  tenant_id TEXT NOT NULL,
  operation TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  request_hash TEXT NOT NULL,
  widget_id UUID,
  expires_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT idempotency_keys_pk
    PRIMARY KEY (tenant_id, operation, idempotency_key),
  CONSTRAINT idempotency_key_length
    CHECK (char_length(idempotency_key) BETWEEN 1 AND 128),
  CONSTRAINT idempotency_operation_check
    CHECK (operation = 'widget.create')
);

CREATE INDEX idempotency_keys_expires_at_idx
  ON idempotency_keys (expires_at);

-- Down Migration
DROP TABLE idempotency_keys;
