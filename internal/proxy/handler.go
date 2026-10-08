package proxy

import (
	"bytes"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/themiyakasun/nexus-gateway/internal/cache"
)

type Upstream struct {
	URL *url.URL
	Alive bool
	ActiveConnections int64
	ReverseProxy *httputil.ReverseProxy
	mux sync.RWMutex
}

type ServerPool struct {
	backends []*Upstream
	current uint64
	Stratergy string
	Ring *HashRing
	Cache *cache.LRUCache
}

func NewUpstream(rawUrl string, c *cache.LRUCache) (*Upstream, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedUrl)

	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.Request.Method != http.MethodGet {
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			return nil
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		cacheKey := resp.Request.URL.Path
		c.Put(cacheKey, bodyBytes, 30*time.Second)

		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		return nil
	}

	return &Upstream{
		URL: parsedUrl,
		Alive: true,
		ReverseProxy: proxy,
	}, nil

}

func (u *Upstream) SetAlive(alive bool){
	u.mux.Lock()
	u.Alive = alive
	u.mux.Unlock()
}

func (u *Upstream) IsAlive() bool {
	u.mux.RLock()
	alive := u.Alive
	u.mux.RUnlock()
	return alive
}

func (s *ServerPool) GetLeastConnectedBackend() *Upstream {
	var bestBackend *Upstream
	var minConnections int64 = math.MaxInt64

	for _, backend := range s.backends {
		if !backend.IsAlive() {
			continue
		}

		conns := atomic.LoadInt64(&backend.ActiveConnections)
		if conns < minConnections {
			minConnections = conns
			bestBackend = backend
		}
	}
	return bestBackend
}

func (s *ServerPool) AddUpstream(upstream *Upstream){
	s.backends = append(s.backends, upstream)
}

func (s *ServerPool) GetNextBackend() *Upstream {
	totalBackends := len(s.backends)
	
	if totalBackends == 0 {
		return nil
	}

	for i := 0; i < totalBackends; i++ {
		next := atomic.AddUint64(&s.current, 1)

		index := int((next -1) % uint64(totalBackends))

		if s.backends[index].IsAlive() {
			return s.backends[index]
		}
	}

	return nil
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Cache != nil 	&& r.Method == http.MethodGet{
		cacheKey := r.URL.Path

		if cachedData, found := s.Cache.Get(cacheKey); found {
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(cachedData)
			return
		}
	}

	w.Header().Set("X-Cache", "MISS")


	var target *Upstream

	switch s.Stratergy {
	case "least-conn":
		target = s.GetLeastConnectedBackend()
	case "consistent-hash":
		clientIP := r.RemoteAddr
		target = s.Ring.Get(clientIP)
	default:
		target = s.GetNextBackend()
	}

	if target == nil {
		http.Error(w, "Service Unavailable: No backends configured", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&target.ActiveConnections, 1)

	defer atomic.AddInt64(&target.ActiveConnections, -1)

	target.ReverseProxy.ServeHTTP(w, r)
}

func ReverseProxy (target string) (http.Handler, error) {
	parsedUrl, err := url.Parse(target)

	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(parsedUrl), nil
}


func PingBackend(targetUrl *url.URL) bool {
	client := http.Client {
		Timeout: 2 * time.Second,
	}

	healthUrl := targetUrl.Scheme + "://" + targetUrl.Host + "/health"

	resp, err := client.Get(healthUrl)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (s *ServerPool) StartHealthCheck(interval time.Duration){
	ticker := time.NewTicker(interval)

	for range ticker.C {
		for _, upstream := range s.backends {
			alive := PingBackend(upstream.URL)

			if upstream.IsAlive() != alive {
				upstream.SetAlive(alive)

				status := "DOWN"


				if alive {
					status = "UP"
				}

				log.Printf("[HealthCheck] %s is %s", upstream.URL, status)
			}
		}
	}
}

