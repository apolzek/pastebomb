CREATE TABLE users (
    id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR (255),
    address VARCHAR (255),
    email VARCHAR (255),
    born_date TIMESTAMP,
    password VARCHAR (255)
);

CREATE TABLE posts (
    id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255),
    content TEXT,
    datetime DATETIME,
    author VARCHAR(255),
    url VARCHAR(255),
    -- post_id VARCHAR(255),
    -- privacy_status ENUM('public', 'private', 'unlisted'),
    -- expiry_date DATETIME,
    -- comments TEXT,
    -- tags VARCHAR(255),
    -- secret VARCHAR(255)
);


-- CREATE DATABASE
-- migrate -database "mysql://root:1234@tcp(localhost:3306)/go_gin_gonic" -path database/migrations up

-- DROP DATABASE
-- migrate -database "mysql://root:1234@tcp(localhost:3306)/go_gin_gonic" -path database/migrations down

-- CREATE TABLE
-- migrate create -ext sql -dir db/migrations create_users_table