package valkey

import (
	"context"
	"encoding/json"
	"time"
	"url_shortener/internal/model"
	"url_shortener/internal/repository"

	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

const (
	urlTtl    = 24 * time.Hour
	urlPrefix = "url:"
)

type url struct {
	c valkey.Client
	r repository.Url
}

func NewUrl(c valkey.Client, r repository.Url) repository.Url {
	return &url{c: c, r: r}
}

func (u *url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	return u.r.Create(ctx, longUrl, shortCode)
}

func (u *url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	return u.r.GetByShortCode(ctx, shortCode)
}

func (u *url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "repository.valkey.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("getting url from cache")
	url, err := u.getFromCache(ctx, shortCode)
	if err == nil {
		logger.Debug().Int("id", url.Id).Msg("got url from cache")
		go u.incrementClicks(context.Background(), shortCode)
		return url, nil
	}
	logger.Debug().Err(err).Msg("failed to get url from cache")
	logger.Debug().Msg("getting url from repository")
	url, err = u.r.GetByShortCode(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url from repository")
		return nil, err
	}
	logger.Debug().Msg("caching url")
	err = u.addToCache(ctx, url)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to cache url")
	}
	logger.Debug().Int("id", url.Id).Msg("got url from repository")
	return url, nil
}

func (u *url) incrementClicks(ctx context.Context, shortCode string) {
	logger := log.With().
		Str("op", "repository.valkey.url.incrementClicks").
		Str("short_code", shortCode).
		Logger()
	logger.Debug().Msg("incrementing clicks")
	_, err := u.r.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		logger.Error().Err(err).Msg("failed to increment clicks")
	} else {
		logger.Debug().Msg("incremented clicks")
	}
}

func (u *url) getFromCache(ctx context.Context, shortCode string) (*model.Url, error) {
	key := urlPrefix + shortCode
	data, err := u.c.Do(ctx, u.c.B().Get().Key(key).Build()).ToString()
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

func (u *url) addToCache(ctx context.Context, url *model.Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	key := urlPrefix + url.ShortCode
	err = u.c.Do(
		ctx,
		u.c.B().
			Set().
			Key(key).
			Value(string(data)).
			Ex(urlTtl).
			Build(),
	).Error()
	if err != nil {
		return err
	}
	return nil
}
