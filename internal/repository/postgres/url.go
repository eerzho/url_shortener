package postgres

import (
	"context"
	"fmt"
	"url_shortener/internal/model"

	"github.com/jmoiron/sqlx"
)

type URL struct {
	db *sqlx.DB
}

func NewURL(
	db *sqlx.DB,
) *URL {
	return &URL{db: db}
}

func (u *URL) Create(ctx context.Context, shortCode, originalURL string) (*model.URL, error) {
	const op = "repository.postgres.URL.Create"

	var url model.URL
	err := u.db.GetContext(ctx, &url,
		`
			insert into urls (short_code, original_url)
			values ($1, $2)
			returning *
		`,
		shortCode, originalURL,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &url, nil
}

func (u *URL) GetOriginalURLByShortCode(ctx context.Context, shortCode string) (string, error) {
	const op = "repository.postgres.URL.GetOriginalURLByShortCode"

	var originalURL string
	err := u.db.GetContext(ctx, &originalURL,
		`
			select original_url from urls where short_code = $1
		`,
		shortCode,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return originalURL, nil
}
