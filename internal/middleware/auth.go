package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/themiyakasun/nexus-gateway/internal/filter"
)

type MockDatabase struct {
	validKeys map[string]string
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		validKeys: map[string]string {
			"sk_live_alice_123": "Alice Corporation",
			"sk_live_bob_456": "Bob industries",
		},
	}
}

func (db *MockDatabase) Query(apiKey string) (string, bool) {
	log.Printf("[SLOW DATABASE HIT] Searching disk for key: %s...", apiKey)
	customer, exists := db.validKeys[apiKey]
	return customer, exists
}

func APIKeyAuthMiddleware(bf *filter.BloomFilter, db *MockDatabase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")

		if !bf.Contains(apiKey) {
			log.Printf("[BLOOM FILTER BLOCKED] Fake key '%s' rejected instantly! (DB skipped)", apiKey)
			http.Error(w, "Invalid API Key (Blocked by Nexus Gateway)", http.StatusUnauthorized)
			return
		}

		customer, exists := db.Query(apiKey)
		if !exists {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

		log.Printf("[AUTHENTICATED] Welcome, %s!", customer)

		r.Header.Set("X-Customer-Name", customer)

		next.ServeHTTP(w, r)
	})
}