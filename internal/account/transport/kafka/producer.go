package kafka

import (
	"context"
	"encoding/json"

	kafka "github.com/segmentio/kafka-go"
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
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{Value: data})
}
