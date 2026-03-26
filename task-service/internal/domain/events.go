package domain

import "time"

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
