-- 004_layer.sql

DO $$ BEGIN
    CREATE TYPE layer_type AS ENUM ('strategic', 'tactical', 'operational');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS layer_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mvru_id         UUID NOT NULL REFERENCES muvrs(id) ON DELETE CASCADE,
    layer           layer_type NOT NULL,
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (mvru_id, layer)
);

CREATE TABLE IF NOT EXISTS layer_routing_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_layer    layer_type NOT NULL,
    target_layer    layer_type NOT NULL,
    condition       JSONB NOT NULL DEFAULT '{}',
    priority        INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_layer_config_mvru ON layer_configs(mvru_id);
