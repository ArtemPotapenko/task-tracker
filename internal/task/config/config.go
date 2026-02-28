package config

import (
	"strings"

	"task-tracker/pkg/env"
)

type Config struct {
	GRPCAddr     string
	DBDriver     string
	DBDSN        string
	JWTSecret    string
	KafkaBrokers []string
	KafkaTopic   string
}

func Load() (Config, error) {
	brokers := splitAndTrim(env.GetEnvOrDefault("KAFKA_BROKERS", "localhost:9092"))
	cfg := Config{
		GRPCAddr:     env.GetEnvOrDefault("GRPC_ADDR", ":50052"),
		DBDriver:     env.GetEnvOrDefault("DB_DRIVER", "pgx"),
		DBDSN:        env.GetEnvOrDefault("DB_DSN", "pgsql:host=localhost port=5433 dbname=testdb user=admin password=secret"),
		JWTSecret:    env.GetEnvOrDefault("JWT_SECRET", "secret"),
		KafkaBrokers: brokers,
		KafkaTopic:   env.GetEnvOrDefault("KAFKA_TOPIC", "task-expired-summary"),
	}
	return cfg, nil
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
