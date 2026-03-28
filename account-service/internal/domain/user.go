package domain

import (
	"context"
	"errors"
	"time"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type OutboxEvent struct {
	ID        int64
	Topic     string
	Key       string
	Payload   []byte
	CreatedAt time.Time
}

var (
	ErrNotFound           = errors.New("not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	CreateWithOutboxEvent(ctx context.Context, user User, event OutboxEvent) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByIDs(ctx context.Context, ids []int64) ([]User, error)
}

type OutboxRepository interface {
	ClaimPending(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkProcessed(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
}
