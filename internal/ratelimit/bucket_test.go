package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucket_BurstAndRefill(t *testing.T) {
	tb := NewTokenBucket(3, 2)

	if !tb.Allow() {
		t.Fatalf("Request 1 should be allowed")
	}
	if !tb.Allow(){
		t.Fatalf("Request 2 should be allowed")
	}
	if !tb.Allow(){
		t.Fatalf("Request 3 should be allowed")
	}

	if tb.Allow(){
		t.Fatalf("Request 4 should be BLOCKED, but was allowed!")
	}

	time.Sleep(550 * time.Millisecond)

	if !tb.Allow() {
		t.Fatalf("Request 5 should be allowed after replenishment!")
	}
}