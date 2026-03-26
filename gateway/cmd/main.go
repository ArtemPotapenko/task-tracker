package main

import (
	"task-tracker/gateway/internal/app"
	"task-tracker/shared-libs/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("gateway: %v", err)
	}
}
