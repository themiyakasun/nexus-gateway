package proxy

import (
	"fmt"
	"testing"
)

func TestConsistentHash_Stickiness(t *testing.T) {
	ring := NewHashRing(50)

	s1, _ := NewUpstream("http://localhost:8081", nil)
	s2, _ := NewUpstream("http://localhost:8082", nil)
	s3, _ := NewUpstream("http://localhost:8083", nil)

	ring.AddUpstream(s1)
	ring.AddUpstream(s2)
	ring.AddUpstream(s3)

	clientIP := "192.168.1.50"
	initialServer := ring.Get(clientIP)

	if initialServer == nil {
		t.Fatalf("Expected a server, got nil")
	}

	for i := 0; i < 100; i++ {
		target := ring.Get(clientIP)
		if target != initialServer {
			t.Fatalf("Inconsistent routing! Expected %s, got %s on iteration %d", initialServer.URL, target.URL, i)
		}
	}
}

func TestConsistentHash_Distribution(t *testing.T) {
	ring := NewHashRing(50)


	s1, _ := NewUpstream("http://localhost:8081", nil)
	s2, _ := NewUpstream("http://localhost:8082", nil)
	s3, _ := NewUpstream("http://localhost:8083", nil)

	ring.AddUpstream(s1)
	ring.AddUpstream(s2)
	ring.AddUpstream(s3)

	counts := make(map[string]int)
	totalUsers := 3000

	for i := 0; i < totalUsers; i++ {
		fakeIP := fmt.Sprintf("10.0.%d.%d", i/256, i%256)
		server := ring.Get(fakeIP)
		counts[server.URL.String()]++
	}

	for url, count := range counts {
		percentage := float64(count) / float64(totalUsers) * 100
		t.Logf("Server %s handled %d users (%.1f%%)", url, count, percentage)

		if count < 600 {
			t.Errorf("Distribution too uneven! Server %s only received %d users", url, count)
		}
	}
}

func TestConsistentHash_Failover(t *testing.T) {
	ring := NewHashRing(50)

		s1, _ := NewUpstream("http://localhost:8081", nil)
	 s2, _ := NewUpstream("http://localhost:8082", nil)

		ring.AddUpstream(s1)
		ring.AddUpstream(s2)

		clientIP := "192.168.1.10"
		firstChoice := ring.Get(clientIP)

		firstChoice.SetAlive(false)

		fallbackChoice := ring.Get(clientIP)

		if fallbackChoice == nil {
			t.Fatalf("Expected fallback server, got nil")
		}

		if fallbackChoice == firstChoice {
			t.Fatalf("Ring routed to dead server %s!", firstChoice.URL)
		}
}