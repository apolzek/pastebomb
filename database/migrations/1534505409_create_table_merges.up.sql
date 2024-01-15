CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    username VARCHAR(100),
    email VARCHAR(255),
    born_date DATE,
    password VARCHAR(255),
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255),
    content TEXT,
    category VARCHAR(255),
    user_id INT NULL,
    url_id VARCHAR(255),
    author VARCHAR(255),
    is_public BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expiration_date VARCHAR(10),
    secret_access VARCHAR(255),
    num_reports INT UNSIGNED DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);


-- CREATE DATABASE

-- migrate -database "mysql://root:1234@tcp(localhost:3306)/go_gin_gonic" -path database/migrations up

-- DROP DATABASE

-- migrate -database "mysql://root:1234@tcp(localhost:3306)/go_gin_gonic" -path database/migrations down

-- CREATE TABLE

-- migrate create -ext sql -dir db/migrations create_users_table