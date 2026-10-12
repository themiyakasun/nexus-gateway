package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
)

func extractAPIKey(r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return strings.TrimSpace(key)
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}

	return ""
}

func RedisAPIKeyMiddleware(rdb *redis.Client, setKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request)  {
		apiKey := extractAPIKey(r)
		if apiKey == "" {
			http.Error(w, "Unauthorized: Missing API Key (use X-API-Key or Authorization: Bearer)", http.StatusUnauthorized)
			return
		}

		isMember, err := rdb.SIsMember(r.Context(), setKey, apiKey).Result()
		if err != nil {
			log.Printf("[REDIS ERROR] Failed to check API key set: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !isMember {
			log.Printf("[AUTH BLOCKED] Invalid API Key: %s", apiKey)
			http.Error(w, "Unauthorized: Invalid API Key", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-Validated-API-Key", apiKey)

		next.ServeHTTP(w, r)
	})
}