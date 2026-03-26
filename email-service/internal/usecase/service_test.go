package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type stubMailer struct {
	sendFunc func(to, subject, body string) error
}

func (m stubMailer) Send(to, subject, body string) error {
	return m.sendFunc(to, subject, body)
}

type stubDedupe struct {
	onceFunc func(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

func (d stubDedupe) Once(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return d.onceFunc(ctx, key, ttl)
}

func TestServiceSendWelcome(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sent := false
		svc := NewService(
			stubMailer{sendFunc: func(to, subject, body string) error {
				sent = true
				if to != "User@Example.com" {
					t.Fatalf("unexpected to = %q", to)
				}
				if subject == "" || body == "" {
					t.Fatalf("expected non-empty subject/body")
				}
				return nil
			}},
			stubDedupe{onceFunc: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				if key != "user@example.com" {
					t.Fatalf("unexpected dedupe key = %q", key)
				}
				if ttl != time.Hour {
					t.Fatalf("unexpected ttl = %s", ttl)
				}
				return true, nil
			}},
			time.Hour,
		)

		if err := svc.SendWelcome(context.Background(), "User@Example.com"); err != nil {
			t.Fatalf("SendWelcome() error = %v", err)
		}
		if !sent {
			t.Fatalf("expected message to be sent")
		}
	})

	t.Run("dedupe blocks", func(t *testing.T) {
		svc := NewService(
			stubMailer{sendFunc: func(to, subject, body string) error {
				t.Fatalf("mailer should not be called")
				return nil
			}},
			stubDedupe{onceFunc: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				return false, nil
			}},
			time.Hour,
		)

		if err := svc.SendWelcome(context.Background(), "user@example.com"); err != nil {
			t.Fatalf("SendWelcome() error = %v, want nil", err)
		}
	})
}

func TestServiceSendDailySummary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var bodyText string
		svc := NewService(
			stubMailer{sendFunc: func(to, subject, body string) error {
				bodyText = body
				return nil
			}},
			stubDedupe{onceFunc: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				if key != "dailyExpired:2026-03-19:7" {
					t.Fatalf("unexpected dedupe key = %q", key)
				}
				return true, nil
			}},
			time.Hour,
		)

		err := svc.SendDailySummary(context.Background(), "user@example.com", 7, 3, 2, "2026-03-19")
		if err != nil {
			t.Fatalf("SendDailySummary() error = %v", err)
		}
		if !strings.Contains(bodyText, "Выполнено: 3") || !strings.Contains(bodyText, "Не выполнено: 2") {
			t.Fatalf("unexpected body = %q", bodyText)
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc := NewService(
			stubMailer{sendFunc: func(to, subject, body string) error { return nil }},
			stubDedupe{onceFunc: func(ctx context.Context, key string, ttl time.Duration) (bool, error) { return true, nil }},
			time.Hour,
		)

		err := svc.SendDailySummary(context.Background(), "user@example.com", 0, 1, 1, "")
		if err == nil || err.Error() != "invalid user id" {
			t.Fatalf("SendDailySummary() error = %v, want invalid user id", err)
		}
	})

	t.Run("dedupe error", func(t *testing.T) {
		wantErr := errors.New("redis down")
		svc := NewService(
			stubMailer{sendFunc: func(to, subject, body string) error {
				t.Fatalf("mailer should not be called")
				return nil
			}},
			stubDedupe{onceFunc: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				return false, wantErr
			}},
			time.Hour,
		)

		err := svc.SendDailySummary(context.Background(), "user@example.com", 1, 1, 1, "")
		if !errors.Is(err, wantErr) {
			t.Fatalf("SendDailySummary() error = %v, want %v", err, wantErr)
		}
	})
}
