package proxy

import (
	"sync/atomic"
	"testing"
)


func TestRoundRobin(t *testing.T){
	pool := &ServerPool{}
	serverA, _ := NewUpstream("http://localhost:8081", nil)
	serverB, _ := NewUpstream("http://localhost:8083", nil)

	pool.AddUpstream(serverA)
	pool.AddUpstream(serverB)

	first := pool.GetNextBackend()
	if first != serverA {
		t.Errorf("Expected serverA on first request, but got %v", first.URL)
	}

	second := pool.GetNextBackend()
	if second != serverB {
		t.Errorf("Expected serverB on second request, but got %v", second.URL)
	}

	third := pool.GetNextBackend()
	if third != serverA {
		t.Errorf("Expected serverA on third request (wrap around), but got %v", third.URL)
	}
}

func TestSkipDeadBackend(t *testing.T) {
	pool := &ServerPool{}
	serverA, _ := NewUpstream("http://localhost:8081", nil)
	serverB, _ := NewUpstream("http://localhost:8082", nil)

	pool.AddUpstream(serverA)
	pool.AddUpstream(serverB)

	serverA.SetAlive(false)

	selected := pool.GetNextBackend()

	if selected != serverB {
		t.Errorf("Expected serverB because serverA is dead, but got %v", selected)
	} 
}

func TestGetLeastConnectedBackend(t *testing.T) {
	pool := &ServerPool{}

	s1, _ := NewUpstream("http://localhost:8081", nil)
	s2, _ := NewUpstream("http://localhost:8082", nil)
	s3, _ := NewUpstream("http://localhost:8083", nil)

	atomic.StoreInt64(&s1.ActiveConnections, 5)
	atomic.StoreInt64(&s2.ActiveConnections, 1)
	atomic.StoreInt64(&s3.ActiveConnections, 3)

	pool.AddUpstream(s1)
	pool.AddUpstream(s2)
	pool.AddUpstream(s3)

	selected := pool.GetLeastConnectedBackend()

	if selected != s2 {
		t.Errorf("Expected server 2 (1 connection), but got %v with %d connections", selected.URL, selected.ActiveConnections)
	}
}

func TestLeastConnection_SkipsDeadBackend(t *testing.T) {
	pool := &ServerPool{}

	s1, _ := NewUpstream("http://localhost:8081", nil)
	s2, _ := NewUpstream("http://localhost:8082", nil)

	atomic.StoreInt64(&s1.ActiveConnections, 0)
	s1.SetAlive(false)

	atomic.StoreInt64(&s2.ActiveConnections, 4)
	s2.SetAlive(true)

	pool.AddUpstream(s1)
	pool.AddUpstream(s2)

	selected := pool.GetLeastConnectedBackend()

	if selected != s2 {
		t.Errorf("Expected Server 2 because Server 1 is dead, but got %v", selected.URL)
	}
}