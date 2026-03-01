package config

import (
	"time"

	"task-tracker/pkg/env"
)

type Config struct {
	KafkaBroker       string
	RegisterTopic     string
	DailySummaryTopic string
	GroupID           string
	AccountGRPCAddr   string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	DedupeTTL         time.Duration

	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	SMTPFrom   string
	SMTPUseTLS bool
	Timeout    time.Duration
}

func Load() (Config, error) {
	useTLS, err := env.GetEnvAsInt("SMTP_USE_TLS", 0)
	if err != nil {
		return Config{}, err
	}
	timeout, err := env.GetEnvAsDuration("SMTP_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	dedupeTTL, err := env.GetEnvAsDuration("EMAIL_DEDUPE_TTL", 48*time.Hour)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := env.GetEnvAsInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		KafkaBroker:       env.GetEnvOrDefault("KAFKA_BROKER", "localhost:9092"),
		RegisterTopic:     env.GetEnvOrDefault("KAFKA_REGISTER_TOPIC", "register"),
		DailySummaryTopic: env.GetEnvOrDefault("KAFKA_DAILY_TOPIC", "task-daily-summary"),
		GroupID:           env.GetEnvOrDefault("KAFKA_GROUP_ID", "email-sender"),
		AccountGRPCAddr:   env.GetEnvOrDefault("ACCOUNT_GRPC_ADDR", "localhost:50051"),
		RedisAddr:         env.GetEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     env.GetEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:           redisDB,
		DedupeTTL:         dedupeTTL,
		SMTPHost:          env.GetEnvOrDefault("SMTP_HOST", ""),
		SMTPPort:          env.GetEnvOrDefault("SMTP_PORT", ""),
		SMTPUser:          env.GetEnvOrDefault("SMTP_USER", ""),
		SMTPPass:          env.GetEnvOrDefault("SMTP_PASS", ""),
		SMTPFrom:          env.GetEnvOrDefault("SMTP_FROM", ""),
		SMTPUseTLS:        useTLS != 0,
		Timeout:           timeout,
	}
	return cfg, nil
}
