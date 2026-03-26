package kafka

import (
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
)

var ErrEmptyBroker = errors.New("kafka broker is empty")

func NewWriter(broker string, topic string) (*kafka.Writer, error) {
	broker = strings.TrimSpace(broker)
	if broker == "" {
		return nil, ErrEmptyBroker
	}

	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}, nil
}
