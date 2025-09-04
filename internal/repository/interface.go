package repository

import (
	"context"
	"url_shortener/internal/model"
)

type URLRepository interface {
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error)
	Create(ctx context.Context, longURL, shortCode string) (*model.URL, error)
}

type ClickRepository interface {
	GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, int, error)
	Create(ctx context.Context, urlID int, ip string, userAgent string) (*model.Click, error)
}
