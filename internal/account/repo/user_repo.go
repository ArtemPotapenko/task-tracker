package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"

	"task-tracker/internal/account/domain"
)

type UserRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) UserRepository {
	return UserRepository{conn: conn}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	query, args, err := squirrel.Insert("users").
		Columns("email", "password").
		Values(user.Email, user.PasswordHash).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build insert users query: %w", err)
	}

	var id int64
	if err := r.conn.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
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
	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

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
