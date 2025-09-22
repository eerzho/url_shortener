package repository

import (
	"context"
	"url_shortener/internal/model"
)

type URL interface {
	Create(ctx context.Context, shortCode, originalURL string) (*model.URL, error)
	GetOriginalURLByShortCode(ctx context.Context, shortCode string) (string, error)
}
