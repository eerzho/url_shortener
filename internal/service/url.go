package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"
)

type Url struct {
	urlRepository repository.Url
}

func NewUrl(urlRepository repository.Url) *Url {
	return &Url{
		urlRepository: urlRepository,
	}
}

func (u *Url) Create(ctx context.Context, longUrl string) (*model.Url, error) {
	if longUrl == "" {
		return nil, fmt.Errorf("longUrl is empty")
	}
	shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(longUrl)))[:6]
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	if shortCode == "" {
		return nil, fmt.Errorf("shortCode is empty")
	}
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	if shortCode == "" {
		return nil, fmt.Errorf("shortCode is empty")
	}
	url, err := u.urlRepository.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}
