package middleware

import (
	"context"
	"net/http"
)

type IpService interface {
	GetIp(ctx context.Context, r *http.Request) string
}
