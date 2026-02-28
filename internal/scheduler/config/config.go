package config

import (
	"time"

	"task-tracker/pkg/env"
)

type Config struct {
	TaskGRPCAddr string
	Interval     time.Duration
	CallTimeout  time.Duration
}

func Load() (Config, error) {
	interval, err := env.GetEnvAsDuration("SCHEDULER_INTERVAL", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}
	callTimeout, err := env.GetEnvAsDuration("SCHEDULER_CALL_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		TaskGRPCAddr: env.GetEnvOrDefault("TASK_GRPC_ADDR", "localhost:50052"),
		Interval:     interval,
		CallTimeout:  callTimeout,
	}
	return cfg, nil
}
