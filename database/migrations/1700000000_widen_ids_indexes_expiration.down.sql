DROP INDEX idx_posts_expiration_date ON posts;
DROP INDEX idx_posts_url_id ON posts;
DROP INDEX idx_users_username ON users;
DROP INDEX idx_users_email ON users;

ALTER TABLE posts MODIFY COLUMN expiration_date VARCHAR(10);
ALTER TABLE users MODIFY COLUMN id INT NOT NULL AUTO_INCREMENT;
ALTER TABLE posts MODIFY COLUMN user_id INT NULL;
ALTER TABLE posts MODIFY COLUMN id INT NOT NULL AUTO_INCREMENT;
