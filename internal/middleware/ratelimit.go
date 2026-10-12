package middleware

import (
	"net/http"
	"strings"

	"github.com/themiyakasun/nexus-gateway/internal/ratelimit"
)

func RateLimitMiddleware(limiter *ratelimit.IPRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]

		bucket := limiter.GetBucket(clientIP)

		if !bucket.Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Too Many Requests (Rate limit exceeded)", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}