package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
		return errors.New("empty email")
	}
	if ok, err := s.allow(ctx, keyRegister(msg.Email)); err != nil || !ok {
		return err
	}

	subject := "Добро пожаловать в Task Tracker"
	body := "Здравствуйте! Ваш аккаунт успешно создан."
	return s.mailer.Send(msg.Email, subject, body)
}

func (s *Service) SendDailySummary(ctx context.Context, email string, userID int64, completed, notCompleted int, date string) error {
	if email == "" {
		return errors.New("empty email")
	}
	if userID <= 0 {
		return errors.New("invalid user id")
	}
	if ok, err := s.allow(ctx, keyDaily(date, userID)); err != nil || !ok {
		return err
	}

	dateLine := date
	if dateLine == "" {
		dateLine = "сегодня"
	}
	subject := "Ежедневный отчет по задачам"
	body := fmt.Sprintf("Ваш отчет за %s:\nВыполнено: %d\nНе выполнено: %d", dateLine, completed, notCompleted)
	return s.mailer.Send(email, subject, body)
}

func (s *Service) allow(ctx context.Context, key string) (bool, error) {
	if s.dedupe == nil {
		return true, nil
	}
	return s.dedupe.Once(ctx, key, s.dedupeTTL)
}

func keyRegister(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func keyDaily(date string, userID int64) string {
	return fmt.Sprintf("%d", userID)
}
