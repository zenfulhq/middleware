package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Limit struct {
	Exceeded  bool
	Limit     int
	Remaining int
	Reset     time.Duration
}

type Limiter interface {
	AddAndCheckLimit(r *http.Request) (Limit, error)
}

func RateLimit(limiter Limiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit, err := limiter.AddAndCheckLimit(r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Add("X-RateLimit-Limit", strconv.Itoa(limit.Limit))
			w.Header().Add("X-RateLimit-Remaining", strconv.Itoa(limit.Remaining))
			w.Header().Add("X-RateLimit-Reset", strconv.Itoa(int(math.Ceil(limit.Reset.Seconds()))))

			if limit.Exceeded {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type GenericLimiter struct {
	PropertyExtractor func(r *http.Request) string
}

func ByRemoteAdd(r *http.Request) string {
	return r.RemoteAddr
}

func ByXFF(trustedCount int) func(r *http.Request) string {
	return func(r *http.Request) string {
		xffs := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		return xffs[len(xffs)-(trustedCount+1)]
	}
}
