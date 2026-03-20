package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	pkgdb "task-tracker/pkg/db"
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

func ApplyUpSQL(t *testing.T, db *sql.DB, path string) {
	t.Helper()

	if err := pkgdb.UpSQL(db, path); err != nil {
		t.Fatalf("ApplyUpSQL(%q) error = %v", path, err)
	}
}
