package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"task-tracker/internal/task/domain"
)

type TaskRepository struct {
	conn *sql.DB
}

const taskColumns = "id, user_id, description, status, date, due_date"

func NewTaskRepository(conn *sql.DB) TaskRepository {
	return TaskRepository{conn: conn}
}

func (r *TaskRepository) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	query, args, err := squirrel.Insert("tasks").
		Columns("user_id", "description", "status", "date", "due_date").
		Values(task.UserID, task.Description, task.Status, task.CreatedAt, task.DueDate).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("build insert tasks query: %w", err)
	}

	var id int64
	if err := r.conn.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return domain.Task{}, fmt.Errorf("insert task: %w", err)
	}

	task.ID = id
	return task, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	query, args, err := squirrel.Select(taskColumns).
		From("tasks").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("select task: %w", err)
	}

	task := domain.Task{}
	err = r.conn.QueryRowContext(ctx, query, args...).Scan(
		&task.ID,
		&task.UserID,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.DueDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("select task: %w", err)
	}
	return task, nil
}

func (r *TaskRepository) GetByIDAndUserID(ctx context.Context, id, userID int64) (domain.Task, error) {
	query, args, err := squirrel.Select(taskColumns).
		From("tasks").
		Where(squirrel.Eq{"id": id, "user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("select task: %w", err)
	}

	task := domain.Task{}
	err = r.conn.QueryRowContext(ctx, query, args...).Scan(
		&task.ID,
		&task.UserID,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.DueDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("select task: %w", err)
	}
	return task, nil
}

func (r *TaskRepository) GetByDueDateBetween(ctx context.Context, from, to time.Time) ([]domain.Task, error) {
	query, args, err := squirrel.Select(taskColumns).
		From("tasks").
		Where(squirrel.GtOrEq{"due_date": from}).
		Where(squirrel.Lt{"due_date": to}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		task := domain.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.DueDate,
		); err != nil {
			return nil, fmt.Errorf("select tasks: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	return tasks, nil
}

func (r *TaskRepository) GetByUserIDAndDueDateBetween(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) {
	query, args, err := squirrel.Select(taskColumns).
		From("tasks").
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.GtOrEq{"due_date": from}).
		Where(squirrel.Lt{"due_date": to}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		task := domain.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.DueDate,
		); err != nil {
			return nil, fmt.Errorf("select tasks: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	return tasks, nil
}

func (r *TaskRepository) GetByDueDateBetweenAndStatusNot(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
	query, args, err := squirrel.Select(taskColumns).
		From("tasks").
		Where(squirrel.GtOrEq{"due_date": from}).
		Where(squirrel.Lt{"due_date": to}).
		Where(squirrel.NotEq{"status": status}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}

	rows, err := r.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		task := domain.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.DueDate,
		); err != nil {
			return nil, fmt.Errorf("select tasks: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	return tasks, nil
}

func (r *TaskRepository) UpdateStatusByIDAndUserID(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
	query, args, err := squirrel.Update("tasks").
		Set("status", status).
		Where(squirrel.Eq{"id": id, "user_id": userID}).
		Suffix("RETURNING " + taskColumns).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}

	task := domain.Task{}
	err = r.conn.QueryRowContext(ctx, query, args...).Scan(
		&task.ID,
		&task.UserID,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.DueDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}
	return task, nil
}

func (r *TaskRepository) UpdateStatusByIDs(ctx context.Context, ids []int64, status domain.TaskStatus) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := squirrel.Update("tasks").
		Set("status", status).
		Where(squirrel.Eq{"id": ids}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("update tasks: %w", err)
	}

	if _, err := r.conn.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update tasks: %w", err)
	}
	return nil
}
