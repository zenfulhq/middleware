package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &wrappedWriter{
			ResponseWriter: w,
		}

		start := time.Now()
		next.ServeHTTP(ww, r)
		end := time.Since(start)

		logger.Info("request handled",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int64("duration", end.Nanoseconds()),
			slog.Int("status", ww.statusCode),
		)
	})
}
