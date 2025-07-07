package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"url_shortener/internal/constant"
	"url_shortener/internal/metrics"
	"url_shortener/internal/model"
	"url_shortener/internal/service"

	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

const (
	urlTtl         = constant.DefaultCacheTTL
	urlPrefix      = constant.DefaultCacheKeyPrefix
	urlWorkerCount = constant.DefaultWorkerCount
	bufferSize     = constant.DefaultBufferSize
)

type Url struct {
	client      valkey.Client
	repo        service.UrlRepository
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	incrementCh chan string
	cacheStats  *metrics.CacheStats
	workerID    int
}

func NewUrl(client valkey.Client, repo service.UrlRepository) *Url {
	ctx, cancel := context.WithCancel(context.Background())

	u := &Url{
		client:      client,
		repo:        repo,
		ctx:         ctx,
		cancel:      cancel,
		incrementCh: make(chan string, bufferSize),
		cacheStats:  metrics.NewCacheStats(),
	}

	u.startWorkers()

	// Start metrics updater
	go u.metricsUpdater()

	return u
}

func (u *Url) Close() {
	log.Info().Msg("shutting down valkey url repository")

	// Cancel context to signal workers to stop
	u.cancel()

	// Close the increment channel
	close(u.incrementCh)

	// Wait for all workers to finish
	u.wg.Wait()

	log.Info().Msg("valkey url repository shutdown complete")
}

func (u *Url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("create"))
	}()

	url, err := u.repo.Create(ctx, longUrl, shortCode)
	if err != nil {
		metrics.RecordDatabaseOperation("create", "error", timer.Duration())
		return nil, err
	}

	metrics.RecordDatabaseOperation("create", "success", timer.Duration())

	// Cache the newly created URL asynchronously
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := u.addToCache(cacheCtx, url); err != nil {
			log.Error().Err(err).Str("short_code", url.ShortCode).Msg("failed to cache newly created url")
			metrics.RecordError("cache", "add_failed")
		}
	}()

	return url, nil
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	timer := metrics.NewTimer()

	logger := log.With().
		Str("op", "repository.valkey.GetByShortCode").
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("getting url from cache")

	// Try to get from cache first
	url, err := u.getFromCache(ctx, shortCode)
	if err == nil {
		logger.Debug().Int("id", url.Id).Msg("got url from cache")
		u.cacheStats.RecordHit()
		metrics.RecordCacheOperation("get", "hit")
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("get_cached"))
		return url, nil
	}

	logger.Debug().Err(err).Msg("cache miss, getting from repository")
	u.cacheStats.RecordMiss()
	metrics.RecordCacheOperation("get", "miss")

	// Get from repository
	url, err = u.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url from repository")
		metrics.RecordDatabaseOperation("get", "error", timer.Duration())
		return nil, err
	}

	metrics.RecordDatabaseOperation("get", "success", timer.Duration())

	// Cache the result asynchronously
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Debug().Msg("caching url")
		if err := u.addToCache(cacheCtx, url); err != nil {
			logger.Error().Err(err).Msg("failed to cache url")
			metrics.RecordError("cache", "add_failed")
		} else {
			logger.Debug().Msg("url cached successfully")
		}
	}()

	logger.Debug().Int("id", url.Id).Msg("got url from repository")
	return url, nil
}

func (u *Url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	timer := metrics.NewTimer()

	logger := log.With().
		Str("op", "repository.valkey.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("getting url from cache")

	// Try to get from cache first
	url, err := u.getFromCache(ctx, shortCode)
	if err == nil {
		logger.Debug().Int("id", url.Id).Msg("got url from cache")
		u.cacheStats.RecordHit()
		metrics.RecordCacheOperation("get", "hit")

		// Schedule async increment
		u.scheduleIncrement(shortCode)

		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("get_cached"))
		return url, nil
	}

	logger.Debug().Err(err).Msg("cache miss, getting from repository")
	u.cacheStats.RecordMiss()
	metrics.RecordCacheOperation("get", "miss")

	// Get from repository without incrementing clicks first
	url, err = u.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url from repository")
		metrics.RecordDatabaseOperation("get", "error", timer.Duration())
		return nil, err
	}

	// Cache the result asynchronously
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Debug().Msg("caching url")
		if err := u.addToCache(cacheCtx, url); err != nil {
			logger.Error().Err(err).Msg("failed to cache url")
			metrics.RecordError("cache", "add_failed")
		}
	}()

	// Schedule async increment
	u.scheduleIncrement(shortCode)

	logger.Debug().Int("id", url.Id).Msg("got url from repository")
	metrics.RecordDatabaseOperation("get", "success", timer.Duration())

	return url, nil
}

func (u *Url) scheduleIncrement(shortCode string) {
	select {
	case u.incrementCh <- shortCode:
		// Successfully scheduled
	default:
		// Channel is full, log warning and record metric
		log.Warn().
			Str("short_code", shortCode).
			Int("queue_size", len(u.incrementCh)).
			Msg("increment channel is full, skipping increment")
		metrics.RecordError("worker", "queue_full")
	}
}

func (u *Url) startWorkers() {
	for i := 0; i < urlWorkerCount; i++ {
		u.wg.Add(1)
		go u.worker(i)
	}
}

func (u *Url) worker(workerID int) {
	defer u.wg.Done()

	logger := log.With().
		Str("op", "repository.valkey.worker").
		Int("worker_id", workerID).
		Logger()

	logger.Debug().Msg("worker started")

	for {
		select {
		case shortCode, ok := <-u.incrementCh:
			if !ok {
				logger.Debug().Msg("increment channel closed, worker stopping")
				return
			}
			u.incrementClicks(shortCode, workerID)
		case <-u.ctx.Done():
			logger.Debug().Msg("context cancelled, worker stopping")
			return
		}
	}
}

func (u *Url) incrementClicks(shortCode string, workerID int) {
	timer := metrics.NewTimer()

	logger := log.With().
		Str("op", "repository.valkey.incrementClicks").
		Str("short_code", shortCode).
		Int("worker_id", workerID).
		Logger()

	logger.Debug().Msg("incrementing clicks")

	// Create a timeout context for the database operation
	ctx, cancel := context.WithTimeout(context.Background(), constant.DefaultDBTimeout)
	defer cancel()

	_, err := u.repo.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		logger.Error().Err(err).Msg("failed to increment clicks")
		metrics.RecordWorkerTask("increment", "error", timer.Duration())
		metrics.RecordError("worker", "increment_failed")
		return
	}

	logger.Debug().Msg("clicks incremented successfully")
	metrics.RecordWorkerTask("increment", "success", timer.Duration())

	// Remove from cache to ensure fresh data on next access
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := u.removeFromCache(cacheCtx, shortCode); err != nil {
			logger.Debug().Err(err).Msg("failed to remove from cache after increment")
		}
	}()
}

func (u *Url) getFromCache(ctx context.Context, shortCode string) (*model.Url, error) {
	key := urlPrefix + shortCode

	cmd := u.client.B().Get().Key(key).Build()
	data, err := u.client.Do(ctx, cmd).ToString()
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var url model.Url
	if err := json.Unmarshal([]byte(data), &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return &url, nil
}

func (u *Url) addToCache(ctx context.Context, url *model.Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal url: %w", err)
	}

	key := urlPrefix + url.ShortCode
	cmd := u.client.B().Set().Key(key).Value(string(data)).Ex(urlTtl).Build()

	if err := u.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (u *Url) removeFromCache(ctx context.Context, shortCode string) error {
	key := urlPrefix + shortCode
	cmd := u.client.B().Del().Key(key).Build()

	if err := u.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

func (u *Url) metricsUpdater() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update cache hit ratio
			ratio := u.cacheStats.HitRatio()
			metrics.UpdateCacheHitRatio("valkey", ratio)

			// Update worker queue size
			queueSize := float64(len(u.incrementCh))
			metrics.SetWorkerQueueSize("increment", queueSize)

		case <-u.ctx.Done():
			return
		}
	}
}

// GetCacheStats returns cache statistics
func (u *Url) GetCacheStats() (hits, misses, requests int64) {
	return u.cacheStats.Stats()
}

// GetQueueSize returns the current queue size
func (u *Url) GetQueueSize() int {
	return len(u.incrementCh)
}

// Health check method
func (u *Url) HealthCheck(ctx context.Context) error {
	cmd := u.client.B().Ping().Build()
	return u.client.Do(ctx, cmd).Error()
}
