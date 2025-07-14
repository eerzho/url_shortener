package postgres

import (
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

const (
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5 * time.Minute
)

func NewPostgresDB(url string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(DefaultMaxOpenConns)
	db.SetMaxIdleConns(DefaultMaxIdleConns)
	db.SetConnMaxLifetime(DefaultConnMaxLifetime)

	return db, nil
}

func MustNewPostgresDB(
	logger *slog.Logger,
	url string,
) *sqlx.DB {
	db, err := NewPostgresDB(url)
	if err != nil {
		logger.Error("failed to connect to postgres",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	return db
}
