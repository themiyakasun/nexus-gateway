package proxy

import "testing"


func TestRoundRobin(t *testing.T){
	pool := &ServerPool{}
	serverA, _ := NewUpstream("http://localhost:8081")
	serverB, _ := NewUpstream("http://localhost:8083")

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
	serverA, _ := NewUpstream("http://localhost:8081")
	serverB, _ := NewUpstream("http://localhost:8082")

	pool.AddUpstream(serverA)
	pool.AddUpstream(serverB)

	serverA.SetAlive(false)

	selected := pool.GetNextBackend()

	if selected != serverB {
		t.Errorf("Expected serverB because serverA is dead, but got %v", selected)
	} 
}