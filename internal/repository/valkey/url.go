package valkey

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	valkeygo "github.com/valkey-io/valkey-go"
)

type URL struct {
	ttl           time.Duration
	logger        *slog.Logger
	client        valkeygo.Client
	urlRepository repository.URL
}

func NewURL(
	ttl time.Duration,
	logger *slog.Logger,
	client valkeygo.Client,
	urlRepository repository.URL,
) *URL {
	return &URL{
		ttl:           ttl,
		logger:        logger,
		client:        client,
		urlRepository: urlRepository,
	}
}

func (u *URL) Create(ctx context.Context, shortCode, originalURL string) (*model.URL, error) {
	const op = "repository.valkey.URL.Create"

	url, err := u.urlRepository.Create(ctx, shortCode, originalURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := u.setCache(ctx, shortCode, originalURL); err != nil {
		u.logger.WarnContext(ctx, "failed to set cache",
			slog.Int("id", url.ID),
			slog.String("short_code", url.ShortCode),
			slog.String("original_url", url.OriginalURL),
			slog.Time("created_at", url.CreatedAt),
			slog.Time("updated_at", url.UpdatedAt),
			slog.Any("error", fmt.Errorf("%s: %w", op, err)),
		)
	}

	return url, nil
}

func (u *URL) GetOriginalURLByShortCode(ctx context.Context, shortCode string) (string, error) {
	const op = "repository.valkey.URL.GetOriginalURLByShortCode"

	originalURL, err := u.getCache(ctx, shortCode)
	if err == nil {
		return originalURL, nil
	}

	if !valkeygo.IsValkeyNil(err) {
		u.logger.WarnContext(ctx, "failed to get cache",
			slog.String("short_code", shortCode),
			slog.Any("error", fmt.Errorf("%s: %w", op, err)),
			slog.String("original_url", originalURL),
		)
	}

	originalURL, err = u.urlRepository.GetOriginalURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := u.setCache(ctx, shortCode, originalURL); err != nil {
		u.logger.WarnContext(ctx, "failed to set cache",
			slog.String("short_code", shortCode),
			slog.Any("error", fmt.Errorf("%s: %w", op, err)),
			slog.String("original_url", originalURL),
		)
	}

	return originalURL, nil
}

func (u *URL) setCache(ctx context.Context, shortCode, originalURL string) error {
	key := u.buildKey(shortCode)
	cmd := u.client.B().Set().Key(key).Value(originalURL).Ex(u.ttl).Build()
	result := u.client.Do(ctx, cmd)
	return result.Error()
}

func (u *URL) getCache(ctx context.Context, shortCode string) (string, error) {
	key := u.buildKey(shortCode)
	cmd := u.client.B().Get().Key(key).Build()
	result := u.client.Do(ctx, cmd)
	if result.Error() != nil {
		return "", result.Error()
	}
	value, err := result.ToString()
	if err != nil {
		return "", err
	}
	return value, nil
}

func (u *URL) buildKey(shortCode string) string {
	return fmt.Sprintf("urls:%s", shortCode)
}
