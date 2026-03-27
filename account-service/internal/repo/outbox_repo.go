package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"task-tracker/account-service/internal/domain"
	"task-tracker/shared-libs/pkg/logger"
)

type OutboxRepository struct {
	conn *sql.DB
}

func NewOutboxRepository(conn *sql.DB) OutboxRepository {
	return OutboxRepository{conn: conn}
}

func (r *OutboxRepository) ClaimPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `
SELECT id, topic, message_key, payload, created_at
FROM outbox_events
WHERE processed_at IS NULL AND locked_at IS NULL
ORDER BY id ASC
LIMIT $1
FOR UPDATE SKIP LOCKED`
	logger.Log.Infof("sql: %s", query)

	rows, err := tx.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("select outbox events: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Log.Infof("close rows: %v", err)
		}
	}()

	events := make([]domain.OutboxEvent, 0, limit)
	ids := make([]int64, 0, limit)
	for rows.Next() {
		var event domain.OutboxEvent
		if err := rows.Scan(&event.ID, &event.Topic, &event.Key, &event.Payload, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, event)
		ids = append(ids, event.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select outbox events: %w", err)
	}
	if len(ids) == 0 {
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx: %w", err)
		}
		return nil, nil
	}

	lockQuery, lockArgs, err := squirrel.Update("outbox_events").
		Set("locked_at", time.Now().UTC()).
		Where(squirrel.Eq{"id": ids}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build lock outbox query: %w", err)
	}
	logger.Log.Infof("sql: %s", lockQuery)

	if _, err = tx.ExecContext(ctx, lockQuery, lockArgs...); err != nil {
		return nil, fmt.Errorf("lock outbox events: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return events, nil
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id int64) error {
	query, args, err := squirrel.Update("outbox_events").
		Set("processed_at", time.Now().UTC()).
		Set("locked_at", nil).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update outbox query: %w", err)
	}
	logger.Log.Infof("sql: %s", query)

	if _, err := r.conn.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update outbox events: %w", err)
	}
	return nil
}

func (r *OutboxRepository) Release(ctx context.Context, id int64) error {
	query, args, err := squirrel.Update("outbox_events").
		Set("locked_at", nil).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build release outbox query: %w", err)
	}
	logger.Log.Infof("sql: %s", query)

	if _, err := r.conn.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("release outbox event: %w", err)
	}
	return nil
}
