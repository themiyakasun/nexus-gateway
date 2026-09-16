package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReverseProxy (t *testing.T) {
 //Fake Backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Backend Response"))
	}))
	defer backend.Close()

	proxyHandler, err := ReverseProxy(backend.URL)
	if err != nil {
		t.Fatalf("Failed to create proxy: %v", err)
	}

	//Fake Request
	req := httptest.NewRequest("GET", "http://my-proxy.com/", nil)
	recorder := httptest.NewRecorder()

	proxyHandler.ServeHTTP(recorder, req)

	//Check Response
	resp := recorder.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Backend Response" {
		t.Errorf("Expected 'Backend Response', got: %s", string(body))
	} 
}