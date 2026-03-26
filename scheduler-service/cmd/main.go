package main

import (
	"task-tracker/scheduler-service/internal/app"
	"task-tracker/shared-libs/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("scheduler: %v", err)
	}
}
