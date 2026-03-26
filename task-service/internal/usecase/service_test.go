package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"task-tracker/task-service/internal/domain"
)

type stubTaskRepo struct {
	createFunc                     func(ctx context.Context, task domain.Task) (domain.Task, error)
	getByIDAndUserIDFunc           func(ctx context.Context, id, userID int64) (domain.Task, error)
	getByUserIDAndDueDateBetweenFn func(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error)
	getByDueDateBetweenFunc        func(ctx context.Context, from, to time.Time) ([]domain.Task, error)
	getByDueDateStatusNotFunc      func(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error)
	updateStatusByIDAndUserIDFunc  func(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error)
	updateStatusByIDsFunc          func(ctx context.Context, ids []int64, status domain.TaskStatus) error
}

func (r *stubTaskRepo) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	return r.createFunc(ctx, task)
}

func (r *stubTaskRepo) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	return domain.Task{}, nil
}

func (r *stubTaskRepo) GetByIDAndUserID(ctx context.Context, id, userID int64) (domain.Task, error) {
	return r.getByIDAndUserIDFunc(ctx, id, userID)
}

func (r *stubTaskRepo) GetByUserIDAndDueDateBetween(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) {
	return r.getByUserIDAndDueDateBetweenFn(ctx, userID, from, to)
}

func (r *stubTaskRepo) GetByDueDateBetween(ctx context.Context, from, to time.Time) ([]domain.Task, error) {
	return r.getByDueDateBetweenFunc(ctx, from, to)
}

func (r *stubTaskRepo) GetByDueDateBetweenAndStatusNot(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
	return r.getByDueDateStatusNotFunc(ctx, from, to, status)
}

func (r *stubTaskRepo) UpdateStatusByIDAndUserID(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
	return r.updateStatusByIDAndUserIDFunc(ctx, id, userID, status)
}

func (r *stubTaskRepo) UpdateStatusByIDs(ctx context.Context, ids []int64, status domain.TaskStatus) error {
	return r.updateStatusByIDsFunc(ctx, ids, status)
}

type stubTokenParser struct {
	parseUserIDFunc func(token string) (int64, error)
}

func (p stubTokenParser) ParseUserID(token string) (int64, error) {
	return p.parseUserIDFunc(token)
}

type stubEventPublisher struct {
	publishFunc func(ctx context.Context, summary ExpiredSummary) error
}

func (p stubEventPublisher) PublishExpiredSummary(ctx context.Context, summary ExpiredSummary) error {
	return p.publishFunc(ctx, summary)
}

func TestTaskServiceCreate(t *testing.T) {
	repo := &stubTaskRepo{
		createFunc: func(ctx context.Context, task domain.Task) (domain.Task, error) {
			if task.UserID != 42 || task.Status != domain.CREATED {
				t.Fatalf("unexpected task on create: %+v", task)
			}
			task.ID = 100
			return task, nil
		},
		getByIDAndUserIDFunc:           func(ctx context.Context, id, userID int64) (domain.Task, error) { return domain.Task{}, nil },
		getByUserIDAndDueDateBetweenFn: func(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) { return nil, nil },
		getByDueDateBetweenFunc:        func(ctx context.Context, from, to time.Time) ([]domain.Task, error) { return nil, nil },
		getByDueDateStatusNotFunc: func(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
			return nil, nil
		},
		updateStatusByIDAndUserIDFunc: func(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
			return domain.Task{}, nil
		},
		updateStatusByIDsFunc: func(ctx context.Context, ids []int64, status domain.TaskStatus) error { return nil },
	}

	svc := NewTaskService(
		repo,
		stubTokenParser{parseUserIDFunc: func(token string) (int64, error) { return 42, nil }},
		nil,
	)
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	task, err := svc.Create(context.Background(), "jwt", "ship feature", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.ID != 100 {
		t.Fatalf("Create() id = %d, want 100", task.ID)
	}
	if !task.CreatedAt.Equal(now) {
		t.Fatalf("Create() created_at = %v, want %v", task.CreatedAt, now)
	}
}

func TestTaskServiceGetToday(t *testing.T) {
	now := time.Date(2026, 3, 19, 10, 15, 0, 0, time.UTC)
	var gotFrom, gotTo time.Time

	repo := &stubTaskRepo{
		createFunc:           func(ctx context.Context, task domain.Task) (domain.Task, error) { return domain.Task{}, nil },
		getByIDAndUserIDFunc: func(ctx context.Context, id, userID int64) (domain.Task, error) { return domain.Task{}, nil },
		getByUserIDAndDueDateBetweenFn: func(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) {
			gotFrom, gotTo = from, to
			return []domain.Task{{ID: 1, UserID: userID}}, nil
		},
		getByDueDateBetweenFunc: func(ctx context.Context, from, to time.Time) ([]domain.Task, error) { return nil, nil },
		getByDueDateStatusNotFunc: func(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
			return nil, nil
		},
		updateStatusByIDAndUserIDFunc: func(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
			return domain.Task{}, nil
		},
		updateStatusByIDsFunc: func(ctx context.Context, ids []int64, status domain.TaskStatus) error { return nil },
	}

	svc := NewTaskService(
		repo,
		stubTokenParser{parseUserIDFunc: func(token string) (int64, error) { return 7, nil }},
		nil,
	)
	svc.now = func() time.Time { return now }

	tasks, err := svc.GetToday(context.Background(), "jwt")
	if err != nil {
		t.Fatalf("GetToday() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("GetToday() len = %d, want 1", len(tasks))
	}
	wantFrom := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	wantTo := wantFrom.Add(24 * time.Hour)
	if !gotFrom.Equal(wantFrom) || !gotTo.Equal(wantTo) {
		t.Fatalf("GetToday() range = [%v, %v), want [%v, %v)", gotFrom, gotTo, wantFrom, wantTo)
	}
}

func TestTaskServiceProcessRecentExpired(t *testing.T) {
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)
	var updatedIDs []int64
	var published ExpiredSummary

	repo := &stubTaskRepo{
		createFunc:                     func(ctx context.Context, task domain.Task) (domain.Task, error) { return domain.Task{}, nil },
		getByIDAndUserIDFunc:           func(ctx context.Context, id, userID int64) (domain.Task, error) { return domain.Task{}, nil },
		getByUserIDAndDueDateBetweenFn: func(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) { return nil, nil },
		getByDueDateBetweenFunc: func(ctx context.Context, from, to time.Time) ([]domain.Task, error) {
			return []domain.Task{
				{ID: 1, UserID: 10, Status: domain.CREATED},
				{ID: 2, UserID: 10, Status: domain.COMPLETED},
				{ID: 3, UserID: 20, Status: domain.AT_WORK},
			}, nil
		},
		getByDueDateStatusNotFunc: func(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
			return nil, nil
		},
		updateStatusByIDAndUserIDFunc: func(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
			return domain.Task{}, nil
		},
		updateStatusByIDsFunc: func(ctx context.Context, ids []int64, status domain.TaskStatus) error {
			updatedIDs = append([]int64(nil), ids...)
			if status != domain.EXPIRED {
				t.Fatalf("unexpected status = %v", status)
			}
			return nil
		},
	}

	svc := NewTaskService(repo, stubTokenParser{}, stubEventPublisher{
		publishFunc: func(ctx context.Context, summary ExpiredSummary) error {
			published = summary
			return nil
		},
	})
	svc.now = func() time.Time { return now }

	if err := svc.ProcessRecentExpired(context.Background()); err != nil {
		t.Fatalf("ProcessRecentExpired() error = %v", err)
	}
	if len(updatedIDs) != 2 || updatedIDs[0] != 1 || updatedIDs[1] != 3 {
		t.Fatalf("ProcessRecentExpired() updated IDs = %v, want [1 3]", updatedIDs)
	}
	if len(published.Users) != 2 {
		t.Fatalf("ProcessRecentExpired() summary users = %d, want 2", len(published.Users))
	}
}

func TestTaskServiceUpdateStatusRejectsInvalidToken(t *testing.T) {
	repo := &stubTaskRepo{
		createFunc:                     func(ctx context.Context, task domain.Task) (domain.Task, error) { return domain.Task{}, nil },
		getByIDAndUserIDFunc:           func(ctx context.Context, id, userID int64) (domain.Task, error) { return domain.Task{}, nil },
		getByUserIDAndDueDateBetweenFn: func(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) { return nil, nil },
		getByDueDateBetweenFunc:        func(ctx context.Context, from, to time.Time) ([]domain.Task, error) { return nil, nil },
		getByDueDateStatusNotFunc: func(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
			return nil, nil
		},
		updateStatusByIDAndUserIDFunc: func(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
			return domain.Task{}, nil
		},
		updateStatusByIDsFunc: func(ctx context.Context, ids []int64, status domain.TaskStatus) error { return nil },
	}

	svc := NewTaskService(
		repo,
		stubTokenParser{parseUserIDFunc: func(token string) (int64, error) { return 0, errors.New("bad token") }},
		nil,
	)

	_, err := svc.UpdateStatus(context.Background(), "bad", 1, domain.COMPLETED)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("UpdateStatus() error = %v, want %v", err, ErrInvalidToken)
	}
}
