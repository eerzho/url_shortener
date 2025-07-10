package service

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type Ip struct {
}

func NewIp() *Ip {
	return &Ip{}
}

func (i *Ip) GetIp(ctx context.Context, r *http.Request) string {
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
