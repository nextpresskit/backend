-- +migrate Up
CREATE SEQUENCE IF NOT EXISTS users_public_id_seq;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS public_id BIGINT;

ALTER TABLE users
    ALTER COLUMN public_id SET DEFAULT nextval('users_public_id_seq');

UPDATE users
SET public_id = nextval('users_public_id_seq')
WHERE public_id IS NULL;

ALTER TABLE users
    ALTER COLUMN public_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_public_id_key ON users(public_id);

-- +migrate Down
DROP INDEX IF EXISTS users_public_id_key;
ALTER TABLE users ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE users DROP COLUMN IF EXISTS public_id;
DROP SEQUENCE IF EXISTS users_public_id_seq;
