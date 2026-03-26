package main

import (
	"task-tracker/email-service/internal/app"
	"task-tracker/shared-libs/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("email service: %v", err)
	}
}
