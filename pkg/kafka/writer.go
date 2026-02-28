package kafka

import (
	"errors"

	kafka "github.com/segmentio/kafka-go"
)

var ErrEmptyBrokers = errors.New("kafka brokers list is empty")

func NewWriter(brokers []string, topic string) (*kafka.Writer, error) {
	if len(brokers) == 0 {
		return nil, ErrEmptyBrokers
	}

	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}, nil
}
