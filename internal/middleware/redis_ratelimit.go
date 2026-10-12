package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/themiyakasun/nexus-gateway/internal/ratelimit"
)

func DistributedRateLimitMiddleware(limiter *ratelimit.RedisSlidingWindowLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]

		allowed, err := limiter.Allow(r.Context(), clientIP)
		if err != nil {
			log.Printf("[REDIS ERROR] Rate limiter failed, %v (Allowing traffic)", err)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too Many Requests (Distrubuted limit exceeded)", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}