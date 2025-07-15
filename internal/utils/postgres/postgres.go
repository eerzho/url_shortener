package postgres

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func NewPostgresDB(
	url string,
	maxOpenConns int,
	maxIdleConns int,
	connMaxLifetime time.Duration,
) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

func MustNewPostgresDB(
	url string,
	maxOpenConns int,
	maxIdleConns int,
	connMaxLifetime time.Duration,
) *sqlx.DB {
	db, err := NewPostgresDB(
		url,
		maxOpenConns,
		maxIdleConns,
		connMaxLifetime,
	)
	if err != nil {
		panic(err)
	}
	return db
}
