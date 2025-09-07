package service

import (
	"context"
	"url_shortener/internal/model"
)

type URLRepository interface {
	Create(ctx context.Context, shortCode, originalURL string) (*model.URL, error)
	GetOriginalURLByShortCode(ctx context.Context, shortCode string) (string, error)
}

type CounterRepository interface {
	Incr(ctx context.Context) (int, error)
}
