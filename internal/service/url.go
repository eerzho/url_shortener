package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"
	"url_shortener/internal/constant"
	"url_shortener/internal/metrics"
	"url_shortener/internal/model"

	"github.com/rs/zerolog/log"
)

type Url struct {
	urlRepository UrlRepository
}

func NewUrl(urlRepository UrlRepository) *Url {
	return &Url{
		urlRepository: urlRepository,
	}
}

func (u *Url) Create(ctx context.Context, longUrl string) (*model.Url, error) {
	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("service_create"))
	}()

	logger := log.With().
		Str("op", "service.url.Create").
		Str("long_url", longUrl).
		Logger()

	logger.Debug().Msg("creating short code")

	// Validate URL
	if err := u.validateURL(longUrl); err != nil {
		logger.Debug().Err(err).Msg("url validation failed")
		metrics.RecordURLCreated("validation_error")
		return nil, fmt.Errorf("url validation failed: %w", err)
	}

	// Generate unique short code
	shortCode, err := u.generateShortCode(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate short code")
		metrics.RecordURLCreated("generation_error")
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}

	logger.Debug().Str("short_code", shortCode).Msg("generated short code")

	// Create URL in repository
	url, err := u.urlRepository.Create(ctx, longUrl, shortCode)
	if err != nil {
		logger.Error().Err(err).Str("short_code", shortCode).Msg("failed to create url in repository")
		metrics.RecordURLCreated("repository_error")
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	logger.Info().
		Int("id", url.Id).
		Str("short_code", url.ShortCode).
		Msg("url created successfully")

	metrics.RecordURLCreated("success")
	return url, nil
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("service_get"))
	}()

	logger := log.With().
		Str("op", "service.url.GetByShortCode").
		Str("short_code", shortCode).
		Logger()

	// Validate short code
	if err := u.validateShortCode(shortCode); err != nil {
		logger.Debug().Err(err).Msg("short code validation failed")
		return nil, fmt.Errorf("invalid short code: %w", err)
	}

	logger.Debug().Msg("getting url by short code")

	url, err := u.urlRepository.GetByShortCode(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url by short code")
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	logger.Debug().Int("id", url.Id).Msg("got url by short code")
	return url, nil
}

func (u *Url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("service_get_and_increment"))
	}()

	logger := log.With().
		Str("op", "service.url.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()

	// Validate short code
	if err := u.validateShortCode(shortCode); err != nil {
		logger.Debug().Err(err).Msg("short code validation failed")
		return nil, fmt.Errorf("invalid short code: %w", err)
	}

	logger.Debug().Msg("getting url by short code and incrementing clicks")

	url, err := u.urlRepository.GetByShortCodeAndIncrementClicks(ctx, shortCode)
	if err != nil {
		logger.Debug().Err(err).Msg("failed to get url and increment clicks")
		metrics.RecordURLRedirect("error")
		return nil, fmt.Errorf("failed to get url and increment clicks: %w", err)
	}

	logger.Debug().
		Int("id", url.Id).
		Int("clicks", url.Clicks).
		Msg("got url and incremented clicks")

	metrics.RecordURLClick(shortCode)
	metrics.RecordURLRedirect("success")
	return url, nil
}

func (u *Url) validateURL(longUrl string) error {
	if len(longUrl) < constant.MinURLLength {
		return fmt.Errorf("url too short: minimum length is %d", constant.MinURLLength)
	}

	if len(longUrl) > constant.MaxURLLength {
		return fmt.Errorf("url too long: maximum length is %d", constant.MaxURLLength)
	}

	// Additional URL validation could be added here
	// For now, we rely on the validator package in the handler layer

	return nil
}

func (u *Url) validateShortCode(shortCode string) error {
	if len(shortCode) == 0 {
		return errors.New("short code cannot be empty")
	}

	if len(shortCode) > constant.DefaultShortCodeLength*2 {
		return fmt.Errorf("short code too long: maximum length is %d", constant.DefaultShortCodeLength*2)
	}

	// Check if short code contains only valid characters
	for _, char := range shortCode {
		if !isValidShortCodeChar(char) {
			return fmt.Errorf("short code contains invalid character: %c", char)
		}
	}

	return nil
}

func isValidShortCodeChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

func (u *Url) generateShortCode(ctx context.Context) (string, error) {
	logger := log.With().
		Str("op", "service.url.generateShortCode").
		Logger()

	for attempt := 0; attempt < constant.MaxGenerationAttempts; attempt++ {
		logger.Debug().Int("attempt", attempt+1).Msg("generating short code")

		shortCode, err := u.generateRandomString(constant.DefaultShortCodeLength)
		if err != nil {
			logger.Error().Err(err).Int("attempt", attempt+1).Msg("failed to generate random string")
			continue
		}

		// Check if the short code already exists
		_, err = u.urlRepository.GetByShortCode(ctx, shortCode)
		if errors.Is(err, constant.ErrNotFound) {
			// Short code doesn't exist, we can use it
			logger.Debug().
				Str("short_code", shortCode).
				Int("attempt", attempt+1).
				Msg("generated unique short code")
			return shortCode, nil
		} else if err != nil {
			// Some other error occurred
			logger.Error().Err(err).Int("attempt", attempt+1).Msg("failed to check short code existence")
			continue
		}

		// Short code exists, try again
		logger.Debug().
			Str("short_code", shortCode).
			Int("attempt", attempt+1).
			Msg("short code already exists, trying again")
	}

	logger.Error().Int("max_attempts", constant.MaxGenerationAttempts).Msg("exhausted all attempts to generate unique short code")
	return "", fmt.Errorf("failed to generate unique short code after %d attempts", constant.MaxGenerationAttempts)
}

func (u *Url) generateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	charset := constant.ShortCodeCharset
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(bytes), nil
}

// GetStats returns service statistics
func (u *Url) GetStats(ctx context.Context) (*ServiceStats, error) {
	// This would typically query the repository for aggregated statistics
	// For now, we'll return a placeholder
	return &ServiceStats{
		TotalURLs:       0,
		TotalClicks:     0,
		LastUpdated:     time.Now(),
		AverageClicks:   0,
		TopShortCodes:   []string{},
		CreatedToday:    0,
		CreatedThisWeek: 0,
	}, nil
}

// ServiceStats represents service-level statistics
type ServiceStats struct {
	TotalURLs       int64     `json:"total_urls"`
	TotalClicks     int64     `json:"total_clicks"`
	LastUpdated     time.Time `json:"last_updated"`
	AverageClicks   float64   `json:"average_clicks"`
	TopShortCodes   []string  `json:"top_short_codes"`
	CreatedToday    int64     `json:"created_today"`
	CreatedThisWeek int64     `json:"created_this_week"`
}

// HealthCheck performs a health check on the service
func (u *Url) HealthCheck(ctx context.Context) error {
	// Test basic functionality by trying to generate a short code
	_, err := u.generateRandomString(constant.DefaultShortCodeLength)
	if err != nil {
		return fmt.Errorf("failed to generate random string: %w", err)
	}

	return nil
}
