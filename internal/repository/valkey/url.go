package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	"github.com/rs/zerolog/log"
	valkeygo "github.com/valkey-io/valkey-go"
)

type Url struct {
	client     valkeygo.Client
	repository repository.Url
}

func New(
	client valkeygo.Client,
	repository repository.Url,
) *Url {
	return &Url{
		client:     client,
		repository: repository,
	}
}

func (u *Url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	const op = "repository.valkey.Url.Create"
	url, err := u.repository.Create(ctx, longUrl, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := u.addToCache(ctx, url); err != nil {
		log.Error().
			Err(fmt.Errorf("%s: %w", op, err)).
			Int("id", url.Id).
			Str("short_code", url.ShortCode).
			Msg("failed to cache url")
	}
	return url, nil
}

func (u *Url) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	if _, err := u.getFromCache(ctx, shortCode); err == nil {
		return true, nil
	}
	return u.repository.ExistsByShortCode(ctx, shortCode)
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	const op = "repository.valkey.Url.GetByShortCode"
	if url, err := u.getFromCache(ctx, shortCode); err == nil {
		return url, nil
	}
	url, err := u.repository.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := u.addToCache(ctx, url); err != nil {
		log.Error().
			Err(fmt.Errorf("%s: %w", op, err)).
			Int("id", url.Id).
			Str("short_code", url.ShortCode).
			Msg("failed to cache url")
	}
	return url, nil
}

func (u *Url) GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error) {
	return u.repository.GetWithClicksCountByShortCode(ctx, shortCode)
}

func (u *Url) addToCache(ctx context.Context, url *model.Url) error {
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
		Ex(24 * time.Hour).
		Build()
	return u.client.Do(ctx, cmd).Error()
}

func (u *Url) getFromCache(ctx context.Context, shortCode string) (*model.Url, error) {
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
	var url model.Url
	err = json.Unmarshal([]byte(data), &url)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (u *Url) buildCacheKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}
