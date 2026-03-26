package main

import (
	"task-tracker/shared-libs/pkg/logger"
	"task-tracker/task-service/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("task service: %v", err)
	}
}
