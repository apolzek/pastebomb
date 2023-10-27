CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255),
    born_date DATE,
    password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);


CREATE TABLE posts (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255),
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    author VARCHAR(255),
    url VARCHAR(255)
);

-- CREATE DATABASE

-- migrate -database "mysql://root:1234@tcp(localhost:3306)/go_gin_gonic" -path database/migrations up

-- DROP DATABASE

-- migrate -database "mysql://root:1234@tcp(localhost:3306)/go_gin_gonic" -path database/migrations down

-- CREATE TABLE

-- migrate create -ext sql -dir db/migrations create_users_table