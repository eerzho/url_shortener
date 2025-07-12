package handler

import (
	"context"
	"net/http"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
)

type UrlService interface {
	Create(ctx context.Context, longUrl, ip, userAgent string) (*model.Url, error)
	Click(ctx context.Context, shortCode, ip, userAgent string) (*model.Url, error)
	GetStats(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error)
}

type IpService interface {
	GetIp(ctx context.Context, r *http.Request) string
}

type ClickService interface {
	GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, *dto.Pagination, error)
}
