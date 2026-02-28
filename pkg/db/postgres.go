package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func Open(ctx context.Context, driver, dsn string, pingTimeout time.Duration) (*sql.DB, error) {
	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := conn.PingContext(pingCtx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return conn, nil
}
