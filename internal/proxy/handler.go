package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Upstream struct {
	URL *url.URL
	Alive bool
	ReverseProxy *httputil.ReverseProxy
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

func (s *ServerPool) AddUpstream(upstream *Upstream){
	s.backends = append(s.backends, upstream)
}

func (s *ServerPool) GetNextBackend() *Upstream {
	if len(s.backends) == 0 {
		return nil
	}

	next := atomic.AddUint64(&s.current, 1)

	index := int(next -1) % len(s.backends)

	return s.backends[index]
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

