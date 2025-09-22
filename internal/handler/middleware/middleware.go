package middleware

import "net/http"

func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func ChainFunc(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return Chain(h, middlewares...)
}
