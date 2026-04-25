-- +migrate Up
DO $$
DECLARE
    c record;
BEGIN
    FOR c IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'user_roles'::regclass
          AND contype IN ('f', 'p')
    LOOP
        EXECUTE format('ALTER TABLE user_roles DROP CONSTRAINT %I', c.conname);
    END LOOP;
END $$;

ALTER TABLE user_roles
    RENAME COLUMN user_id TO user_id_uuid;

ALTER TABLE user_roles
    ADD COLUMN user_id BIGINT;

UPDATE user_roles ur
SET user_id = u.public_id
FROM users u
WHERE u.id = ur.user_id_uuid;

ALTER TABLE user_roles
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE user_roles
    DROP COLUMN user_id_uuid;

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(public_id) ON DELETE CASCADE;

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

-- +migrate Down
DO $$
DECLARE
    c record;
BEGIN
    FOR c IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'user_roles'::regclass
          AND contype IN ('f', 'p')
    LOOP
        EXECUTE format('ALTER TABLE user_roles DROP CONSTRAINT %I', c.conname);
    END LOOP;
END $$;

ALTER TABLE user_roles
    RENAME COLUMN user_id TO user_id_int;

ALTER TABLE user_roles
    ADD COLUMN user_id UUID;

UPDATE user_roles ur
SET user_id = u.id
FROM users u
WHERE u.public_id = ur.user_id_int;

ALTER TABLE user_roles
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE user_roles
    DROP COLUMN user_id_int;

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
