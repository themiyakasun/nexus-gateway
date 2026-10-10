package proxy

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

type HashRing struct {
	vnodes int
	keys []uint32
	ring map[uint32]*Upstream
	mux sync.RWMutex
}

func NewHashRing(vnodes int) *HashRing {
	return &HashRing {
		vnodes: vnodes,
		ring: make(map[uint32]*Upstream),
	}
}

func (h *HashRing) AddUpstream(upstream *Upstream){
	h.mux.Lock()
	defer h.mux.Unlock()

	for i := 0; i < h.vnodes; i++ {
		vnodeName := fmt.Sprintf("%s#%d", upstream.URL.String(), i)

		point := crc32.ChecksumIEEE([]byte(vnodeName))

		h.keys = append(h.keys, point)
		h.ring[point] = upstream
	}

	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

func (h *HashRing) RemoveUpstream(upstream *Upstream) {
	h.mux.Lock()
	defer h.mux.Unlock()

	newKeys := make([]uint32, 0, len(h.keys))
	for _, point := range h.keys {
		if h.ring[point] == upstream {
			delete(h.ring, point)
		}else {
			newKeys = append(newKeys, point)
		}
	}

	h.keys = newKeys
}

func (h *HashRing) Get(key string) *Upstream {
	h.mux.RLock()
	defer h.mux.RUnlock()

	if len(h.keys) == 0 {
		return nil
	}

	hash := crc32.ChecksumIEEE([]byte(key))

	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	if idx == len(h.keys) {
		idx = 0
	}

	totalPoints := len(h.keys)
	for i := 0; i < totalPoints; i++ {
		point := h.keys[(idx+i)%totalPoints]
		candidate := h.ring[point]

		if candidate.IsAlive() {
			return candidate
		}
	}

	return nil
}