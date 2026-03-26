package main

import (
	"task-tracker/account-service/internal/app"
	"task-tracker/shared-libs/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("account service: %v", err)
	}
}
