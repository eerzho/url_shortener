package postgres

import (
	"context"
	"database/sql"
	"errors"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type url struct {
	p *sqlx.DB
}

func NewUrl(p *sqlx.DB) repository.Url {
	return &url{p: p}
}

func (u *url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "repository.postgres.url.Create").
		Str("long_url", longUrl).
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("creating url")
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
		logger.Debug().Err(err).Msg("failed to create url")
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, constant.ErrAlreadyExists
			}
		}
		return nil, err
	}
	logger.Debug().Int("id", url.Id).Msg("created url")
	return &url, nil
}

func (u *url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "repository.postgres.url.GetByShortCode").
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("getting url")
	var url model.Url
	err := u.p.GetContext(
		ctx,
		&url,
		`select * from urls where short_code = $1`,
		shortCode,
	)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	logger.Debug().Int("id", url.Id).Msg("got url")
	return &url, nil
}

func (u *url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "repository.postgres.url.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("getting url")
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
		logger.Debug().Err(err).Msg("failed to get url")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	logger.Debug().Int("id", url.Id).Msg("got url")
	return &url, nil
}
