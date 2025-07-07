package handler

import (
	"context"
	"url_shortener/internal/model"
)

type UrlService interface {
	Create(ctx context.Context, longUrl string) (*model.Url, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error)
	GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error)
}
