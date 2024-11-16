package ratelimiter

import (
	"context"
	"log/slog"
	"net/http"
)

type Algorithm interface {
	IsAllowed(ctx context.Context, ClientID string) (bool, error)
}

type RateLimiter struct {
	algo   Algorithm
	logger *slog.Logger
}

func New(algo Algorithm, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		algo:   algo,
		logger: logger,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		allowed, err := rl.algo.IsAllowed(r.Context(), userID)
		if err != nil {
			rl.logger.Error("failed to perform check", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}