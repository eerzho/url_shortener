package postgres

import (
	"context"
	"database/sql"
	"errors"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"

	"github.com/jmoiron/sqlx"
)

type Url struct {
	db *sqlx.DB
}

func NewUrl(db *sqlx.DB) *Url {
	return &Url{db: db}
}

func (u *Url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	var url model.Url
	err := u.db.GetContext(ctx, &url,
		`
			insert into urls (long_url, short_code)
			values ($1, $2)
			returning *
		`,
		longUrl, shortCode,
	)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (u *Url) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	var count int
	err := u.db.GetContext(ctx, &count,
		`select count(*) from urls where short_code = $1`,
		shortCode,
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	var url model.Url
	err := u.db.GetContext(ctx, &url,
		`select * from urls where short_code = $1`,
		shortCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &url, nil
}

func (u *Url) GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error) {
	var url model.UrlWithClicksCount
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
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &url, nil
}
