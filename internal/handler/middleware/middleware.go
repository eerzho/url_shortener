package middleware

import (
	"net"
	"net/http"
	"strings"
)

func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func ChainFunc(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return Chain(h, middlewares...)
}

func getIp(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if realIp := r.Header.Get("X-Real-IP"); realIp != "" {
		if net.ParseIP(realIp) != nil {
			return realIp
		}
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}
