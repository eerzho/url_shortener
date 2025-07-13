package service

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type IP struct {
}

func NewIP() *IP {
	return &IP{}
}

func (i *IP) GetIP(_ context.Context, r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}
