package repository

import (
	"context"
	"url_shortener/internal/model"
)

type URLReader interface {
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error)
}

type URLWriter interface {
	Create(ctx context.Context, longURL, shortCode string) (*model.URL, error)
}

type ClickReader interface {
	GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, int, error)
}

type ClickWriter interface {
	Create(ctx context.Context, urlID int, ip string, userAgent string) (*model.Click, error)
}
