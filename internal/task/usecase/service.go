package usecase

import (
	"context"
	"errors"
	"time"

	"task-tracker/internal/task/domain"
	"task-tracker/pkg/logger"
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

func NewTaskService(repo domain.TaskRepository, tokens TokenParser, events TaskEventPublisher) *TaskService {
	return &TaskService{repo: repo, tokens: tokens, events: events, now: time.Now}
}

func (s *TaskService) Create(ctx context.Context, token, description string, dueDate time.Time) (domain.Task, error) {
	if description == "" || dueDate.IsZero() {
		logger.Log.Infof("task create: invalid input")
		return domain.Task{}, ErrInvalidInput
	}

	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		logger.Log.Infof("task create: invalid token err=%v", err)
		return domain.Task{}, ErrInvalidToken
	}

	task := domain.Task{
		UserID:      userID,
		Description: description,
		Status:      domain.CREATED,
		CreatedAt:   s.now(),
		DueDate:     dueDate,
	}
	created, err := s.repo.Create(ctx, task)
	if err != nil {
		logger.Log.Infof("task create: repo error user_id=%d err=%v", userID, err)
		return domain.Task{}, err
	}
	logger.Log.Infof("task create: success id=%d user_id=%d", created.ID, userID)
	return created, nil
}

func (s *TaskService) GetByID(ctx context.Context, token string, id int64) (domain.Task, error) {
	if id <= 0 {
		logger.Log.Infof("task get by id: invalid id=%d", id)
		return domain.Task{}, ErrInvalidInput
	}

	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		logger.Log.Infof("task get by id: invalid token err=%v", err)
		return domain.Task{}, ErrInvalidToken
	}

	task, err := s.repo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		logger.Log.Infof("task get by id: repo error id=%d user_id=%d err=%v", id, userID, err)
		return domain.Task{}, err
	}
	logger.Log.Infof("task get by id: success id=%d user_id=%d", id, userID)
	return task, nil
}

func (s *TaskService) GetToday(ctx context.Context, token string) ([]domain.Task, error) {
	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		logger.Log.Infof("task get today: invalid token err=%v", err)
		return nil, ErrInvalidToken
	}

	now := s.now()
	loc := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	tasks, err := s.repo.GetByUserIDAndDueDateBetween(ctx, userID, start, end)
	if err != nil {
		logger.Log.Infof("task get today: repo error user_id=%d err=%v", userID, err)
		return nil, err
	}
	logger.Log.Infof("task get today: success user_id=%d count=%d", userID, len(tasks))
	return tasks, nil
}

func (s *TaskService) GetRecentExpired(ctx context.Context) ([]domain.Task, error) {
	now := s.now()
	from := now.Add(-10 * time.Minute)

	tasks, err := s.repo.GetByDueDateBetweenAndStatusNot(ctx, from, now, domain.COMPLETED)
	if err != nil {
		logger.Log.Infof("task get recent expired: repo error err=%v", err)
		return nil, err
	}
	logger.Log.Infof("task get recent expired: success count=%d", len(tasks))
	return tasks, nil
}

func (s *TaskService) ProcessRecentExpired(ctx context.Context) error {
	now := s.now()
	from := now.Add(-10 * time.Minute)

	tasks, err := s.repo.GetByDueDateBetween(ctx, from, now)
	if err != nil {
		logger.Log.Infof("task process recent expired: repo error err=%v", err)
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
		logger.Log.Infof("task process recent expired: update status error count=%d err=%v", len(toExpire), err)
		return err
	}

	if s.events == nil {
		logger.Log.Infof("task process recent expired: no publisher count=%d", len(summary.Users))
		return nil
	}
	if err := s.events.PublishExpiredSummary(ctx, summary); err != nil {
		logger.Log.Infof("task process recent expired: publish error err=%v", err)
		return err
	}
	logger.Log.Infof("task process recent expired: success users=%d", len(summary.Users))
	return nil
}

func (s *TaskService) UpdateStatus(ctx context.Context, token string, id int64, status domain.TaskStatus) (domain.Task, error) {
	if id <= 0 {
		logger.Log.Infof("task update status: invalid id=%d", id)
		return domain.Task{}, ErrInvalidInput
	}
	if status != domain.CREATED && status != domain.AT_WORK && status != domain.COMPLETED {
		logger.Log.Infof("task update status: invalid status=%v", status)
		return domain.Task{}, ErrInvalidInput
	}

	userID, err := s.tokens.ParseUserID(token)
	if err != nil {
		logger.Log.Infof("task update status: invalid token err=%v", err)
		return domain.Task{}, ErrInvalidToken
	}

	task, err := s.repo.UpdateStatusByIDAndUserID(ctx, id, userID, status)
	if err != nil {
		logger.Log.Infof("task update status: repo error id=%d user_id=%d err=%v", id, userID, err)
		return domain.Task{}, err
	}
	logger.Log.Infof("task update status: success id=%d user_id=%d status=%v", id, userID, status)
	return task, nil
}
