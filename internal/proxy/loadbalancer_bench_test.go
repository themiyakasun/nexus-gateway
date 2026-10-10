package proxy

import (
	"fmt"
	"testing"
)

func setupBenchmarkPool() (*ServerPool, *HashRing) {
	pool := &ServerPool{}
	ring := NewHashRing(50)

	for i := 1; i <= 5; i++ {
		url := fmt.Sprintf("http://localhost:808%d", i)
		upstream, _ := NewUpstream(url, nil)

		pool.AddUpstream(upstream)
		ring.AddUpstream(upstream)
	}

	return pool, ring
}

func BenchmarkRoundRobin(b *testing.B) {
	pool, _ := setupBenchmarkPool()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pool.GetNextBackend()
	}
}

func BenchmarkLeastConnection(b *testing.B) {
	pool, _ := setupBenchmarkPool()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pool.GetLeastConnectedBackend()
	}
}

func BenchmarkConsistentHashing(b *testing.B) {
	_, ring := setupBenchmarkPool()
	clientIP := "192.168.1.100"

	for i := 0; i < b.N; i++ {
		_ = ring.Get(clientIP)
	}
}

func BenchmarkRoundRobin_Parallel(b *testing.B) {
	pool, _ := setupBenchmarkPool()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pool.GetNextBackend()
		}
	})
}