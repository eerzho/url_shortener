package service

import (
	"context"
	"crypto/rand"
	"errors"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
)

type Url struct {
	urlRepository UrlRepository
}

func NewUrl(urlRepository UrlRepository) *Url {
	return &Url{
		urlRepository: urlRepository,
	}
}

func (u *Url) Create(ctx context.Context, longUrl string) (*model.Url, error) {
	shortCode, err := u.generateShortCode(ctx)
	if err != nil {
		return nil, err
	}
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	url, err := u.urlRepository.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) generateShortCode(ctx context.Context) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	const attempts = 10
	for range attempts {
		b := make([]byte, length)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}

		for i := range b {
			b[i] = charset[b[i]%byte(len(charset))]
		}

		shortCode := string(b)
		_, err := u.urlRepository.GetByShortCode(ctx, shortCode)
		if errors.Is(err, constant.ErrNotFound) {
			return shortCode, nil
		}
	}
	return "", constant.ErrAlreadyExists
}
