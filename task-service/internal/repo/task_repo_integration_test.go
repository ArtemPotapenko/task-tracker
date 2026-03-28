package repo

import (
	"context"
	"testing"
	"time"

	"task-tracker/shared-libs/pkg/testutil"
	"task-tracker/task-service/internal/domain"
)

func TestTaskRepositoryIntegration(t *testing.T) {
	db := testutil.StartPostgres(t, "taskdb")
	testutil.ApplyUpSQL(t, db, "../../migrations/00001_create_tasks.sql")

	if _, err := db.Exec("TRUNCATE TABLE tasks RESTART IDENTITY"); err != nil {
		t.Fatalf("TRUNCATE tasks error = %v", err)
	}

	repo := NewTaskRepository(db)
	outboxRepo := NewOutboxRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)

	created, err := repo.Create(ctx, domain.Task{
		UserID:      1,
		Description: "ship feature",
		Status:      domain.CREATED,
		CreatedAt:   now,
		DueDate:     now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByIDAndUserID(ctx, created.ID, 1)
	if err != nil {
		t.Fatalf("GetByIDAndUserID() error = %v", err)
	}
	if got.Description != "ship feature" {
		t.Fatalf("GetByIDAndUserID() description = %q", got.Description)
	}

	today, err := repo.GetByUserIDAndDueDateBetween(ctx, 1, now, now.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("GetByUserIDAndDueDateBetween() error = %v", err)
	}
	if len(today) != 1 {
		t.Fatalf("GetByUserIDAndDueDateBetween() len = %d, want 1", len(today))
	}

	updated, err := repo.UpdateStatusByIDAndUserID(ctx, created.ID, 1, domain.COMPLETED)
	if err != nil {
		t.Fatalf("UpdateStatusByIDAndUserID() error = %v", err)
	}
	if updated.Status != domain.COMPLETED {
		t.Fatalf("UpdateStatusByIDAndUserID() status = %v, want %v", updated.Status, domain.COMPLETED)
	}

	if err := repo.UpdateExpiredAndEnqueueSummary(ctx, []int64{created.ID}, domain.OutboxEvent{
		Topic:   "daily-summary",
		Key:     "1710842400",
		Payload: []byte(`{"window_start":1710841800,"window_end":1710842400,"users":[{"user_id":1,"completed":1,"not_completed":0}]}`),
	}); err != nil {
		t.Fatalf("UpdateExpiredAndEnqueueSummary() error = %v", err)
	}

	events, err := outboxRepo.ClaimPending(ctx, 10)
	if err != nil {
		t.Fatalf("ClaimPending() error = %v", err)
	}
	if len(events) != 1 || events[0].Topic != "daily-summary" {
		t.Fatalf("ClaimPending() = %+v", events)
	}
}
