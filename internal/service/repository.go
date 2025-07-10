package service

import (
	"context"
	"url_shortener/internal/model"
)

type UrlRepository interface {
	Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error)
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error)
	GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error)
}

type ClickRepository interface {
	Create(ctx context.Context, urlId int, ip string, userAgent string) (*model.Click, error)
}
