-- +migrate Up
CREATE TABLE IF NOT EXISTS media (
    id            UUID PRIMARY KEY,
    uploader_id   UUID        NOT NULL REFERENCES users(id),
    original_name TEXT        NOT NULL,
    storage_name  TEXT        NOT NULL UNIQUE,
    mime_type     TEXT        NOT NULL,
    size_bytes    BIGINT      NOT NULL,
    storage_path  TEXT        NOT NULL,
    public_url    TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_uploader_id ON media(uploader_id);
CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at);

-- +migrate Down
DROP TABLE IF EXISTS media;

