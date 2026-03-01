package config

import "task-tracker/pkg/env"

type Config struct {
	GRPCAddr    string
	DBDriver    string
	DBDSN       string
	JWTSecret   string
	KafkaBroker string
	KafkaTopic  string
}

func Load() (Config, error) {
	cfg := Config{
		GRPCAddr:    env.GetEnvOrDefault("GRPC_ADDR", ":50052"),
		DBDriver:    env.GetEnvOrDefault("DB_DRIVER", "pgx"),
		DBDSN:       env.GetEnvOrDefault("DB_DSN", "pgsql:host=localhost port=5433 dbname=testdb user=admin password=secret"),
		JWTSecret:   env.GetEnvOrDefault("JWT_SECRET", "secret"),
		KafkaBroker: env.GetEnvOrDefault("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  env.GetEnvOrDefault("KAFKA_TOPIC", "task-expired-summary"),
	}
	return cfg, nil
}
