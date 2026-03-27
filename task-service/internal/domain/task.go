package domain

import (
	"context"
	"errors"
	"time"
)

type Task struct {
	ID          int64
	UserID      int64
	Description string
	Status      TaskStatus
	CreatedAt   time.Time
	DueDate     time.Time
}

type TaskStatus int

const (
	CREATED TaskStatus = iota
	AT_WORK
	COMPLETED
	EXPIRED
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

type OutboxEvent struct {
	ID        int64
	Topic     string
	Key       string
	Payload   []byte
	CreatedAt time.Time
}

type TaskRepository interface {
	Create(ctx context.Context, task Task) (Task, error)
	GetByID(ctx context.Context, id int64) (Task, error)
	GetByIDAndUserID(ctx context.Context, id, userID int64) (Task, error)
	GetByUserIDAndDueDateBetween(ctx context.Context, userID int64, from, to time.Time) ([]Task, error)
	GetByDueDateBetween(ctx context.Context, from, to time.Time) ([]Task, error)
	GetByDueDateBetweenAndStatusNot(ctx context.Context, from, to time.Time, status TaskStatus) ([]Task, error)
	UpdateExpiredAndEnqueueSummary(ctx context.Context, ids []int64, summary ExpiredSummary) error
	UpdateStatusByIDAndUserID(ctx context.Context, id, userID int64, status TaskStatus) (Task, error)
	UpdateStatusByIDs(ctx context.Context, ids []int64, status TaskStatus) error
}

type OutboxRepository interface {
	ClaimPending(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkProcessed(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
}
