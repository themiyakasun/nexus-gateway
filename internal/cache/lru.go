package cache

import (
	"sync"
	"time"
)

type Node struct {
	key string
	value []byte
	expiresAt time.Time
	prev *Node
	next *Node
}

type LRUCache struct {
	capacity int
	items map[string]*Node
	head *Node
	tail *Node
	mux sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	head := &Node{}
	tail := &Node{}

	head.next = tail
	tail.prev = head

	return &LRUCache{
		capacity: capacity,
		items: make(map[string]*Node),
		head: head,
		tail: tail,
	}
}

func (c *LRUCache) addNode(node *Node) {
	node.prev = c.head
	node.next = c.head.next

	c.head.next.prev = node
	c.head.next = node
}

func (c *LRUCache) removeNode(node *Node) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (c *LRUCache) moveToHead(node *Node) {
	c.removeNode(node)
	c.addNode(node)
}

func (c *LRUCache) removeTail() *Node {
	oldest := c.tail.prev
	c.removeNode(oldest)
	return oldest
}

func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mux.Lock()
	defer c.mux.Unlock()

	node, exists := c.items[key]
	if !exists {
		return nil,false
	}

	if time.Now().After(node.expiresAt) {
		c.removeNode(node)
		delete(c.items, key)
		return nil, false
	}

	c.moveToHead(node)

	return node.value, true
}

func (c *LRUCache) Put(key string, value []byte, ttl time.Duration) {
	c.mux.Lock()
	defer c.mux.Unlock()

	expiresAt := time.Now().Add(ttl)

	if node, exists := c.items[key]; exists {
		node.value = value
		node.expiresAt = expiresAt
		c.moveToHead(node)
		return
	}

	newNode := &Node {
		key: key,
		value: value,
		expiresAt: expiresAt,
	}
	c.items[key] = newNode
	c.addNode(newNode)

	if len(c.items) > c.capacity {
		oldest := c.removeTail()
		delete(c.items, oldest.key)
	}
}

func (c *LRUCache) Len() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return len(c.items)
}