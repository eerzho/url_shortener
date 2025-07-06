package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"
)

type Url interface {
	Create(ctx context.Context, longUrl string) (*model.Url, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error)
	GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error)
}

type url struct {
	urlRepository repository.Url
}

func NewUrl(urlRepository repository.Url) Url {
	return &url{
		urlRepository: urlRepository,
	}
}

func (u *url) Create(ctx context.Context, longUrl string) (*model.Url, error) {
	if longUrl == "" {
		return nil, errors.New("logn_url is empty")
	}
	shortCode, err := u.generateShortCode(ctx, longUrl)
	if err != nil {
		return nil, err
	}
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	if shortCode == "" {
		return nil, errors.New("short_code is empty")
	}
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	if shortCode == "" {
		return nil, errors.New("short_code is empty")
	}
	url, err := u.urlRepository.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *url) generateShortCode(ctx context.Context, longUrl string) (string, error) {
	input := longUrl
	for attempts := range 5 {
		shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))[:6]
		url, _ := u.urlRepository.GetByShortCode(ctx, shortCode)
		if url == nil {
			return shortCode, nil
		}
		input = fmt.Sprintf("%s_%d", longUrl, attempts)
	}
	return "", errors.New("failed to generate unique short code")
}
