package handler

import (
	"context"
	"net/http"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
)

type URLService interface {
	Create(ctx context.Context, longURL, ip, userAgent string) (*model.URL, error)
	Click(ctx context.Context, shortCode, ip, userAgent string) (*model.URL, error)
	GetStats(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error)
}

type IPService interface {
	GetIP(ctx context.Context, r *http.Request) string
}

type ClickService interface {
	GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, *dto.Pagination, error)
}
