package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/themiyakasun/nexus-gateway/config"
	"github.com/themiyakasun/nexus-gateway/internal/proxy"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handler, err := proxy.ReverseProxy("http://localhost:8081")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting proxy on :%d", cfg.Server.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), handler); err != nil {
		log.Fatal(err)
	}
}
