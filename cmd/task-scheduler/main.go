package main

import (
	"task-tracker/internal/scheduler/app"
	"task-tracker/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("scheduler: %v", err)
	}
}
