package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	valkeygo "github.com/valkey-io/valkey-go"
)

type URL struct {
	ttlURLCache time.Duration
	logger      *slog.Logger
	client      valkeygo.Client
	repo        repository.URLRepository
}

func NewURL(
	ttlURLCache time.Duration,
	logger *slog.Logger,
	client valkeygo.Client,
	repo repository.URLRepository,
) *URL {
	return &URL{
		ttlURLCache: ttlURLCache,
		logger:      logger,
		client:      client,
		repo:        repo,
	}
}

func (u *URL) Create(ctx context.Context, longURL, shortCode string) (*model.URL, error) {
	const op = "repository.valkey.Url.Create"
	url, err := u.repo.Create(ctx, longURL, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err = u.addToCache(ctx, url); err != nil {
		u.logger.ErrorContext(ctx, "failed to cache url",
			slog.Int("id", url.ID),
			slog.String("short_code", url.ShortCode),
			slog.Any("error", fmt.Errorf("%s: %w", op, err)),
		)
	}
	return url, nil
}

func (u *URL) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	if _, err := u.getFromCache(ctx, shortCode); err == nil {
		return true, nil
	}
	return u.repo.ExistsByShortCode(ctx, shortCode)
}

func (u *URL) GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {
	const op = "repository.valkey.Url.GetByShortCode"
	if url, err := u.getFromCache(ctx, shortCode); err == nil {
		return url, nil
	}
	url, err := u.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err = u.addToCache(ctx, url); err != nil {
		u.logger.ErrorContext(ctx, "failed to cache url",
			slog.Int("id", url.ID),
			slog.String("short_code", url.ShortCode),
			slog.Any("error", fmt.Errorf("%s: %w", op, err)),
		)
	}
	return url, nil
}

func (u *URL) GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.URLWithClicksCount, error) {
	return u.repo.GetWithClicksCountByShortCode(ctx, shortCode)
}

func (u *URL) addToCache(ctx context.Context, url *model.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	key := u.buildCacheKey(url.ShortCode)
	cmd := u.client.
		B().
		Set().
		Key(key).
		Value(string(data)).
		Ex(u.ttlURLCache).
		Build()
	return u.client.Do(ctx, cmd).Error()
}

func (u *URL) getFromCache(ctx context.Context, shortCode string) (*model.URL, error) {
	key := u.buildCacheKey(shortCode)
	cmd := u.client.
		B().
		Get().
		Key(key).
		Build()
	result := u.client.Do(ctx, cmd)
	if result.Error() != nil {
		return nil, result.Error()
	}
	data, err := result.ToString()
	if err != nil {
		return nil, err
	}
	var url model.URL
	err = json.Unmarshal([]byte(data), &url)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (u *URL) buildCacheKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}
