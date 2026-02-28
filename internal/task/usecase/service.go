package usecase

import (
	"context"
	"errors"
	"time"

	"task-tracker/internal/task/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrInvalidInput = errors.New("invalid input")
)

type TokenParser interface {
	ParseUserID(token string) (int64, error)
}

type TaskService struct {
	repo   domain.TaskRepository
	tokens TokenParser
	events TaskEventPublisher
	now    func() time.Time
}

func NewTaskService(repo domain.TaskRepository, tokens TokenParser, events TaskEventPublisher) TaskService {
	return TaskService{repo: repo, tokens: tokens, events: events, now: time.Now}
}

func (s TaskService) Create(ctx context.Context, token, description string, dueDate time.Time) (domain.Task, error) {
	if description == "" || dueDate.IsZero() {
		return domain.Task{}, ErrInvalidInput
	}

	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		return domain.Task{}, ErrInvalidToken
	}

	task := domain.Task{
		UserID:      userID,
		Description: description,
		Status:      domain.CREATED,
		CreatedAt:   s.now(),
		DueDate:     dueDate,
	}
	return s.repo.Create(ctx, task)
}

func (s TaskService) GetByID(ctx context.Context, token string, id int64) (domain.Task, error) {
	if id <= 0 {
		return domain.Task{}, ErrInvalidInput
	}

	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		return domain.Task{}, ErrInvalidToken
	}

	return s.repo.GetByIDAndUserID(ctx, id, userID)
}

func (s TaskService) GetToday(ctx context.Context, token string) ([]domain.Task, error) {
	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	now := s.now()
	loc := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	return s.repo.GetByUserIDAndDueDateBetween(ctx, userID, start, end)
}

func (s TaskService) GetRecentExpired(ctx context.Context) ([]domain.Task, error) {
	now := s.now()
	from := now.Add(-10 * time.Minute)

	return s.repo.GetByDueDateBetweenAndStatusNot(ctx, from, now, domain.COMPLETED)
}

func (s TaskService) ProcessRecentExpired(ctx context.Context) error {
	now := s.now()
	from := now.Add(-10 * time.Minute)

	tasks, err := s.repo.GetByDueDateBetween(ctx, from, now)
	if err != nil {
		return err
	}

	summary := ExpiredSummary{
		WindowStart: from,
		WindowEnd:   now,
		Users:       make([]UserExpiredSummary, 0),
	}

	counts := make(map[int64]*UserExpiredSummary)
	var toExpire []int64
	for _, task := range tasks {
		stats, ok := counts[task.UserID]
		if !ok {
			stats = &UserExpiredSummary{UserID: task.UserID}
			counts[task.UserID] = stats
		}

		if task.Status == domain.COMPLETED {
			stats.Completed++
		} else {
			stats.NotCompleted++
			toExpire = append(toExpire, task.ID)
		}
	}

	for _, stats := range counts {
		summary.Users = append(summary.Users, *stats)
	}

	if err := s.repo.UpdateStatusByIDs(ctx, toExpire, domain.EXPIRED); err != nil {
		return err
	}

	if s.events == nil {
		return nil
	}
	return s.events.PublishExpiredSummary(ctx, summary)
}

func (s TaskService) UpdateStatus(ctx context.Context, token string, id int64, status domain.TaskStatus) (domain.Task, error) {
	if id <= 0 {
		return domain.Task{}, ErrInvalidInput
	}
	if status != domain.CREATED && status != domain.AT_WORK && status != domain.COMPLETED {
		return domain.Task{}, ErrInvalidInput
	}

	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		return domain.Task{}, ErrInvalidToken
	}

	return s.repo.UpdateStatusByIDAndUserID(ctx, id, userID, status)
}
