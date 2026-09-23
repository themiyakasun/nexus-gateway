package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/themiyakasun/nexus-gateway/config"
	"github.com/themiyakasun/nexus-gateway/internal/proxy"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool := &proxy.ServerPool{}

	for _, u := range cfg.Upstreams {
		upstream, err := proxy.NewUpstream(u.URL)
		if err != nil {
			log.Fatalf("Invalid upstream URL %s: %v", u.URL, err)
		}
		pool.AddUpstream(upstream)
		log.Printf("Added upstream: %s", u.URL)
	}

	go pool.StartHealthCheck(5 * time.Second)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Nexus gateway listening on %s...", addr)

	if err := http.ListenAndServe(addr, pool); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
