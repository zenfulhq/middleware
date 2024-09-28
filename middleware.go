package middleware

import "net/http"

type Middleware func(h http.Handler) http.Handler
