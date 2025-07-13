package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"url_shortener/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestIp_GetIp(t *testing.T) {
	ipService := service.NewIp()
	ctx := context.Background()

	tests := []struct {
		name    string
		request func() *http.Request
		want    string
	}{
		{
			name: "X-Forwarded-For single IP",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Forwarded-For", "127.0.0.2")
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.2",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Forwarded-For", "127.0.0.2, 127.0.0.3, 127.0.0.4")
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.2",
		},
		{
			name: "X-Forwarded-For with spaces",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Forwarded-For", "  127.0.0.2  , 127.0.0.3")
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.2",
		},
		{
			name: "X-Forwarded-For invalid IP, fallback to X-Real-IP",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Forwarded-For", "invalid-ip")
				r.Header.Set("X-Real-IP", "127.0.0.2")
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.2",
		},
		{
			name: "X-Real-IP only",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Real-IP", "127.0.0.2")
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.2",
		},
		{
			name: "X-Real-IP invalid, fallback to RemoteAddr",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Real-IP", "invalid-ip")
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.1",
		},
		{
			name: "RemoteAddr with port",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.RemoteAddr = "127.0.0.1:8080"
				return r
			},
			want: "127.0.0.1",
		},
		{
			name: "RemoteAddr without port",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.RemoteAddr = "127.0.0.1"
				return r
			},
			want: "127.0.0.1",
		},
		{
			name: "Priority order tes",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "http://example.com", nil)
				r.Header.Set("X-Forwarded-For", "127.0.0.2")
				r.Header.Set("X-Real-IP", "127.0.0.3")
				r.RemoteAddr = "127.0.0.1"
				return r
			},
			want: "127.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.request()
			ip := ipService.GetIp(ctx, r)
			assert.Equal(t, tt.want, ip)
		})
	}
}

func BenchmarkIp_GetIp_XForwardedFor(b *testing.B) {
	ipService := service.NewIp()
	ctx := context.Background()

	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.Header.Set("X-Forwarded-For", "127.0.0.2")
	r.RemoteAddr = "127.0.0.1:8080"

	b.ResetTimer()
	for b.Loop() {
		ipService.GetIp(ctx, r)
	}
}

func BenchmarkIp_GetIp_XForwardedForMultiple(b *testing.B) {
	ipService := service.NewIp()
	ctx := context.Background()

	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.Header.Set("X-Forwarded-For", "127.0.0.2, 127.0.0.3, 127.0.0.4, 127.0.0.5")
	r.RemoteAddr = "127.0.0.1:8080"

	b.ResetTimer()
	for b.Loop() {
		ipService.GetIp(ctx, r)
	}
}

func BenchmarkIp_GetIp_XRealIP(b *testing.B) {
	ipService := service.NewIp()
	ctx := context.Background()

	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.Header.Set("X-Real-IP", "127.0.0.2")
	r.RemoteAddr = "127.0.0.1:8080"

	b.ResetTimer()
	for b.Loop() {
		ipService.GetIp(ctx, r)
	}
}

func BenchmarkIp_GetIp_RemoteAddr(b *testing.B) {
	ipService := service.NewIp()
	ctx := context.Background()

	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.RemoteAddr = "127.0.0.1:8080"

	b.ResetTimer()
	for b.Loop() {
		ipService.GetIp(ctx, r)
	}
}

func BenchmarkIp_GetIp_InvalidFallback(b *testing.B) {
	ipService := service.NewIp()
	ctx := context.Background()

	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.Header.Set("X-Forwarded-For", "invalid-ip")
	r.Header.Set("X-Real-IP", "invalid-ip")
	r.RemoteAddr = "127.0.0.1:8080"

	b.ResetTimer()
	for b.Loop() {
		ipService.GetIp(ctx, r)
	}
}
