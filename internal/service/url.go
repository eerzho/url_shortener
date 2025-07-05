package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	"github.com/rs/zerolog/log"
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
	logger := log.With().
		Str("op", "service.url.Create").
		Str("long_url", longUrl).
		Logger()
	logger.Debug().Msg("creating url")
	if longUrl == "" {
		err := errors.New("logn_url is empty")
		logger.Debug().Err(err).Msg("failed to create url")
		return nil, err
	}
	shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(longUrl)))[:6]
	logger.Debug().Str("short_code", shortCode).Msg("generated short_code")
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to create url")
		return nil, err
	}
	logger.Debug().Int("id", url.Id).Msg("created url")
	return url, nil
}

func (u *url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "service.url.GetByShortCode").
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("getting url")
	if shortCode == "" {
		err := errors.New("short_code is empty")
		logger.Debug().Err(err).Msg("failed to get url")
		return nil, err
	}
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url")
		return nil, err
	}
	logger.Debug().Int("id", url.Id).Msg("got url")
	return url, nil
}

func (u *url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "service.url.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("getting url")
	if shortCode == "" {
		err := errors.New("short_code is empty")
		logger.Debug().Err(err).Msg("failed to get url")
		return nil, err
	}
	url, err := u.urlRepository.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url")
		return nil, err
	}
	logger.Debug().Int("id", url.Id).Msg("got url")
	return url, nil
}
