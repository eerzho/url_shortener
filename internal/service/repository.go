package service

import (
	"context"
	"url_shortener/internal/model"
)

type URLRepository interface {
	Create(ctx context.Context, longURL, shortCode string) (*model.URL, error)
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error)
}

type ClickRepository interface {
	GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, int, error)
	Create(ctx context.Context, urlID int, ip, userAgent string) (*model.Click, error)
}
