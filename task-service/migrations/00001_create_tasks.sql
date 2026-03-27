-- +goose Up
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    description TEXT NOT NULL,
    status INTEGER NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    due_date TIMESTAMPTZ NOT NULL
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

CREATE INDEX tasks_user_id_due_date_idx ON tasks (user_id, due_date);
CREATE INDEX tasks_due_date_idx ON tasks (due_date);
CREATE INDEX tasks_due_date_status_idx ON tasks (due_date, status);
CREATE INDEX outbox_events_pending_idx ON outbox_events (processed_at, locked_at, id);

-- +goose Down
DROP INDEX IF EXISTS outbox_events_pending_idx;
DROP INDEX IF EXISTS tasks_due_date_status_idx;
DROP INDEX IF EXISTS tasks_due_date_idx;
DROP INDEX IF EXISTS tasks_user_id_due_date_idx;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS tasks;
