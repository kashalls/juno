// Package db provides juno's embedded persistence layer: feature state
// (temp voice channel bookkeeping) that must survive process restarts.
package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// DB wraps a single-writer SQLite connection. discordgo dispatches gateway
// handlers concurrently, so writes are serialized through one connection
// rather than relying on SQLite's file locking under contention.
type DB struct {
	conn *sql.DB
}

func Open(ctx context.Context, path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.ExecContext(ctx, "PRAGMA busy_timeout = 5000; PRAGMA journal_mode = WAL;"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set pragmas: %w", err)
	}
	if _, err := conn.ExecContext(ctx, schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}
