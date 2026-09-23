package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Upstream struct {
	URL *url.URL
	Alive bool
	ReverseProxy *httputil.ReverseProxy
	mux sync.RWMutex
}

type ServerPool struct {
	backends []*Upstream
	current uint64
}

func NewUpstream(rawUrl string) (*Upstream, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	return &Upstream{
		URL: parsedUrl,
		Alive: true,
		ReverseProxy: httputil.NewSingleHostReverseProxy(parsedUrl),
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
	target := s.GetNextBackend()

	if target == nil {
		http.Error(w, "Service Unavailable: No backends configured", http.StatusServiceUnavailable)
		return
	}

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

