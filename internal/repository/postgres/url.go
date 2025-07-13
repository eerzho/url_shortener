package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"

	"github.com/jmoiron/sqlx"
)

type URL struct {
	db *sqlx.DB
}

func NewURL(db *sqlx.DB) *URL {
	return &URL{db: db}
}

func (u *URL) Create(ctx context.Context, longURL, shortCode string) (*model.URL, error) {
	const op = "repository.postgres.Url.Create"
	var url model.URL
	err := u.db.GetContext(ctx, &url,
		`
			insert into urls (long_url, short_code)
			values ($1, $2)
			returning *
		`,
		longURL, shortCode,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &url, nil
}

func (u *URL) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	const op = "repository.postgres.Url.ExistsByShortCode"
	var count int
	err := u.db.GetContext(ctx, &count,
		`select count(*) from urls where short_code = $1`,
		shortCode,
	)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return count > 0, nil
}

func (u *URL) GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {
	const op = "repository.postgres.Url.GetByShortCode"
	var url model.URL
	err := u.db.GetContext(ctx, &url,
		`select * from urls where short_code = $1`,
		shortCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, constant.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &url, nil
}

func (u *URL) GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error) {
	const op = "repository.postgres.Url.GetWithClicksCountByShortCode"
	var url model.URLWithClicksCount
	err := u.db.GetContext(ctx, &url,
		`
			select u.*, count(c.id) clicks_count from urls u
			left join clicks c on u.id = c.url_id
			where short_code = $1
			group by u.id, u.short_code, u.long_url, u.created_at, u.updated_at
		`,
		shortCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, constant.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &url, nil
}
