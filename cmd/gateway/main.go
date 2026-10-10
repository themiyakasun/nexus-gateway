package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/themiyakasun/nexus-gateway/config"
	"github.com/themiyakasun/nexus-gateway/internal/cache"
	"github.com/themiyakasun/nexus-gateway/internal/filter"
	"github.com/themiyakasun/nexus-gateway/internal/middleware"
	"github.com/themiyakasun/nexus-gateway/internal/proxy"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	lruCache := cache.NewLRUCache(1000)
	validRoutes := filter.NewBloomFilter(10000, 3)

	validRoutes.Add("/")
	validRoutes.Add("/health")
	validRoutes.Add("/products")

	db := middleware.NewMockDatabase()

	apiKeyFilter := filter.NewBloomFilter(10000, 3)

	apiKeyFilter.Add("sk_live_alice_123")
	apiKeyFilter.Add("sk_live_bob_456")

	pool := &proxy.ServerPool{
		Stratergy: "consistent-hash",
		Ring: proxy.NewHashRing(50),
		Cache: lruCache,
	}


	for _, u := range cfg.Upstreams {
		upstream, err := proxy.NewUpstream(u.URL, lruCache)
		if err != nil {
			log.Fatalf("Invalid upstream URL %s: %v", u.URL, err)
		}

		pool.Ring.AddUpstream(upstream)
		log.Printf("Added upstream: %s", u.URL)
	}

	

	go pool.StartHealthCheck(5 * time.Second)

	hRouteCheck := middleware.BloomFilterMiddleware(validRoutes, pool)

	hAuthCheck := middleware.APIKeyAuthMiddleware(apiKeyFilter, db, hRouteCheck)

	finalHandler := middleware.GzipMiddleware(hAuthCheck)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Nexus gateway listening on %s...", addr)

	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
