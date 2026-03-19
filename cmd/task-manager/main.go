package main

import (
	"task-tracker/internal/task/app"
	"task-tracker/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("task service: %v", err)
	}
}
