-- The admin flag replaces the old "owner: <name>" header check, which granted
-- administrative access to anyone who guessed a string.
ALTER TABLE users ADD COLUMN is_admin TINYINT(1) NOT NULL DEFAULT 0;

-- url_id is the public handle of a paste. Making it unique is what turns a
-- random slug collision into a loud insert failure the application retries,
-- instead of two pastes silently sharing one handle.
DROP INDEX idx_posts_url_id ON posts;
CREATE UNIQUE INDEX idx_posts_url_id ON posts (url_id);

-- TEXT tops out at 65535 bytes. A paste may carry 1 MiB of content and is
-- stored base64-encoded, roughly 1.37 MiB, so writes above 48 KiB were being
-- truncated. MEDIUMTEXT holds 16 MiB.
ALTER TABLE posts MODIFY COLUMN content MEDIUMTEXT;

-- The columns below are required by the application and were nullable.
ALTER TABLE users MODIFY COLUMN name VARCHAR(255) NOT NULL;
ALTER TABLE users MODIFY COLUMN username VARCHAR(100) NOT NULL;
ALTER TABLE users MODIFY COLUMN email VARCHAR(255) NOT NULL;
ALTER TABLE users MODIFY COLUMN password VARCHAR(255) NOT NULL;
ALTER TABLE users MODIFY COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1;
