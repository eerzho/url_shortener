package postgres

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func NewPostgresDB(url string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func MustNewPostgresDB(url string) *sqlx.DB {
	db, err := NewPostgresDB(url)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v\n", err)
	}
	return db
}
