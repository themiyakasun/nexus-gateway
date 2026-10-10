package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/themiyakasun/nexus-gateway/internal/filter"
)

func TestAPIKeyAuthMiddleware(t *testing.T) {
	db := NewMockDatabase()
	bf := filter.NewBloomFilter(1000, 3)
	bf.Add("sk_live_alice_123")

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := APIKeyAuthMiddleware(bf, db, dummyHandler)

	reqFake := httptest.NewRequest("GET", "/", nil)
	reqFake.Header.Set("Authorization", "Bearer sk_fake_random")
	recFake := httptest.NewRecorder()

	wrapped.ServeHTTP(recFake, reqFake)
	if recFake.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for fake key, got %d", recFake.Code)
	}

	reqValid := httptest.NewRequest("GET", "/", nil)
	reqValid.Header.Set("Authorization", "Bearer sk_live_alice_123")
	recValid := httptest.NewRecorder()

	wrapped.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid key, got %d", recValid.Code)
	}
}