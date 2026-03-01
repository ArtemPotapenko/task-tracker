package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrEmptyRedisAddr = errors.New("redis address is empty")

type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Close() error
}

type Client struct {
	client *redis.Client
}

func NewClient(addr, password string, db int, pingTimeout time.Duration) (*Client, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, ErrEmptyRedisAddr
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return c.client.SetNX(ctx, key, value, expiration)
}

func (c *Client) Close() error {
	return c.client.Close()
}
