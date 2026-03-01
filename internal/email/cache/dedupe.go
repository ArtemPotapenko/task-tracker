package cache

import (
	"context"
	"errors"
	"time"
)

var ErrEmptyKey = errors.New("empty key")

type RedisDedupe struct {
	client RedisClient
}

type RedisClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
}

func NewRedisDedupe(client RedisClient) RedisDedupe {
	return RedisDedupe{client: client}
}

func (r *RedisDedupe) Once(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, ErrEmptyKey
	}
	return r.client.SetNX(ctx, key, "1", ttl)
}
