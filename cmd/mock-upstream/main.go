package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Response struct {
	NodeID string        `json:"node_id"`
	Hostname string 					`json:"hostname"`
	Timestamp string `json:"timestamp"`
	Method string `json:"method"`
	Path string `json:"path"`
	Headers map[string][]string `json:"headers"`
}

func createServer(serverId string, port string){
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"server": serverId,
		})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, err := os.Hostname()

		if err != nil {
			hostname = "unknown"
		}

		response := Response {
			NodeID: serverId,
			Hostname: hostname,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Method: r.Method,
			Path: r.URL.Path,
			Headers: r.Header,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("[%s] starting on http://localhost:%s", serverId, port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[%s] server error: %v", serverId, err)
	}

}

func main() {
	

	go createServer("server-1", "8081")
	go createServer("server-2", "8082")
	go createServer("server-3", "8083")

	log.Println("Mock upstream cluster running:")
	log.Println("  server-1 -> :8081")
	log.Println("  server-2 -> :8082")
	log.Println("  server-3 -> :8083")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down mock cluster...")
	
}