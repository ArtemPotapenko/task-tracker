-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE outbox_events (
    id BIGSERIAL PRIMARY KEY,
    topic TEXT NOT NULL,
    message_key TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX users_email_uq ON users (email);
CREATE INDEX outbox_events_pending_idx ON outbox_events (processed_at, locked_at, id);

-- +goose Down
DROP INDEX IF EXISTS outbox_events_pending_idx;
DROP INDEX IF EXISTS users_email_uq;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS users;
