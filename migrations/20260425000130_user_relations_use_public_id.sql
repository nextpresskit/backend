-- +migrate Up
-- posts.author_id / reviewer_user_id / last_edited_by_user_id
ALTER TABLE posts RENAME COLUMN author_id TO author_id_uuid;
ALTER TABLE posts ADD COLUMN author_id BIGINT;
UPDATE posts p SET author_id = u.public_id FROM users u WHERE u.id = p.author_id_uuid;
ALTER TABLE posts ALTER COLUMN author_id SET NOT NULL;
ALTER TABLE posts DROP COLUMN author_id_uuid;

ALTER TABLE posts RENAME COLUMN reviewer_user_id TO reviewer_user_id_uuid;
ALTER TABLE posts ADD COLUMN reviewer_user_id BIGINT;
UPDATE posts p SET reviewer_user_id = u.public_id FROM users u WHERE u.id = p.reviewer_user_id_uuid;
ALTER TABLE posts DROP COLUMN reviewer_user_id_uuid;

ALTER TABLE posts RENAME COLUMN last_edited_by_user_id TO last_edited_by_user_id_uuid;
ALTER TABLE posts ADD COLUMN last_edited_by_user_id BIGINT;
UPDATE posts p SET last_edited_by_user_id = u.public_id FROM users u WHERE u.id = p.last_edited_by_user_id_uuid;
ALTER TABLE posts DROP COLUMN last_edited_by_user_id_uuid;

-- pages.author_id
ALTER TABLE pages RENAME COLUMN author_id TO author_id_uuid;
ALTER TABLE pages ADD COLUMN author_id BIGINT;
UPDATE pages p SET author_id = u.public_id FROM users u WHERE u.id = p.author_id_uuid;
ALTER TABLE pages ALTER COLUMN author_id SET NOT NULL;
ALTER TABLE pages DROP COLUMN author_id_uuid;

-- media.uploader_id
ALTER TABLE media RENAME COLUMN uploader_id TO uploader_id_uuid;
ALTER TABLE media ADD COLUMN uploader_id BIGINT;
UPDATE media m SET uploader_id = u.public_id FROM users u WHERE u.id = m.uploader_id_uuid;
ALTER TABLE media ALTER COLUMN uploader_id SET NOT NULL;
ALTER TABLE media DROP COLUMN uploader_id_uuid;

-- post_coauthors.user_id
ALTER TABLE post_coauthors RENAME COLUMN user_id TO user_id_uuid;
ALTER TABLE post_coauthors ADD COLUMN user_id BIGINT;
UPDATE post_coauthors pc SET user_id = u.public_id FROM users u WHERE u.id = pc.user_id_uuid;
ALTER TABLE post_coauthors ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE post_coauthors DROP CONSTRAINT post_coauthors_pkey;
ALTER TABLE post_coauthors DROP COLUMN user_id_uuid;
ALTER TABLE post_coauthors ADD CONSTRAINT post_coauthors_pkey PRIMARY KEY (post_id, user_id);

-- post_changelog.user_id (nullable)
ALTER TABLE post_changelog RENAME COLUMN user_id TO user_id_uuid;
ALTER TABLE post_changelog ADD COLUMN user_id BIGINT;
UPDATE post_changelog pc SET user_id = u.public_id FROM users u WHERE u.id = pc.user_id_uuid;
ALTER TABLE post_changelog DROP COLUMN user_id_uuid;

-- +migrate Down
-- post_changelog.user_id
ALTER TABLE post_changelog RENAME COLUMN user_id TO user_id_int;
ALTER TABLE post_changelog ADD COLUMN user_id UUID;
UPDATE post_changelog pc SET user_id = u.id FROM users u WHERE u.public_id = pc.user_id_int;
ALTER TABLE post_changelog DROP COLUMN user_id_int;

-- post_coauthors.user_id
ALTER TABLE post_coauthors RENAME COLUMN user_id TO user_id_int;
ALTER TABLE post_coauthors ADD COLUMN user_id UUID;
UPDATE post_coauthors pc SET user_id = u.id FROM users u WHERE u.public_id = pc.user_id_int;
ALTER TABLE post_coauthors DROP CONSTRAINT post_coauthors_pkey;
ALTER TABLE post_coauthors DROP COLUMN user_id_int;
ALTER TABLE post_coauthors ADD CONSTRAINT post_coauthors_pkey PRIMARY KEY (post_id, user_id);

-- media.uploader_id
ALTER TABLE media RENAME COLUMN uploader_id TO uploader_id_int;
ALTER TABLE media ADD COLUMN uploader_id UUID;
UPDATE media m SET uploader_id = u.id FROM users u WHERE u.public_id = m.uploader_id_int;
ALTER TABLE media ALTER COLUMN uploader_id SET NOT NULL;
ALTER TABLE media DROP COLUMN uploader_id_int;

-- pages.author_id
ALTER TABLE pages RENAME COLUMN author_id TO author_id_int;
ALTER TABLE pages ADD COLUMN author_id UUID;
UPDATE pages p SET author_id = u.id FROM users u WHERE u.public_id = p.author_id_int;
ALTER TABLE pages ALTER COLUMN author_id SET NOT NULL;
ALTER TABLE pages DROP COLUMN author_id_int;

-- posts.author_id / reviewer_user_id / last_edited_by_user_id
ALTER TABLE posts RENAME COLUMN last_edited_by_user_id TO last_edited_by_user_id_int;
ALTER TABLE posts ADD COLUMN last_edited_by_user_id UUID;
UPDATE posts p SET last_edited_by_user_id = u.id FROM users u WHERE u.public_id = p.last_edited_by_user_id_int;
ALTER TABLE posts DROP COLUMN last_edited_by_user_id_int;

ALTER TABLE posts RENAME COLUMN reviewer_user_id TO reviewer_user_id_int;
ALTER TABLE posts ADD COLUMN reviewer_user_id UUID;
UPDATE posts p SET reviewer_user_id = u.id FROM users u WHERE u.public_id = p.reviewer_user_id_int;
ALTER TABLE posts DROP COLUMN reviewer_user_id_int;

ALTER TABLE posts RENAME COLUMN author_id TO author_id_int;
ALTER TABLE posts ADD COLUMN author_id UUID;
UPDATE posts p SET author_id = u.id FROM users u WHERE u.public_id = p.author_id_int;
ALTER TABLE posts ALTER COLUMN author_id SET NOT NULL;
ALTER TABLE posts DROP COLUMN author_id_int;
