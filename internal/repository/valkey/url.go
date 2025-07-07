package valkey

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	"url_shortener/internal/model"
	"url_shortener/internal/service"

	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

const (
	urlTtl         = 24 * time.Hour
	urlPrefix      = "url:"
	urlWorkerCount = 3
	bufferSize     = 500
)

type Url struct {
	client      valkey.Client
	repo        service.UrlRepository
	wg          sync.WaitGroup
	doneCh      chan struct{}
	incrementCh chan string
}

func NewUrl(client valkey.Client, repo service.UrlRepository) *Url {
	u := &Url{
		client:      client,
		repo:        repo,
		doneCh:      make(chan struct{}),
		incrementCh: make(chan string, bufferSize),
	}
	u.startWorkers()
	return u
}

func (u *Url) Close() {
	close(u.incrementCh)
	u.wg.Wait()
	close(u.doneCh)
}

func (u *Url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	return u.repo.Create(ctx, longUrl, shortCode)
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	return u.repo.GetByShortCode(ctx, shortCode)
}

func (u *Url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	logger := log.With().
		Str("op", "repository.valkey.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("getting url from cache")

	url, err := u.getFromCache(ctx, shortCode)
	if err == nil {
		logger.Debug().Int("id", url.Id).Msg("got url from cache")

		u.scheduleIncrement(shortCode)

		return url, nil
	}

	logger.Debug().Err(err).Msg("failed to get url from cache")

	logger.Debug().Msg("getting url from repository")

	url, err = u.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url from repository")
		return nil, err
	}

	go func() {
		logger.Debug().Msg("caching url")

		err = u.addToCache(context.Background(), url)
		if err != nil {
			logger.Debug().Err(err).Msg("failed to cache url")
		}
	}()

	logger.Debug().Int("id", url.Id).Msg("got url from repository")

	return url, nil
}

func (u *Url) scheduleIncrement(shortCode string) {
	select {
	case u.incrementCh <- shortCode:
	default:
		log.Warn().Str("short_code", shortCode).Msg("increment channel is full, skipping increment")
	}
}

func (u *Url) startWorkers() {
	for range urlWorkerCount {
		u.wg.Add(1)
		go u.worker()
	}
}

func (u *Url) worker() {
	defer u.wg.Done()

	for {
		select {
		case shortCode, ok := <-u.incrementCh:
			if !ok {
				return
			}
			u.incrementClicks(shortCode)
		case <-u.doneCh:
			return
		}
	}
}

func (u *Url) incrementClicks(shortCode string) {
	logger := log.With().
		Str("op", "repository.valkey.url.incrementClicks").
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("incrementing clicks")

	_, err := u.repo.GetByShortCodeAndIncrementClicks(context.Background(), shortCode)
	if err != nil {
		logger.Error().Err(err).Msg("failed to increment clicks")
	} else {
		logger.Debug().Msg("incremented clicks")
	}
}

func (u *Url) getFromCache(ctx context.Context, shortCode string) (*model.Url, error) {
	key := urlPrefix + shortCode
	data, err := u.client.Do(ctx, u.client.B().Get().Key(key).Build()).ToString()
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

func (u *Url) addToCache(ctx context.Context, url *model.Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	key := urlPrefix + url.ShortCode
	err = u.client.Do(
		ctx,
		u.client.B().
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
