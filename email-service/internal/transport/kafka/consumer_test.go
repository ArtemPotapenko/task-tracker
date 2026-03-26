package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"task-tracker/email-service/internal/usecase"
)

type stubReader struct {
	messages   []Message
	fetchIndex int
	committed  int
}

func (r *stubReader) FetchMessage(ctx context.Context) (Message, error) {
	if r.fetchIndex >= len(r.messages) {
		return Message{}, context.Canceled
	}
	msg := r.messages[r.fetchIndex]
	r.fetchIndex++
	return msg, nil
}

func (r *stubReader) CommitMessages(ctx context.Context, msg Message) error {
	r.committed++
	return nil
}

type stubUsersClient struct {
	getUsersByIDsFunc func(ctx context.Context, ids []int64) (map[int64]string, error)
}

func (c stubUsersClient) GetUsersByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	return c.getUsersByIDsFunc(ctx, ids)
}

type recordMailer struct {
	sent []string
}

func (m *recordMailer) Send(to, subject, body string) error {
	m.sent = append(m.sent, to)
	return nil
}

type allowAllDedupe struct{}

func (allowAllDedupe) Once(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return true, nil
}

func TestConsumerConsumeRegister(t *testing.T) {
	payload, _ := json.Marshal(RegisterMessage{Email: "user@example.com"})
	reader := &stubReader{messages: []Message{{Value: payload}}}
	mailer := &recordMailer{}
	service := usecase.NewService(mailer, allowAllDedupe{}, time.Hour)
	consumer := NewConsumer(service)

	errCh := make(chan error, 1)
	consumer.ConsumeRegister(context.Background(), reader, errCh)

	select {
	case err := <-errCh:
		t.Fatalf("unexpected errCh value: %v", err)
	default:
	}
	if len(mailer.sent) != 1 || mailer.sent[0] != "user@example.com" {
		t.Fatalf("unexpected sent emails = %v", mailer.sent)
	}
	if reader.committed != 1 {
		t.Fatalf("committed = %d, want 1", reader.committed)
	}
}

func TestConsumerConsumeDaily(t *testing.T) {
	payload, _ := json.Marshal(DailySummaryMessage{
		Date: "2026-03-19",
		Users: []DailySummaryUser{
			{UserID: 1, Completed: 2, NotCompleted: 1},
			{UserID: 2, Completed: 1, NotCompleted: 4},
		},
	})
	reader := &stubReader{messages: []Message{{Value: payload}}}
	mailer := &recordMailer{}
	service := usecase.NewService(mailer, allowAllDedupe{}, time.Hour)
	consumer := NewConsumer(service)

	errCh := make(chan error, 1)
	consumer.ConsumeDaily(context.Background(), reader, stubUsersClient{
		getUsersByIDsFunc: func(ctx context.Context, ids []int64) (map[int64]string, error) {
			return map[int64]string{
				1: "one@example.com",
				2: "two@example.com",
			}, nil
		},
	}, errCh)

	select {
	case err := <-errCh:
		t.Fatalf("unexpected errCh value: %v", err)
	default:
	}
	if len(mailer.sent) != 2 {
		t.Fatalf("sent len = %d, want 2", len(mailer.sent))
	}
	if reader.committed != 1 {
		t.Fatalf("committed = %d, want 1", reader.committed)
	}
}
