-- Up Migration
CREATE TABLE widgets (
  tenant_id TEXT NOT NULL,
  id UUID NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  version INTEGER NOT NULL CHECK (version > 0),
  CONSTRAINT widgets_pk PRIMARY KEY (tenant_id, id),
  CONSTRAINT widgets_tenant_name_key UNIQUE (tenant_id, name),
  CONSTRAINT widgets_name_length CHECK (char_length(name) BETWEEN 1 AND 100),
  CONSTRAINT widgets_description_length CHECK (
    description IS NULL OR char_length(description) <= 500
  )
);

CREATE INDEX widgets_tenant_created_id_idx
  ON widgets (tenant_id, created_at, id);

-- Down Migration
DROP TABLE widgets;
