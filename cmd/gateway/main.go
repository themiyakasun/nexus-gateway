package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/themiyakasun/nexus-gateway/config"
	"github.com/themiyakasun/nexus-gateway/internal/admin"
	"github.com/themiyakasun/nexus-gateway/internal/cache"
	"github.com/themiyakasun/nexus-gateway/internal/database"
	"github.com/themiyakasun/nexus-gateway/internal/events"
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
	initialTable, err := database.RebuildRoutingTable(context.Background(), db, lruCache)
	if err != nil {
		log.Printf("Warning: failed initial routing table load: %v", err)
		initialTable = proxy.NewRoutingTable(nil)
	}

	atomicRouter := proxy.NewAtomicRouter(initialTable)

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

	boardcaster := events.NewBroadcaster(rdb)

	adminRepo := admin.NewRepository(db)
	adminHandler := admin.NewHandler(adminRepo, boardcaster)

	events.StartSubscriber(context.Background(), rdb, func(event events.ConfigEvent) {
		log.Printf("[CLUSTER SYNC] Received event: %s at %s", event.Type, event.Timestamp.Format(time.RFC3339))

		switch event.Type {
		case events.EventRouteCreated, events.EventRouteDeleted, events.EventServiceCreated, events.EventServiceDeleted:
			log.Println("[RCU UPDATE] Rebuilding routing table from database...")

			newTable, err := database.RebuildRoutingTable(context.Background(), db, lruCache)
			if err != nil {
				log.Printf("[RCU UPDATE ERROR] Failed to rebuild routing table: %v", err)
				return
			}

			atomicRouter.Swap(newTable)
			log.Println("[RCU UPDATE Atomically swapped to new active routing table")
		case events.EventAPIKeyCreated:
			var res admin.CreateAPIKeyResponse
			if err := json.Unmarshal(event.Payload, &res); err == nil {
				rdb.SAdd(context.Background(), "valid_api_keys", res.RawKey)
				log.Printf("[CLUSTER SYNC] API Key '%s' (%s) activated in Redis", res.APIKey.Name, res.APIKey.KeyPrefix)
			}
		}


	})

	rootMux := http.NewServeMux()
	adminHandler.RegisterRoutes(rootMux)

	rootMux.Handle("/", finalHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr: addr,
		Handler: rootMux,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func(){
		log.Printf("Nexus gateway listening on %s...", addr)
		if err := http.ListenAndServe(addr, rootMux); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("[SHUTDOWN] Signal received. Draining active connections...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("[SHUTDOWN ERROR] Forced shutdown: %v", err)
	}

	log.Println("[SHUTDOWN] All active connections drained successfully. Gateway stopped")
}
