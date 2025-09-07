package service

import (
	"context"
	"fmt"
	"url_shortener/internal/model"
	"url_shortener/internal/utils/base62"
)

type URL struct {
	urlRepository     URLRepository
	counterRepository CounterRepository
}

func NewURL(
	urlRepository URLRepository,
	counterRepository CounterRepository,
) *URL {
	return &URL{
		urlRepository:     urlRepository,
		counterRepository: counterRepository,
	}
}

func (u *URL) Create(ctx context.Context, originalURL string) (*model.URL, error) {
	const op = "service.URL.Create"

	shortCode, err := u.generateShortCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	url, err := u.urlRepository.Create(ctx, shortCode, originalURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (u *URL) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	const op = "service.URL.GetOriginalURL"

	originalURL, err := u.urlRepository.GetOriginalURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return originalURL, nil
}

func (u *URL) generateShortCode(ctx context.Context) (string, error) {
	const op = "service.URL.generateShortCode"

	num, err := u.counterRepository.Incr(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return base62.Encode(num), nil
}
