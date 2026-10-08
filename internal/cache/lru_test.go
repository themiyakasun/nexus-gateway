package cache

import (
	"testing"
	"time"
)

func TestLRUCache_Eviction(t *testing.T) {
	c := NewLRUCache(2)

	c.Put("pageA", []byte("Content A"), time.Minute)
	c.Put("pageB", []byte("Content B"), time.Minute)

	_, found := c.Get("pageA")
	if !found {
		t.Fatalf("Expected pageA to be found")
	}

	c.Put("pageC", []byte("Content C"), time.Minute)

	_,foundB := c.Get("pageB")
	if foundB {
		t.Errorf("Expected pageB to be evicted, but it was found!")
	}

	_, foundA := c.Get("pageA")
	_, foundC := c.Get("pageC")
	if !foundA || !foundC {
		t.Errorf("Expected pageA and pageC to still be in cache")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	c := NewLRUCache(5)

	c.Put("liveScore", []byte("2 - 1"), 50*time.Millisecond)

	val, found := c.Get("liveScore")
	if !found || string(val) != "2 - 1" {
		t.Fatalf("Expected liveScore to be fresh and found immediately")
	}

	time.Sleep(60 * time.Millisecond)

	_, foundAfterExpiry := c.Get("liveScore")
	if foundAfterExpiry {
		t.Errorf("Expected liveScore to be expired, but it was still found in cache!")
	}
}