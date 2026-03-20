package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func UpSQL(conn *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	content := string(data)
	upPart, _, _ := strings.Cut(content, "-- +goose Down")
	upPart = strings.Replace(upPart, "-- +goose Up", "", 1)
	upPart = strings.TrimSpace(upPart)
	if upPart == "" {
		return fmt.Errorf("migration %s has empty up section", path)
	}

	if _, err := conn.Exec(upPart); err != nil {
		return fmt.Errorf("exec migration %s: %w", path, err)
	}

	return nil
}
