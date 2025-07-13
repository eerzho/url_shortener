package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
	"url_shortener/pkg/async"

	"github.com/rs/zerolog"
)

type URL struct {
	logger          zerolog.Logger
	pool            *async.WorkerPool
	urlRepository   URLRepository
	clickRepository ClickRepository
}

func NewURL(
	logger zerolog.Logger,
	urlRepository URLRepository,
	clickRepository ClickRepository,
) *URL {
	pool := async.NewWorkerPool(10, 50_000)
	pool.Start()

	return &URL{
		logger:          logger,
		pool:            pool,
		urlRepository:   urlRepository,
		clickRepository: clickRepository,
	}
}

func (u *URL) Close() {
	u.pool.Shutdown()
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
	clickJob := func(jobCtx context.Context, workerId int) {
		logger := u.logger.
			With().
			Str("ip", ip).
			Int("url_id", url.ID).
			Int("worker_id", workerId).
			Str("user_agent", userAgent).
			Str("url_short_code", shortCode).
			Logger()
		logger.Debug().Msg("creating click")
		click, err := u.clickRepository.Create(jobCtx, url.ID, ip, userAgent)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create click")
			return
		}
		logger.Debug().Int("click_id", click.ID).Msg("created click")
	}
	err = u.pool.Submit(clickJob)
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
