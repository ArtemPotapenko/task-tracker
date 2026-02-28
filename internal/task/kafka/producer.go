package kafka

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/segmentio/kafka-go"

	"task-tracker/internal/task/usecase"
)

type ExpiredSummaryMessage struct {
	WindowStart int64                `json:"window_start"`
	WindowEnd   int64                `json:"window_end"`
	Users       []UserSummaryMessage `json:"users"`
}

type UserSummaryMessage struct {
	UserID       int64 `json:"user_id"`
	Completed    int   `json:"completed"`
	NotCompleted int   `json:"not_completed"`
}

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(writer *kafka.Writer) *Publisher {
	return &Publisher{writer: writer}
}

func (p *Publisher) PublishExpiredSummary(ctx context.Context, summary usecase.ExpiredSummary) error {
	users := make([]UserSummaryMessage, 0, len(summary.Users))
	for _, user := range summary.Users {
		users = append(users, UserSummaryMessage{
			UserID:       user.UserID,
			Completed:    user.Completed,
			NotCompleted: user.NotCompleted,
		})
	}

	payload := ExpiredSummaryMessage{
		WindowStart: summary.WindowStart.Unix(),
		WindowEnd:   summary.WindowEnd.Unix(),
		Users:       users,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(payload.WindowEnd, 10)),
		Value: data,
	})
}
