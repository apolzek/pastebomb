-- Posts.id and posts.user_id were INT but the Go model used int8,
-- capping growth. Widen to BIGINT to match the new int64 fields.
ALTER TABLE posts MODIFY COLUMN id BIGINT NOT NULL AUTO_INCREMENT;
ALTER TABLE posts MODIFY COLUMN user_id BIGINT NULL;

-- users.id stays BIGINT to keep the FK type aligned.
ALTER TABLE users MODIFY COLUMN id BIGINT NOT NULL AUTO_INCREMENT;

-- expiration_date was VARCHAR(10) and never enforced. Convert to a real
-- TIMESTAMP so queries can use range filters and indexes.
ALTER TABLE posts MODIFY COLUMN expiration_date TIMESTAMP NULL;

-- Lookups by email and username were full scans. Make both unique to also
-- enforce the application-level check that lived in the controllers.
CREATE UNIQUE INDEX idx_users_email ON users (email);
CREATE UNIQUE INDEX idx_users_username ON users (username);

-- url_id is the public slug, hit on every paste read.
CREATE INDEX idx_posts_url_id ON posts (url_id);

-- expiration_date is filtered on every public read; index helps the purge
-- job too.
CREATE INDEX idx_posts_expiration_date ON posts (expiration_date);
