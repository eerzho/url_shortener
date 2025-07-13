package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
	"url_shortener/pkg/async"

	"github.com/rs/zerolog/log"
)

type Url struct {
	pool            *async.WorkerPool
	urlRepository   UrlRepository
	clickRepository ClickRepository
}

func NewUrl(
	urlRepository UrlRepository,
	clickRepository ClickRepository,
) *Url {
	pool := async.NewWorkerPool(10, 1_000)
	pool.Start()

	return &Url{
		pool:            pool,
		urlRepository:   urlRepository,
		clickRepository: clickRepository,
	}
}

func (u *Url) Close() {
	u.pool.Shutdown()
}

func (u *Url) Create(ctx context.Context, longUrl, ip, userAgent string) (*model.Url, error) {
	const op = "service.Url.Create"
	shortCode := u.generateShortCode(longUrl, ip, userAgent)
	exists, err := u.urlRepository.ExistsByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", op, constant.ErrAlreadyExists)
	}
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *Url) Click(ctx context.Context, shortCode, ip, userAgent string) (*model.Url, error) {
	const op = "service.Url.Click"
	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	clickJob := func(jobCtx context.Context, workerId int) {
		logger := log.With().
			Str("ip", ip).
			Int("url_id", url.Id).
			Int("worker_id", workerId).
			Str("user_agent", userAgent).
			Str("url_short_code", shortCode).
			Logger()
		logger.Debug().Msg("creating click")
		click, err := u.clickRepository.Create(jobCtx, url.Id, ip, userAgent)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create click")
			return
		}
		logger.Debug().Int("click_id", click.Id).Msg("created click")
	}
	err = u.pool.Submit(clickJob)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *Url) GetStats(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error) {
	const op = "service.Url.GetStats"
	url, err := u.urlRepository.GetWithClicksCountByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *Url) generateShortCode(longUrl, ip, userAgent string) string {
	key := fmt.Sprintf("%s_%s_%s", longUrl, ip, userAgent)
	shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	return shortCode[:6]
}
