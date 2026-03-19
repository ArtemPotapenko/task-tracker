-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE UNIQUE INDEX users_email_uq ON users (email);

-- +goose Down
DROP INDEX IF EXISTS users_email_uq;
DROP TABLE IF EXISTS users;
