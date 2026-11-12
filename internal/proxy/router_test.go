package proxy

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAtomicRouterConcurrentSwapsUnderLoad(t *testing.T) {
	// 1. Create two mock upstream backend servers returning 200 OK
	backendA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend-A"))
	}))
	defer backendA.Close()

	backendB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend-B"))
	}))
	defer backendB.Close()

	// Helper to build a fully configured ServerPool
	makePool := func(rawURL string) *ServerPool {
		p := &ServerPool{
			Stratergy: "consistent-hash",
			Ring:      NewHashRing(50),
		}
		up, err := NewUpstream(rawURL, nil)
		if err != nil {
			t.Fatalf("Failed to create upstream: %v", err)
		}
		p.AddUpstream(up)
		p.Ring.AddUpstream(up)
		return p
	}

	poolA := makePool(backendA.URL)
	poolB := makePool(backendB.URL)

	// 2. Initialize AtomicRouter with initial table pointing to Backend A
	initialTable := NewRoutingTable([]RouteEntry{
		{
			PathPrefix: "/api",
			Pool:       poolA,
		},
	})
	atomicRouter := NewAtomicRouter(initialTable)

	// Test coordination primitives
	var (
		stopTraffic   atomic.Bool
		totalRequests atomic.Int64
		non200Count   atomic.Int64
		wgWorkers     sync.WaitGroup
	)

	// 3. Spawn 10 concurrent worker goroutines flooding traffic
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		wgWorkers.Add(1)
		go func(workerID int) {
			defer wgWorkers.Done()
			for !stopTraffic.Load() {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
				req.RemoteAddr = "192.168.1.100:54321" // Ensure RemoteAddr is present for consistent hashing

				currentTable := atomicRouter.Get()
				route, ok := currentTable.Match(req.URL.Path)
				if !ok || route.Pool == nil {
					non200Count.Add(1)
					totalRequests.Add(1)
					continue
				}

				route.Pool.ServeHTTP(rec, req)

				totalRequests.Add(1)
				if rec.Code != http.StatusOK {
					non200Count.Add(1)
				}
			}
		}(i)
	}

	// 4. Simultaneously perform 100 routing table swaps in the background
	numSwaps := 100
	for i := 0; i < numSwaps; i++ {
		var activePool *ServerPool
		if i%2 == 0 {
			activePool = poolB
		} else {
			activePool = poolA
		}

		newTable := NewRoutingTable([]RouteEntry{
			{
				PathPrefix: "/api",
				Pool:       activePool,
			},
		})

		// RCU Atomic Swap!
		atomicRouter.Swap(newTable)

		// Brief sleep between swaps to simulate real-world updates
		time.Sleep(5 * time.Millisecond)
	}

	// 5. Signal workers to finish and wait for all in-flight requests to complete
	stopTraffic.Store(true)
	wgWorkers.Wait()

	t.Logf("Completed %d total requests across %d atomic route swaps", totalRequests.Load(), numSwaps)

	// 6. Verify 0 dropped connections / non-200 responses
	if non200 := non200Count.Load(); non200 != 0 {
		t.Fatalf("FAILED: Encountered %d non-200 / dropped requests during atomic swaps!", non200)
	}

	if total := totalRequests.Load(); total < 1000 {
		t.Errorf("Expected substantial load (>1000 requests), but only processed %d", total)
	}
}