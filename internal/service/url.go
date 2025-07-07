package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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

func (u *Url) generateShortCode(ctx context.Context, longUrl string) (string, error) {
	input := longUrl
	for attempts := range 5 {
		shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))[:6]
		_, err := u.urlRepository.GetByShortCode(ctx, shortCode)
		if err != nil {
			if errors.Is(err, constant.ErrNotFound) {
				return shortCode, nil
			}
			return "", err
		}
		input = fmt.Sprintf("%s_%d", longUrl, attempts)
	}
	return "", constant.ErrAlreadyExists
}
