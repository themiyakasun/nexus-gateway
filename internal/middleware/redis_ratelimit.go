package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/themiyakasun/nexus-gateway/internal/ratelimit"
)

func DistributedRateLimitMiddleware(limiter *ratelimit.RedisSlidingWindowLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]

		result, err := limiter.Allow(r.Context(), clientIP)
		if err != nil {
			log.Printf("[REDIS ERROR] Rate limiter failed, %v (Allowing traffic)", err)
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.Reset, 10))

		if !result.Allowed {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too Many Requests (Distrubuted limit exceeded)", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}