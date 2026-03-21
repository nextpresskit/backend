-- +migrate Up
CREATE TABLE IF NOT EXISTS plugins (
    id         UUID PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    slug       TEXT        NOT NULL UNIQUE,
    enabled    BOOLEAN     NOT NULL DEFAULT FALSE,
    version    TEXT        NOT NULL DEFAULT '1.0.0',
    config     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS plugins;

