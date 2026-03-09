package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite" // sqlite driver
)

// Builder is a global instance of the sql builder.
// sqlite is compatible with postgres-style dollar placeholders.
var Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Migrator is a function that applies database migrations.
type Migrator func(ctx context.Context, db *sqlx.DB) error

// Connect opens a database connection and verifies it with a ping.
func Connect(ctx context.Context, datasource string) (*sqlx.DB, error) {
	db, err := sql.Open("sqlite", datasource)
	if err != nil {
		return nil, fmt.Errorf("failed to open the db: %w", err)
	}

	// enable foreign keys (disabled by default in sqlite)
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	dbx := sqlx.NewDb(db, "sqlite")

	if err = pingDatabase(ctx, dbx); err != nil {
		return nil, fmt.Errorf("failed to ping the db: %w", err)
	}

	return dbx, nil
}

// ConnectAndMigrate creates the database handle and migrates the database.
func ConnectAndMigrate(ctx context.Context, datasource string, migrator Migrator) (*sqlx.DB, error) {
	dbx, err := Connect(ctx, datasource)
	if err != nil {
		return nil, err
	}

	if err = migrator(ctx, dbx); err != nil {
		return nil, fmt.Errorf("failed to setup the db: %w", err)
	}

	return dbx, nil
}

// pingDatabase attempts to ping the database with retries.
func pingDatabase(ctx context.Context, db *sqlx.DB) error {
	var err error
	for i := 1; i <= 5; i++ {
		err = db.PingContext(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(time.Duration(i) * 100 * time.Millisecond)
	}
	return err
}
