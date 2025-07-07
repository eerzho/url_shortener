package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"url_shortener/internal/constant"
	"url_shortener/internal/metrics"
	"url_shortener/internal/model"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Url struct {
	db *sqlx.DB
}

func NewUrl(db *sqlx.DB) *Url {
	return &Url{db: db}
}

func (u *Url) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	// Set timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, constant.DefaultDBTimeout)
		defer cancel()
	}

	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("postgres_create"))
	}()

	logger := log.With().
		Str("op", "repository.postgres.url.Create").
		Str("long_url", longUrl).
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("creating url in database")

	var url model.Url
	query := `
		INSERT INTO urls (long_url, short_code)
		VALUES ($1, $2)
		RETURNING id, short_code, long_url, clicks, created_at, updated_at
	`

	err := u.db.GetContext(ctx, &url, query, longUrl, shortCode)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create url in database")

		// Handle specific PostgreSQL errors
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation
				logger.Debug().Msg("short code already exists")
				metrics.RecordDatabaseOperation("create", "duplicate", timer.Duration())
				return nil, constant.ErrAlreadyExists
			case "23514": // check_violation
				logger.Debug().Msg("check constraint violation")
				metrics.RecordDatabaseOperation("create", "constraint_violation", timer.Duration())
				return nil, fmt.Errorf("constraint violation: %w", err)
			case "23502": // not_null_violation
				logger.Debug().Msg("not null constraint violation")
				metrics.RecordDatabaseOperation("create", "not_null_violation", timer.Duration())
				return nil, fmt.Errorf("required field is null: %w", err)
			default:
				logger.Error().Str("pg_code", string(pqErr.Code)).Msg("unexpected postgres error")
				metrics.RecordDatabaseOperation("create", "postgres_error", timer.Duration())
				return nil, fmt.Errorf("database error: %w", err)
			}
		}

		// Handle context errors
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error().Msg("database operation timed out")
			metrics.RecordDatabaseOperation("create", "timeout", timer.Duration())
			return nil, fmt.Errorf("database operation timed out: %w", err)
		}

		if errors.Is(err, context.Canceled) {
			logger.Debug().Msg("database operation cancelled")
			metrics.RecordDatabaseOperation("create", "cancelled", timer.Duration())
			return nil, fmt.Errorf("database operation cancelled: %w", err)
		}

		// Generic error
		metrics.RecordDatabaseOperation("create", "error", timer.Duration())
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	logger.Info().
		Int("id", url.Id).
		Str("short_code", url.ShortCode).
		Msg("url created successfully in database")

	metrics.RecordDatabaseOperation("create", "success", timer.Duration())
	return &url, nil
}

func (u *Url) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	// Set timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, constant.DefaultDBTimeout)
		defer cancel()
	}

	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("postgres_get"))
	}()

	logger := log.With().
		Str("op", "repository.postgres.url.GetByShortCode").
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("getting url from database")

	var url model.Url
	query := `
		SELECT id, short_code, long_url, clicks, created_at, updated_at
		FROM urls
		WHERE short_code = $1
	`

	err := u.db.GetContext(ctx, &url, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug().Msg("url not found in database")
			metrics.RecordDatabaseOperation("get", "not_found", timer.Duration())
			return nil, constant.ErrNotFound
		}

		// Handle context errors
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error().Msg("database operation timed out")
			metrics.RecordDatabaseOperation("get", "timeout", timer.Duration())
			return nil, fmt.Errorf("database operation timed out: %w", err)
		}

		if errors.Is(err, context.Canceled) {
			logger.Debug().Msg("database operation cancelled")
			metrics.RecordDatabaseOperation("get", "cancelled", timer.Duration())
			return nil, fmt.Errorf("database operation cancelled: %w", err)
		}

		logger.Error().Err(err).Msg("failed to get url from database")
		metrics.RecordDatabaseOperation("get", "error", timer.Duration())
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	logger.Debug().
		Int("id", url.Id).
		Int("clicks", url.Clicks).
		Msg("url retrieved successfully from database")

	metrics.RecordDatabaseOperation("get", "success", timer.Duration())
	return &url, nil
}

func (u *Url) GetByShortCodeAndIncrementClicks(ctx context.Context, shortCode string) (*model.Url, error) {
	// Set timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, constant.DefaultDBTimeout)
		defer cancel()
	}

	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("postgres_get_and_increment"))
	}()

	logger := log.With().
		Str("op", "repository.postgres.url.GetByShortCodeAndIncrementClicks").
		Str("short_code", shortCode).
		Logger()

	logger.Debug().Msg("getting url and incrementing clicks in database")

	var url model.Url
	query := `
		UPDATE urls
		SET clicks = clicks + 1, updated_at = CURRENT_TIMESTAMP
		WHERE short_code = $1
		RETURNING id, short_code, long_url, clicks, created_at, updated_at
	`

	err := u.db.GetContext(ctx, &url, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug().Msg("url not found in database")
			metrics.RecordDatabaseOperation("get_and_increment", "not_found", timer.Duration())
			return nil, constant.ErrNotFound
		}

		// Handle context errors
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error().Msg("database operation timed out")
			metrics.RecordDatabaseOperation("get_and_increment", "timeout", timer.Duration())
			return nil, fmt.Errorf("database operation timed out: %w", err)
		}

		if errors.Is(err, context.Canceled) {
			logger.Debug().Msg("database operation cancelled")
			metrics.RecordDatabaseOperation("get_and_increment", "cancelled", timer.Duration())
			return nil, fmt.Errorf("database operation cancelled: %w", err)
		}

		logger.Error().Err(err).Msg("failed to get url and increment clicks in database")
		metrics.RecordDatabaseOperation("get_and_increment", "error", timer.Duration())
		return nil, fmt.Errorf("failed to get url and increment clicks: %w", err)
	}

	logger.Debug().
		Int("id", url.Id).
		Int("clicks", url.Clicks).
		Msg("url retrieved and clicks incremented successfully in database")

	metrics.RecordDatabaseOperation("get_and_increment", "success", timer.Duration())
	return &url, nil
}

// GetStats returns database statistics
func (u *Url) GetStats(ctx context.Context) (*DatabaseStats, error) {
	// Set timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, constant.LongDBTimeout)
		defer cancel()
	}

	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("postgres_get_stats"))
	}()

	logger := log.With().
		Str("op", "repository.postgres.url.GetStats").
		Logger()

	logger.Debug().Msg("getting database statistics")

	var stats DatabaseStats
	query := `
		SELECT
			COUNT(*) as total_urls,
			COALESCE(SUM(clicks), 0) as total_clicks,
			COALESCE(AVG(clicks), 0) as average_clicks,
			COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) as created_today,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as created_this_week
		FROM urls
	`

	err := u.db.GetContext(ctx, &stats, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get database statistics")
		metrics.RecordDatabaseOperation("get_stats", "error", timer.Duration())
		return nil, fmt.Errorf("failed to get database statistics: %w", err)
	}

	stats.LastUpdated = time.Now()

	logger.Debug().
		Int64("total_urls", stats.TotalURLs).
		Int64("total_clicks", stats.TotalClicks).
		Msg("database statistics retrieved successfully")

	metrics.RecordDatabaseOperation("get_stats", "success", timer.Duration())
	return &stats, nil
}

// GetTopShortCodes returns the most clicked short codes
func (u *Url) GetTopShortCodes(ctx context.Context, limit int) ([]TopShortCode, error) {
	// Set timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, constant.DefaultDBTimeout)
		defer cancel()
	}

	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("postgres_get_top_short_codes"))
	}()

	logger := log.With().
		Str("op", "repository.postgres.url.GetTopShortCodes").
		Int("limit", limit).
		Logger()

	logger.Debug().Msg("getting top short codes from database")

	var topShortCodes []TopShortCode
	query := `
		SELECT short_code, long_url, clicks, created_at
		FROM urls
		ORDER BY clicks DESC, created_at DESC
		LIMIT $1
	`

	err := u.db.SelectContext(ctx, &topShortCodes, query, limit)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get top short codes from database")
		metrics.RecordDatabaseOperation("get_top_short_codes", "error", timer.Duration())
		return nil, fmt.Errorf("failed to get top short codes: %w", err)
	}

	logger.Debug().
		Int("count", len(topShortCodes)).
		Msg("top short codes retrieved successfully from database")

	metrics.RecordDatabaseOperation("get_top_short_codes", "success", timer.Duration())
	return topShortCodes, nil
}

// HealthCheck performs a health check on the database
func (u *Url) HealthCheck(ctx context.Context) error {
	// Set a short timeout for health checks
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger := log.With().
		Str("op", "repository.postgres.url.HealthCheck").
		Logger()

	logger.Debug().Msg("performing database health check")

	var result int
	query := `SELECT 1`

	err := u.db.GetContext(ctx, &result, query)
	if err != nil {
		logger.Error().Err(err).Msg("database health check failed")
		return fmt.Errorf("database health check failed: %w", err)
	}

	logger.Debug().Msg("database health check passed")
	return nil
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	TotalURLs       int64     `db:"total_urls" json:"total_urls"`
	TotalClicks     int64     `db:"total_clicks" json:"total_clicks"`
	AverageClicks   float64   `db:"average_clicks" json:"average_clicks"`
	CreatedToday    int64     `db:"created_today" json:"created_today"`
	CreatedThisWeek int64     `db:"created_this_week" json:"created_this_week"`
	LastUpdated     time.Time `json:"last_updated"`
}

// TopShortCode represents a top short code entry
type TopShortCode struct {
	ShortCode string    `db:"short_code" json:"short_code"`
	LongURL   string    `db:"long_url" json:"long_url"`
	Clicks    int       `db:"clicks" json:"clicks"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// DeleteOldURLs deletes URLs older than the specified duration
func (u *Url) DeleteOldURLs(ctx context.Context, olderThan time.Duration) (int64, error) {
	// Set timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, constant.LongDBTimeout)
		defer cancel()
	}

	timer := metrics.NewTimer()
	defer func() {
		timer.ObserveDuration(metrics.DatabaseOperationDuration.WithLabelValues("postgres_delete_old"))
	}()

	logger := log.With().
		Str("op", "repository.postgres.url.DeleteOldURLs").
		Dur("older_than", olderThan).
		Logger()

	logger.Debug().Msg("deleting old URLs from database")

	query := `
		DELETE FROM urls
		WHERE created_at < $1
		AND clicks = 0
	`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := u.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete old URLs from database")
		metrics.RecordDatabaseOperation("delete_old", "error", timer.Duration())
		return 0, fmt.Errorf("failed to delete old URLs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get rows affected count")
		metrics.RecordDatabaseOperation("delete_old", "error", timer.Duration())
		return 0, fmt.Errorf("failed to get rows affected count: %w", err)
	}

	logger.Info().
		Int64("rows_deleted", rowsAffected).
		Msg("old URLs deleted successfully from database")

	metrics.RecordDatabaseOperation("delete_old", "success", timer.Duration())
	return rowsAffected, nil
}
