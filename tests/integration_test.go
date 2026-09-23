package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/themiyakasun/nexus-gateway/internal/proxy"
)

func TestEvenDistributionRequests(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, f *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"server": "backend-1"})
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, f *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"server": "backend-2"})
	}))
	defer backend2.Close()

	backend3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, f *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"server": "backend-3"})
	}))
	defer backend3.Close()

	pool := &proxy.ServerPool{}
	u1, _ := proxy.NewUpstream(backend1.URL)
	u2, _ := proxy.NewUpstream(backend2.URL)
	u3, _ := proxy.NewUpstream(backend3.URL)

	pool.AddUpstream(u1)
	pool.AddUpstream(u2)
	pool.AddUpstream(u3)

	gateway := httptest.NewServer(pool)
	defer gateway.Close()

	counts := make(map[string]int)
	totalRequests := 30

	for i := 0; i < totalRequests; i++ {
		res, err := http.Get(gateway.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}

		var data map[string]string
		json.NewDecoder(res.Body).Decode(&data)
		res.Body.Close()

		serverName := data["server"]
		counts[serverName]++
	}

	expectedPerSever := totalRequests / 3

	for server, count := range counts {
		t.Logf("Result: %s received %d requests", server, count)

		if count != expectedPerSever {
			t.Errorf("Server %s received %d requests, expected exactly %d", server, count, expectedPerSever)
		}
	}
}

