package kafka

import (
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
)

func NewReader(broker string, topic string, groupID string) (*kafka.Reader, error) {
	broker = strings.TrimSpace(broker)
	if broker == "" {
		return nil, ErrEmptyBroker
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errors.New("kafka topic is empty")
	}
	if strings.TrimSpace(groupID) == "" {
		return nil, errors.New("kafka group id is empty")
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	}), nil
}
