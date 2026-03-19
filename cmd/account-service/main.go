package main

import (
	"task-tracker/internal/account/app"
	"task-tracker/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatalf("account service: %v", err)
	}
}
