-- +migrate Up
CREATE TABLE IF NOT EXISTS posts (
    id           UUID PRIMARY KEY,
    author_id    UUID        NOT NULL REFERENCES users(id),
    title        TEXT        NOT NULL,
    slug         TEXT        NOT NULL UNIQUE,
    content      TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'draft', -- draft|published|archived
    published_at TIMESTAMPTZ NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS posts;

