package main

import (
	"task-tracker/internal/email/app"
	"task-tracker/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("email service: %v", err)
	}
}
