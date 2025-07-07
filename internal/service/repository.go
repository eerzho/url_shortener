package service

import (
	"context"
	"url_shortener/internal/model"
)

type UrlRepository interface {
	Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error)
	GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error)
}
