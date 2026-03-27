package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"

	"task-tracker/shared-libs/pkg/logger"
	"task-tracker/task-service/internal/domain"
)

type messageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type OutboxPublisher struct {
	writer       messageWriter
	outbox       domain.OutboxRepository
	pollInterval time.Duration
	batchSize    int
}

func NewOutboxPublisher(writer messageWriter, outbox domain.OutboxRepository, pollInterval time.Duration, batchSize int) *OutboxPublisher {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &OutboxPublisher{writer: writer, outbox: outbox, pollInterval: pollInterval, batchSize: batchSize}
}

func (p *OutboxPublisher) Run(ctx context.Context) error {
	if err := p.publishBatch(ctx); err != nil {
		logger.Log.Infof("kafka outbox publish: batch error err=%v", err)
	}

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.publishBatch(ctx); err != nil {
				logger.Log.Infof("kafka outbox publish: batch error err=%v", err)
			}
		}
	}
}

func (p *OutboxPublisher) publishBatch(ctx context.Context) error {
	for i := 0; i < p.batchSize; i++ {
		events, err := p.outbox.ClaimPending(ctx, 1)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		event := events[0]
		msg := kafka.Message{Topic: event.Topic, Value: event.Payload}
		if event.Key != "" {
			msg.Key = []byte(event.Key)
		}

		if err := p.writer.WriteMessages(ctx, msg); err != nil {
			logger.Log.Infof("kafka outbox publish: write error id=%d topic=%s err=%v", event.ID, event.Topic, err)
			if releaseErr := p.outbox.Release(ctx, event.ID); releaseErr != nil {
				logger.Log.Infof("kafka outbox publish: release error id=%d err=%v", event.ID, releaseErr)
			}
			return err
		}
		if err := p.outbox.MarkProcessed(ctx, event.ID); err != nil {
			logger.Log.Infof("kafka outbox publish: mark processed error id=%d err=%v", event.ID, err)
			return err
		}
		logger.Log.Infof("kafka outbox publish: success id=%d topic=%s", event.ID, event.Topic)
	}
	return nil
}
