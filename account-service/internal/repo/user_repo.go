package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"

	"task-tracker/account-service/internal/domain"
	"task-tracker/shared-libs/pkg/logger"
)

type UserRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) UserRepository {
	return UserRepository{conn: conn}
}

func (r *UserRepository) CreateWithOutboxEvent(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query, args, err := squirrel.Insert("users").
		Columns("email", "password").
		Values(user.Email, user.PasswordHash).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build insert users query: %w", err)
	}
	logger.Log.Infof("sql: %s", query)

	var id int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	outboxQuery, outboxArgs, err := squirrel.Insert("outbox_events").
		Columns("topic", "message_key", "payload").
		Values(event.Topic, event.Key, event.Payload).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build insert outbox query: %w", err)
	}
	logger.Log.Infof("sql: %s", outboxQuery)

	if _, err = tx.ExecContext(ctx, outboxQuery, outboxArgs...); err != nil {
		return domain.User{}, fmt.Errorf("insert outbox event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return domain.User{}, fmt.Errorf("commit tx: %w", err)
	}

	user.ID = id
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query, args, err := squirrel.Select("id", "email", "password").
		From("users").
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("select user: %w", err)
	}
	logger.Log.Infof("sql: %s", query)

	user := domain.User{}
	err = r.conn.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("select user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}
	query, args, err := squirrel.Select("id", "email", "password").
		From("users").
		Where(squirrel.Eq{"id": ids}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	logger.Log.Infof("sql: %s", query)
	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Log.Infof("close rows: %v", err)
		}
	}()

	var users []domain.User
	for rows.Next() {
		user := domain.User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash); err != nil {
			return nil, fmt.Errorf("select users: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	return users, nil
}
