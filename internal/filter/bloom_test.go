package filter

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(1000, 3)

	bf.Add("/products")
	bf.Add("/users")
	bf.Add("/health")

	if !bf.Contains("/products") {
		t.Errorf("Expected /products to be found!")
	}

	if !bf.Contains("/users") {
		t.Errorf("Expected /users to be found!")
	}

	if !bf.Contains("/health") {
		t.Errorf("Expected /health to be found!")
	}

	fakeRoutes := []string {
		"/malicious-script.php",
		"/admin/hack",
		"/fake-product-9999",
	}

	for _, fake := range fakeRoutes {
		if bf.Contains(fake) {
			t.Errorf("Did not expect %s to be in the Bloom Filter!", fake)
		}
	}
}