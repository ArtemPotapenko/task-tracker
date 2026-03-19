package testutil

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func StartPostgres(t *testing.T, dbName string) *sql.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgrescontainer.Run(
		ctx,
		"postgres:16-alpine",
		postgrescontainer.BasicWaitStrategies(),
		postgrescontainer.WithDatabase(dbName),
		postgrescontainer.WithUsername("admin"),
		postgrescontainer.WithPassword("secret"),
	)
	if err != nil {
		t.Fatalf("postgrescontainer.Run() error = %v", err)
	}

	t.Cleanup(func() {
		terminateCtx, terminateCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer terminateCancel()
		if err := container.Terminate(terminateCtx); err != nil {
			t.Fatalf("Terminate() error = %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("ConnectionString() error = %v", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer pingCancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		t.Fatalf("PingContext() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func ApplyGooseUpSQL(t *testing.T, db *sql.DB, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	content := string(data)
	upPart, _, _ := strings.Cut(content, "-- +goose Down")
	upPart = strings.Replace(upPart, "-- +goose Up", "", 1)
	upPart = strings.TrimSpace(upPart)

	if upPart == "" {
		t.Fatalf("migration %q has empty up section", path)
	}

	if _, err := db.Exec(upPart); err != nil {
		t.Fatalf("Exec(up migration %q) error = %v", path, err)
	}
}
