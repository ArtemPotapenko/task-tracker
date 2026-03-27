package repo

import (
	"context"
	"testing"

	"task-tracker/account-service/internal/domain"
	"task-tracker/shared-libs/pkg/testutil"
)

func TestUserRepositoryIntegration(t *testing.T) {
	db := testutil.StartPostgres(t, "accountdb")
	testutil.ApplyUpSQL(t, db, "../../migrations/00001_create_users.sql")

	if _, err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY"); err != nil {
		t.Fatalf("TRUNCATE users error = %v", err)
	}

	repo := NewUserRepository(db)
	outboxRepo := NewOutboxRepository(db)
	ctx := context.Background()

	created, err := repo.CreateWithRegisteredEvent(ctx, domain.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("Create() id = 0, want non-zero")
	}

	got, err := repo.GetByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if got.Email != "user@example.com" || got.PasswordHash != "hash" {
		t.Fatalf("GetByEmail() = %+v", got)
	}

	users, err := repo.GetByIDs(ctx, []int64{created.ID})
	if err != nil {
		t.Fatalf("GetByIDs() error = %v", err)
	}
	if len(users) != 1 || users[0].ID != created.ID {
		t.Fatalf("GetByIDs() = %+v", users)
	}

	events, err := outboxRepo.ClaimPending(ctx, 10)
	if err != nil {
		t.Fatalf("ClaimPending() error = %v", err)
	}
	if len(events) != 1 || events[0].Topic != "register" {
		t.Fatalf("ClaimPending() = %+v", events)
	}
}
