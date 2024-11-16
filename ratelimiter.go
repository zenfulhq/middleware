package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/zenfulhq/middleware/ratelimiter/algorithm"
)

type Algorithm interface {
	IsAllowed(ctx context.Context, ClientID string) (bool, error)
}

type RateLimiter struct {
	algo   Algorithm
	logger *slog.Logger
}

func NewRateLimiter(limit int64, period time.Duration) (*RateLimiter, error) {
	algo, err := algorithm.NewSlidingWindow(limit, period, nil)
	if err != nil {
		return nil, err
	}

	return &RateLimiter{
		algo: algo,
		//logger: logger,
	}, nil
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
			//rl.logger.Error("failed to perform check", slog.Any("error", err))
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
