package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"task-tracker/pkg/logger"
	"time"
)

type Mailer interface {
	Send(to, subject, body string) error
}

type DedupeStore interface {
	Once(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type Service struct {
	mailer    Mailer
	dedupe    DedupeStore
	dedupeTTL time.Duration
}

func NewService(mailer Mailer, dedupe DedupeStore, dedupeTTL time.Duration) *Service {
	return &Service{mailer: mailer, dedupe: dedupe, dedupeTTL: dedupeTTL}
}

type RegisterMessage struct {
	Email string `json:"email"`
}

type DailySummaryUser struct {
	UserID       int64 `json:"user_id"`
	Completed    int   `json:"completed"`
	NotCompleted int   `json:"not_completed"`
}

type DailySummaryMessage struct {
	Date  string             `json:"date"`
	Users []DailySummaryUser `json:"users"`
}

func (s *Service) SendWelcome(ctx context.Context, msg RegisterMessage) error {
	if msg.Email == "" {
		logger.Log.Infof("email send welcome: empty email")
		return errors.New("empty email")
	}
	if ok, err := s.allow(ctx, keyRegister(msg.Email)); err != nil || !ok {
		if err != nil {
			logger.Log.Infof("email send welcome: dedupe error email=%s err=%v", msg.Email, err)
		}
		return err
	}

	subject := "Добро пожаловать в Task Tracker"
	body := "Здравствуйте! Ваш аккаунт успешно создан."
	if err := s.mailer.Send(msg.Email, subject, body); err != nil {
		logger.Log.Infof("email send welcome: send error email=%s err=%v", msg.Email, err)
		return err
	}
	logger.Log.Infof("email send welcome: success email=%s", msg.Email)
	return nil
}

func (s *Service) SendDailySummary(ctx context.Context, email string, userID int64, completed, notCompleted int, date string) error {
	if email == "" {
		logger.Log.Infof("email send daily: empty email user_id=%d", userID)
		return errors.New("empty email")
	}
	if userID <= 0 {
		logger.Log.Infof("email send daily: invalid user id=%d", userID)
		return errors.New("invalid user id")
	}
	if ok, err := s.allow(ctx, keyDaily(date, userID)); err != nil || !ok {
		if err != nil {
			logger.Log.Infof("email send daily: dedupe error user_id=%d err=%v", userID, err)
		}
		return err
	}

	dateLine := date
	if dateLine == "" {
		dateLine = "сегодня"
	}
	subject := "Ежедневный отчет по задачам"
	body := fmt.Sprintf("Ваш отчет за %s:\nВыполнено: %d\nНе выполнено: %d", dateLine, completed, notCompleted)
	if err := s.mailer.Send(email, subject, body); err != nil {
		logger.Log.Infof("email send daily: send error user_id=%d email=%s err=%v", userID, email, err)
		return err
	}
	logger.Log.Infof("email send daily: success user_id=%d email=%s", userID, email)
	return nil
}

func (s *Service) allow(ctx context.Context, key string) (bool, error) {
	if s.dedupe == nil {
		return true, nil
	}
	ok, err := s.dedupe.Once(ctx, key, s.dedupeTTL)
	if err != nil {
		logger.Log.Infof("email dedupe: error key=%s err=%v", key, err)
		return false, err
	}
	if !ok {
		logger.Log.Infof("email dedupe: blocked key=%s", key)
	}
	return ok, nil
}

func keyRegister(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func keyDaily(date string, userID int64) string {
	return fmt.Sprintf("%d", userID)
}
