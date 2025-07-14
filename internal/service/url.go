package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
	"url_shortener/pkg/async"
)

const (
	DefaultWorkerCount = 10
	DefaultBufferSize  = 50_000
)

type URL struct {
	logger      *slog.Logger
	pool        *async.WorkerPool
	urlReader   URLReader
	urlWriter   URLWriter
	clickWriter ClickWriter
}

func NewURL(
	logger *slog.Logger,
	urlReader URLReader,
	urlWriter URLWriter,
	clickWriter ClickWriter,
) *URL {
	pool := async.NewWorkerPool(
		DefaultWorkerCount,
		DefaultBufferSize,
	)
	pool.Start()

	return &URL{
		logger:      logger,
		pool:        pool,
		urlReader:   urlReader,
		urlWriter:   urlWriter,
		clickWriter: clickWriter,
	}
}

func (u *URL) Close() {
	u.pool.Shutdown()
}

func (u *URL) Create(ctx context.Context, longURL, ip, userAgent string) (*model.URL, error) {
	const op = "service.Url.Create"
	shortCode := u.generateShortCode(longURL, ip, userAgent)
	exists, err := u.urlReader.ExistsByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", op, constant.ErrAlreadyExists)
	}
	url, err := u.urlWriter.Create(ctx, longURL, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *URL) Click(ctx context.Context, shortCode, ip, userAgent string) (*model.URL, error) {
	const op = "service.Url.Click"
	url, err := u.urlReader.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	clickJob := func(jobCtx context.Context, workerId int) {
		logger := u.logger.With(
			slog.String("ip", ip),
			slog.Int("url_id", url.ID),
			slog.Int("worker_id", workerId),
			slog.String("user_agent", userAgent),
			slog.String("url_short_code", shortCode),
		)

		logger.DebugContext(jobCtx, "creating click")

		click, e := u.clickWriter.Create(jobCtx, url.ID, ip, userAgent)
		if e != nil {
			logger.ErrorContext(jobCtx, "failed to create click",
				slog.Any("error", e),
			)
			return
		}

		logger.DebugContext(jobCtx, "created click",
			slog.Int("click_id", click.ID),
		)
	}
	err = u.pool.Submit(clickJob)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (u *URL) GetStats(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error) {
	const op = "service.Url.GetStats"
	url, err := u.urlReader.GetWithClicksCountByShortCode(ctx, shortCode)
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
