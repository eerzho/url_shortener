package middleware

import (
	"context"
	"net/http"
)

type IPService interface {
	GetIP(ctx context.Context, r *http.Request) string
}
