package kafka

import (
	"context"
	"encoding/json"

	kafka "github.com/segmentio/kafka-go"
	"task-tracker/pkg/logger"
)

type RegisterMessage struct {
	Email string `json:"email"`
}

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(writer *kafka.Writer) *Publisher {
	return &Publisher{writer: writer}
}

func (p *Publisher) PublishRegistered(ctx context.Context, email string) error {
	payload := RegisterMessage{Email: email}
	data, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Infof("kafka publish registered: marshal error email=%s err=%v", email, err)
		return err
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{Value: data}); err != nil {
		logger.Log.Infof("kafka publish registered: write error email=%s err=%v", email, err)
		return err
	}
	logger.Log.Infof("kafka publish registered: success email=%s", email)
	return nil
}
