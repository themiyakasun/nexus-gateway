package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/themiyakasun/nexus-gateway/config"
	"github.com/themiyakasun/nexus-gateway/internal/cache"
	"github.com/themiyakasun/nexus-gateway/internal/database"
	"github.com/themiyakasun/nexus-gateway/internal/filter"
	"github.com/themiyakasun/nexus-gateway/internal/middleware"
	"github.com/themiyakasun/nexus-gateway/internal/proxy"
	"github.com/themiyakasun/nexus-gateway/internal/ratelimit"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	db, err := database.Connect(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	lruCache := cache.NewLRUCache(1000)
	validRoutes := filter.NewBloomFilter(10000, 3)

	validRoutes.Add("/")
	validRoutes.Add("/health")
	validRoutes.Add("/products")

	// db := middleware.NewMockDatabase()

	apiKeyFilter := filter.NewBloomFilter(10000, 3)

	apiKeyFilter.Add("sk_live_alice_123")
	apiKeyFilter.Add("sk_live_bob_456")

	ipLimiter := ratelimit.NewIPRateLimiter(5, 2)



	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	rdb.SAdd(ctx, "valid_api_keys", "sk_live_alice_123", "sk_live_bob_456")
	log.Println("Seeded Redis Set 'valid_api_keys' with active customer keys.")


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
	// Station 3: Redis Set API Key Auth (Validates X-API-Key & Bearer against Redis)
	hAPIKey := middleware.RedisAPIKeyMiddleware(rdb, "valid_api_keys", hRouteCheck)

	// hAuthCheck := middleware.APIKeyAuthMiddleware(apiKeyFilter, db, hAPIKey)

	disLimiter := ratelimit.NewRedisSlidingWindowLimiter(rdb, 1*time.Minute, 10)

	hDistRate := middleware.DistributedRateLimitMiddleware(disLimiter, hAPIKey)
	// Station 2A: Local In-Memory Burst Limiter (5 req/sec)
	hLocalRate := middleware.RateLimitMiddleware(ipLimiter, hDistRate)
	// Station 1: Gzip Compression on the outside
	finalHandler := middleware.GzipMiddleware(hLocalRate)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Nexus gateway listening on %s...", addr)

	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
