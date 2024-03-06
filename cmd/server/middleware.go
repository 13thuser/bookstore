package main

import (
	"context"
	"net/http"
)

// contextKey is the type for the context key
type contextKey string
type Middleware func(http.Handler) http.Handler

const (
	HEADER_AUTHORIZATION = "Authorization"
	TOKEN                = "token"
)

// authMiddleware is a middleware that checks for the presence of the Authorization header
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(HEADER_AUTHORIZATION)
		ctx := context.WithValue(r.Context(), contextKey(TOKEN), token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
