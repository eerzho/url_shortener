package repository

import (
	"context"
	"url_shortener/internal/model"
)

type Url interface {
	Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error)
	GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error)
}
