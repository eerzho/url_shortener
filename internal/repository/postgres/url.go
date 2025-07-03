package postgres

import (
	"context"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	"github.com/jmoiron/sqlx"
)

type url struct {
	p *sqlx.DB
}

func NewUrl(p *sqlx.DB) repository.Url {
	return &url{p: p}
}

func (u *url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	var url model.Url
	err := u.p.GetContext(
		ctx,
		&url,
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

func (u *url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	var url model.Url
	err := u.p.GetContext(
		ctx,
		&url,
		`select * from urls where short_code = $1`,
		shortCode,
	)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (u *url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	var url model.Url
	err := u.p.GetContext(
		ctx,
		&url,
		`
			update urls set clicks = clicks + 1, updated_at = now()
			where short_code = $1
			returning *
		`,
		shortCode,
	)
	if err != nil {
		return nil, err
	}
	return &url, nil
}
