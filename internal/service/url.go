package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
)

type URL struct {
	urlRepository   URLRepository
	clickRepository ClickRepository
}

func NewURL(
	urlRepository URLRepository,
	clickRepository ClickRepository,
) *URL {
	return &URL{
		urlRepository:   urlRepository,
		clickRepository: clickRepository,
	}
}

func (u *URL) Create(ctx context.Context, longURL, ip, userAgent string) (*model.URL, error) {
	const op = "service.Url.Create"
	shortCode := u.generateShortCode(longURL, ip, userAgent)
	exists, err := u.urlRepository.ExistsByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", op, constant.ErrAlreadyExists)
	}
	url, err := u.urlRepository.Create(ctx, longURL, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *URL) Click(ctx context.Context, shortCode, ip, userAgent string) (*model.URL, error) {
	const op = "service.Url.Click"
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = u.clickRepository.Create(ctx, url.ID, ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (u *URL) GetStats(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error) {
	const op = "service.Url.GetStats"
	url, err := u.urlRepository.GetWithClicksCountByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *URL) generateShortCode(longURL, ip, userAgent string) string {
	key := fmt.Sprintf("%s_%s_%s", longURL, ip, userAgent)
	shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	return shortCode[:6]
}
