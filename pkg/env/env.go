package env

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvAsInt(key string, defaultValue int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func GetEnvAsDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("empty duration value")
	}
	if isDigits(raw) {
		hours, err := strconv.Atoi(raw)
		if err != nil {
			return 0, err
		}
		return time.Duration(hours) * time.Hour, nil
	}
	return time.ParseDuration(raw)
}

func isDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
