package repository

import (
	"context"
	"url_shortener/internal/model"
)

type Url interface {
	Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error)
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error)
	GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error)
}

type Click interface {
	GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, int, error)
	Create(ctx context.Context, urlId int, ip string, userAgent string) (*model.Click, error)
}
