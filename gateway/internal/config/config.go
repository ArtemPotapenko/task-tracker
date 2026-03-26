package config

import (
	"time"

	"task-tracker/shared-libs/pkg/env"
)

type Config struct {
	HTTPAddr        string
	AccountGRPCAddr string
	TaskGRPCAddr    string
	OpenAPIDir      string
	RateLimitRPS    int
	RateLimitBurst  int
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	timeout, err := env.GetEnvAsDuration("GATEWAY_SHUTDOWN_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	rateLimitRPS, err := env.GetEnvAsInt("GATEWAY_RATE_LIMIT_RPS", 20)
	if err != nil {
		return Config{}, err
	}
	rateLimitBurst, err := env.GetEnvAsInt("GATEWAY_RATE_LIMIT_BURST", 40)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:        env.GetEnvOrDefault("GATEWAY_HTTP_ADDR", ":8080"),
		AccountGRPCAddr: env.GetEnvOrDefault("ACCOUNT_GRPC_ADDR", ":50051"),
		TaskGRPCAddr:    env.GetEnvOrDefault("TASK_GRPC_ADDR", ":50052"),
		OpenAPIDir:      env.GetEnvOrDefault("GATEWAY_OPENAPI_DIR", "proto-lib/gen/public/openapi"),
		RateLimitRPS:    rateLimitRPS,
		RateLimitBurst:  rateLimitBurst,
		ShutdownTimeout: timeout,
	}
	return cfg, nil
}
