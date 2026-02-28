package config

import (
	"time"

	"task-tracker/pkg/env"
)

type Config struct {
	GRPCAddr   string
	DBDriver   string
	DBDSN      string
	JWTSecret  string
	JWTTTL     time.Duration
	BcryptCost int
}

func Load() (Config, error) {
	jwtTTL, err := env.GetEnvAsDuration("JWT_TTL", 12*time.Hour)
	if err != nil {
		return Config{}, err
	}

	bcryptCost, err := env.GetEnvAsInt("BCRYPT_COST", 0)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		GRPCAddr:   env.GetEnvOrDefault("GRPC_ADDR", ":50051"),
		DBDriver:   env.GetEnvOrDefault("DB_DRIVER", "pgx"),
		DBDSN:      env.GetEnvOrDefault("DB_DSN", "pgsql:host=localhost port=5433 dbname=testdb user=admin password=secret"),
		JWTSecret:  env.GetEnvOrDefault("JWT_SECRET", "secret"),
		JWTTTL:     jwtTTL,
		BcryptCost: bcryptCost,
	}
	return cfg, nil
}
