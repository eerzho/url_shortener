package constant

import "time"

const (
	// Short code generation
	DefaultShortCodeLength = 6
	MaxGenerationAttempts  = 10
	ShortCodeCharset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Database timeouts
	DefaultDBTimeout = 5 * time.Second
	LongDBTimeout    = 10 * time.Second

	// Cache configuration
	DefaultCacheTTL       = 24 * time.Hour
	DefaultCacheKeyPrefix = "url:"

	// Rate limiting
	DefaultRateLimit = 10
	DefaultBurst     = 20
	DefaultCacheSize = 10_000

	// Worker configuration
	DefaultWorkerCount = 3
	DefaultBufferSize  = 500

	// HTTP timeouts
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second
	DefaultIdleTimeout  = 60 * time.Second

	// Shutdown timeout
	DefaultShutdownTimeout = 30 * time.Second

	// Validation limits
	MinURLLength = 10
	MaxURLLength = 2048

	// Environment values
	EnvDev   = "dev"
	EnvStage = "stage"
	EnvProd  = "prod"

	// HTTP Status Messages
	RateLimitExceededMessage   = "Rate limit exceeded"
	InternalServerErrorMessage = "Internal server error"
	BadRequestMessage          = "Bad request"
	NotFoundMessage            = "Resource not found"
	ConflictMessage            = "Resource already exists"
)
