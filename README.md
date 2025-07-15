# URL Shortener REST API

A URL shortener service built with Go as a template and learning project.

## Features

- URL shortening with collision-resistant short codes
- Click tracking and analytics
- Rate limiting with LRU cache
- Asynchronous click processing with worker pool
- Valkey caching for URL lookups
- PostgreSQL for persistent storage
- Swagger documentation

## Quick Start

1. Clone the repository
2. Copy environment variables:
   ```bash
   cp .env.example .env
   ```
3. Start the services:
   ```bash
   make build
   make up
   ```

The API will be available at `http://localhost:8082`

Swagger documentation: `http://localhost:8082/swagger/`

### In-Memory LRU Cache for Rate Limiting
In a real production environment with multiple pods, rate limiting should either use external storage to synchronize limits across instances, or be handled at the infrastructure level

### Custom Worker Pool for Click Processing
For production systems, this should be replaced with external message queues or cloud-native async solutions for better reliability, scalability, and observability.

### Why Custom Implementations?
Potomu chto mogu
