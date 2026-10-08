package cache

import (
	"testing"
)

func TestLRUCache_Eviction(t *testing.T) {
	c := NewLRUCache(2)

	c.Put("pageA", []byte("Content A"))
	c.Put("pageB", []byte("Content B"))

	_, found := c.Get("pageA")
	if !found {
		t.Fatalf("Expected pageA to be found")
	}

	c.Put("pageC", []byte("Content C"))

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