package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"task-tracker/internal/email/usecase"
	"task-tracker/pkg/logger"
)

type Message struct {
	Value []byte
}

type MessageReader interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msg Message) error
}

type UsersClient interface {
	GetUsersByIDs(ctx context.Context, ids []int64) (map[int64]string, error)
}

type Consumer struct {
	service *usecase.Service
}

func NewConsumer(service *usecase.Service) Consumer {
	return Consumer{service: service}
}

func (c *Consumer) ConsumeRegister(ctx context.Context, reader MessageReader, errCh chan<- error) {
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			errCh <- err
			return
		}
		logger.Log.Infof("kafka register: message received")

		var payload usecase.RegisterMessage
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			logger.Log.Infof("kafka register: invalid payload err=%v", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}
		if err := c.service.SendWelcome(ctx, payload); err != nil {
			logger.Log.Infof("send welcome: %v", err)
		}
		_ = reader.CommitMessages(ctx, msg)
	}
}

func (c *Consumer) ConsumeDaily(ctx context.Context, reader MessageReader, users UsersClient, errCh chan<- error) {
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			errCh <- err
			return
		}

		var payload usecase.DailySummaryMessage
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			logger.Log.Infof("kafka daily: invalid payload err=%v", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}
		if len(payload.Users) == 0 {
			logger.Log.Infof("kafka daily: empty users")
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		ids := make([]int64, 0, len(payload.Users))
		for _, user := range payload.Users {
			if user.UserID > 0 {
				ids = append(ids, user.UserID)
			}
		}
		if len(ids) == 0 {
			logger.Log.Infof("kafka daily: no valid user ids")
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		emailByID, err := users.GetUsersByIDs(ctx, ids)
		if err != nil {
			logger.Log.Infof("get users by ids: %v", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		for _, user := range payload.Users {
			email := emailByID[user.UserID]
			if email == "" {
				logger.Log.Infof("kafka daily: missing email user_id=%d", user.UserID)
				continue
			}
			if err := c.service.SendDailySummary(ctx, email, user.UserID, user.Completed, user.NotCompleted, payload.Date); err != nil {
				logger.Log.Infof("send daily summary: %v", err)
			}
		}
		_ = reader.CommitMessages(ctx, msg)
	}
}
