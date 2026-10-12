package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTMiddleware(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	dummyHandler := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request)  {
		if r.Header.Get("X-User-ID") != "user_42" {
			t.Errorf("Expected X-User-ID 'user_42', got %s", r.Header.Get("X-User-ID"))
		}
		w.WriteHeader(http.StatusOK)
	})

	jwtValidator := JWTMiddleware(publicKey, dummyHandler)

	validClaims := &CustomClaims{
		UserID: "user_42",
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	tokenValid := jwt.NewWithClaims(jwt.SigningMethodRS256, validClaims)
	tokenStringValid, _ := tokenValid.SignedString(privateKey)

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("Authorization", "Bearer "+tokenStringValid)
	rec1 := httptest.NewRecorder()

	jwtValidator.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for valid token, got %d", rec1.Code)
	}

	expiredClaims := &CustomClaims{
		UserID: "user_42",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	tokenExpired := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
	tokenStringExpired, _ := tokenExpired.SignedString(privateKey)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStringExpired)
	rec2 := httptest.NewRecorder()

	jwtValidator.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for expired token, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("Authorization", "Bearer totally_fake_tampered_token")
	rec3 := httptest.NewRecorder()
	
	jwtValidator.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for fake token, got %d", rec3.Code)
	}
}