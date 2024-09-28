package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zenfulhq/middleware"
)

func TestLogging(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := slog.New(slog.NewJSONHandler(buf, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandler := middleware.Logging(logger, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	type LogLine struct {
		Time     time.Time `json:"time"`
		Level    string    `json:"level"`
		Message  string    `json:"msg"`
		Method   string    `json:"method"`
		Path     string    `json:"path"`
		Duration int64     `json:"duration"`
		Status   int       `json:"status"`
	}

	var line LogLine

	assert.NoError(t, json.NewDecoder(buf).Decode(&line))

	assert.Equal(t, line.Level, "INFO")
	assert.Equal(t, line.Message, "request handled")
	assert.Equal(t, line.Method, http.MethodGet)
	assert.Equal(t, line.Path, "/")
	assert.Greater(t, line.Duration, int64(0))
	assert.Equal(t, line.Status, http.StatusOK)
}
