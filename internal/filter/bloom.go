package filter

import (
	"fmt"
	"hash/crc32"
	"sync"
)

type BloomFilter struct {
	size int
	k int
	bits []byte
	mux sync.RWMutex
}

func NewBloomFilter(size int, k int) *BloomFilter {
	byteCount := (size + 7) / 8

	return &BloomFilter{
		size: size,
		k: k,
		bits: make([]byte, byteCount),
	}
}

func (b *BloomFilter) getHashIndices(item string) []int {
	indices := make([]int, b.k)

	for i := 0; i < b.k; i++ {
		salted := fmt.Sprintf("%d%s", i, item)
		hash := crc32.ChecksumIEEE([]byte(salted))

		indices[i] = int(hash % uint32(b.size))
	}

	return indices
}

func (b *BloomFilter) Add(item string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	indices := b.getHashIndices(item)
	for _, idx := range indices {
		byteIndex := idx / 8
		bitIndex := idx % 8

		b.bits[byteIndex] |= (1 << bitIndex)
	}
}

func (b *BloomFilter) Contains(item string) bool {
	b.mux.RLock()
	defer b.mux.RUnlock()

	indices := b.getHashIndices(item)
	for _, idx := range indices {
		byteIndex := idx / 8
		bitIndex := idx % 8

		if(b.bits[byteIndex] & (1 << bitIndex)) == 0 {
			return false
		}
	}

	return true
}