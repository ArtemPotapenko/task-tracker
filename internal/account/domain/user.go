package domain

import (
	"context"
	"errors"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

var (
	ErrNotFound           = errors.New("not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByIDs(ctx context.Context, ids []int64) ([]User, error)
}
