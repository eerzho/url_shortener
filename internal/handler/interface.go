package handler

import (
	"context"
	"url_shortener/internal/model"
)

type URLService interface {
	Create(ctx context.Context, originalURL string) (*model.URL, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}
