package repo

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/task/domain"
	"task-tracker/pkg/testutil"
)

func TestTaskRepositoryIntegration(t *testing.T) {
	db := testutil.StartPostgres(t, "taskdb")
	testutil.ApplyGooseUpSQL(t, db, "../../../migrations/task/00001_create_tasks.sql")

	if _, err := db.Exec("TRUNCATE TABLE tasks RESTART IDENTITY"); err != nil {
		t.Fatalf("TRUNCATE tasks error = %v", err)
	}

	repo := NewTaskRepository(db)
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
}
