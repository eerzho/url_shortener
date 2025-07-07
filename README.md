# URL Shortener Service

A production-ready URL shortener service built with Go, featuring high performance, scalability, and comprehensive monitoring.

## Features

- **High Performance**: Built with Go for excellent concurrent performance
- **Scalable Architecture**: Clean architecture with dependency injection
- **Caching**: Redis-compatible Valkey for fast lookups
- **Rate Limiting**: Configurable rate limiting with IP-based tracking
- **Monitoring**: Prometheus metrics and structured logging
- **Database**: PostgreSQL with optimized queries and indexes
- **API Documentation**: Swagger/OpenAPI documentation
- **Health Checks**: Built-in health check endpoints
- **Graceful Shutdown**: Proper resource cleanup on shutdown
- **Docker Support**: Containerized deployment with Docker Compose

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Handler  │    │     Service     │    │   Repository    │
│                 │    │                 │    │                 │
│  - Validation   │◄──►│ - Business Logic│◄──►│  - Data Access  │
│  - Middleware   │    │ - URL Generation│    │  - Caching      │
│  - Rate Limiting│    │ - Validation    │    │  - Database     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Prometheus    │    │     Logging     │    │   PostgreSQL    │
│    Metrics      │    │   (Structured)  │    │   + Valkey      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## API Endpoints

### Core Endpoints

- `POST /urls` - Create a short URL
- `GET /urls/{short_code}` - Get URL information
- `GET /urls/{short_code}/stats` - Get URL statistics
- `GET /{short_code}` - Redirect to original URL

### System Endpoints

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /swagger/` - API documentation

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd url_shortener
   ```

2. **Start the services**
   ```bash
   make up
   ```

3. **Run database migrations**
   ```bash
   make migrate-up
   ```

4. **Test the service**
   ```bash
   # Create a short URL
   curl -X POST http://localhost:8080/urls \
     -H "Content-Type: application/json" \
     -d '{"long_url": "https://example.com"}'

   # Use the returned short_code to test redirect
   curl -I http://localhost:8080/{short_code}
   ```

### Local Development

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start dependencies**
   ```bash
   docker-compose up -d postgres valkey
   ```

4. **Run migrations**
   ```bash
   make migrate-up
   ```

5. **Start the application**
   ```bash
   go run ./cmd/http
   ```

## Configuration

Configuration is handled through environment variables:

### Required Variables

```env
APP_NAME=url_shortener
APP_ENV=dev                    # dev, stage, prod
HTTP_PORT=8080

# Database
POSTGRES_DB=url_shortener
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_PORT=5432
POSTGRES_URL=postgres://user:password@postgres:5432/url_shortener?sslmode=disable

# Cache
VALKEY_PORT=6379
VALKEY_URL=valkey:6379
```

### Optional Variables

- `LOG_LEVEL`: Set logging level (debug, info, warn, error)
- `RATE_LIMIT`: Requests per second per IP (default: 10)
- `RATE_BURST`: Burst size for rate limiting (default: 20)
- `CACHE_TTL`: Cache TTL in seconds (default: 86400)

## API Usage

### Create Short URL

```bash
curl -X POST http://localhost:8080/urls \
  -H "Content-Type: application/json" \
  -d '{
    "long_url": "https://www.example.com/very/long/url/that/needs/shortening"
  }'
```

Response:
```json
{
  "id": 1,
  "short_code": "abc123",
  "long_url": "https://www.example.com/very/long/url/that/needs/shortening",
  "clicks": 0,
  "created_at": 1640995200,
  "updated_at": 1640995200
}
```

### Get URL Information

```bash
curl http://localhost:8080/urls/abc123
```

### Get URL Statistics

```bash
curl http://localhost:8080/urls/abc123/stats
```

### Redirect to Original URL

```bash
curl -I http://localhost:8080/abc123
```

## Monitoring

### Metrics

The service exposes Prometheus metrics at `/metrics`:

- HTTP request metrics (duration, count, status codes)
- Database operation metrics
- Cache hit/miss ratios
- Rate limiting metrics
- Worker queue metrics
- Error rates

### Logging

Structured logging with contextual information:

```json
{
  "level": "info",
  "timestamp": "2024-01-01T12:00:00Z",
  "message": "short URL created successfully",
  "request_id": "req_123",
  "short_code": "abc123",
  "long_url": "https://example.com",
  "user_agent": "curl/7.68.0"
}
```

### Health Checks

Check service health:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0"
}
```

## Development

### Project Structure

```
url_shortener/
├── cmd/
│   └── http/           # HTTP server entry point
├── internal/
│   ├── app/           # Dependency injection setup
│   ├── config/        # Configuration management
│   ├── constant/      # Application constants
│   ├── handler/       # HTTP handlers and middleware
│   ├── metrics/       # Prometheus metrics
│   ├── model/         # Data models
│   ├── repository/    # Data access layer
│   ├── service/       # Business logic layer
│   └── utils/         # Utility functions
├── migrations/        # Database migrations
├── docker/           # Docker configurations
├── docs/             # API documentation
└── docker-compose.yaml
```

### Adding New Features

1. **Database Changes**: Add migrations in `migrations/`
2. **API Endpoints**: Add handlers in `internal/handler/`
3. **Business Logic**: Add services in `internal/service/`
4. **Data Access**: Add repositories in `internal/repository/`
5. **Documentation**: Update Swagger annotations

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Generate swagger docs
make swag
```

## Database Schema

### URLs Table

```sql
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    clicks INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_urls_created_at ON urls(created_at);
CREATE INDEX idx_urls_clicks ON urls(clicks);
```

## Performance Considerations

- **Caching**: Valkey cache reduces database load
- **Connection Pooling**: Optimized database connections
- **Indexes**: Strategic database indexes for fast queries
- **Rate Limiting**: Prevents abuse and ensures fair usage
- **Async Processing**: Click counting happens asynchronously
- **Graceful Shutdown**: Ensures data consistency

## Security Features

- **Input Validation**: Comprehensive request validation
- **Rate Limiting**: IP-based rate limiting
- **SQL Injection Protection**: Parameterized queries
- **XSS Protection**: Security headers on responses
- **Private Network Protection**: Blocks localhost/private IPs

## Deployment

### Docker Compose (Production)

1. **Configure environment**
   ```bash
   cp .env.example .env
   # Update with production values
   ```

2. **Deploy services**
   ```bash
   docker-compose -f docker-compose.prod.yaml up -d
   ```

### Kubernetes

See `k8s/` directory for Kubernetes manifests.

### Environment-specific Configuration

- **Development**: Full logging, debug endpoints
- **Staging**: Warn-level logging, performance monitoring
- **Production**: Info-level logging, security headers, monitoring

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Check `POSTGRES_URL` configuration
   - Ensure PostgreSQL is running
   - Verify network connectivity

2. **Cache Issues**
   - Check `VALKEY_URL` configuration
   - Ensure Valkey is running
   - Monitor cache hit rates

3. **High Memory Usage**
   - Check rate limiter cache size
   - Monitor worker queue sizes
   - Review cache TTL settings

### Debug Mode

Enable debug logging:
```bash
APP_ENV=dev go run ./cmd/http
```

### Monitoring Queries

Check database performance:
```sql
-- Find slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- Check index usage
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE tablename = 'urls';
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Update documentation
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions or issues:
- Create an issue in the repository
- Check the documentation
- Review the logs for error details