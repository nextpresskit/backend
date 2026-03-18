-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    first_name TEXT        NOT NULL,
    last_name  TEXT        NOT NULL,
    email      TEXT        NOT NULL UNIQUE,
    password   TEXT        NOT NULL,
    active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- +migrate Down
DROP TABLE IF EXISTS users;

