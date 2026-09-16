package main

import (
	"log"
	"net/http"

	"github.com/themiyakasun/nexus-gateway/internal/proxy"
)

func main() {
	handler, err := proxy.ReverseProxy("http://localhost:8081")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting proxy on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
