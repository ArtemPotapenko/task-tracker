-- +goose Up
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    description TEXT NOT NULL,
    status INTEGER NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    due_date TIMESTAMPTZ NOT NULL
);

CREATE INDEX tasks_user_id_due_date_idx ON tasks (user_id, due_date);
CREATE INDEX tasks_due_date_idx ON tasks (due_date);
CREATE INDEX tasks_due_date_status_idx ON tasks (due_date, status);

-- +goose Down
DROP INDEX IF EXISTS tasks_due_date_status_idx;
DROP INDEX IF EXISTS tasks_due_date_idx;
DROP INDEX IF EXISTS tasks_user_id_due_date_idx;
DROP TABLE IF EXISTS tasks;
