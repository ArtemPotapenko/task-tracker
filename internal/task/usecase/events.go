package usecase

import (
	"context"
	"time"
)

type UserExpiredSummary struct {
	UserID       int64
	Completed    int
	NotCompleted int
}

type ExpiredSummary struct {
	WindowStart time.Time
	WindowEnd   time.Time
	Users       []UserExpiredSummary
}

type TaskEventPublisher interface {
	PublishExpiredSummary(ctx context.Context, summary ExpiredSummary) error
}
