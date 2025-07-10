package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"url_shortener/internal/model"
)

type Url struct {
	urlRepository   UrlRepository
	clickRepository ClickRepository
}

func NewUrl(urlRepository UrlRepository, clickRepository ClickRepository) *Url {
	return &Url{
		urlRepository:   urlRepository,
		clickRepository: clickRepository,
	}
}

func (u *Url) Create(ctx context.Context, longUrl, ip, userAgent string) (*model.Url, error) {
	shortCode := u.generateShortCode(longUrl, ip, userAgent)
	exists, err := u.urlRepository.ExistsByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("short code already exists")
	}
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) Click(ctx context.Context, shortCode, ip, userAgent string) (*model.Url, error) {
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	_, err = u.clickRepository.Create(ctx, url.Id, ip, userAgent) // TODO: run in worker pool
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) GetStats(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error) {
	url, err := u.urlRepository.GetWithClicksCountByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (u *Url) generateShortCode(longUrl, ip, userAgent string) string {
	key := fmt.Sprintf("%s_%s_%s", longUrl, ip, userAgent)
	shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	return shortCode[:6]
}
