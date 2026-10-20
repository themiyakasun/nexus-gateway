package proxy

import (
	"strings"
	"sync/atomic"
)

type RouteEntry struct {
	ID string
	PathPrefix string
	ServiceID string
	ServiceURL string
	StripPrefix bool
	AuthRequired bool
	Pool *ServerPool
}

type RoutingTable struct {
	routes []RouteEntry
}

func NewRoutingTable(routes []RouteEntry) *RoutingTable {
	return &RoutingTable{routes: routes}
}

func (rt *RoutingTable) Match(path string) (*RouteEntry, bool) {
	if rt == nil {
		return nil, false
	}

	var bestMatch *RouteEntry
	longestLen := 0

	for i := range rt.routes {
		route := &rt.routes[i]
		if strings.HasPrefix(path, route.PathPrefix) {
			if len(route.PathPrefix) > longestLen {
				longestLen = len(route.PathPrefix)
				bestMatch = route
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}
	return nil, false
}

type AtomicRouter struct {
	current atomic.Pointer[RoutingTable]
}

func NewAtomicRouter(initialTable *RoutingTable) *AtomicRouter {
	r := &AtomicRouter{}
	r.current.Store(initialTable)
	return r
}

func (r *AtomicRouter) Get() *RoutingTable {
	return r.current.Load()
}

func (r *AtomicRouter) Swap(newTable *RoutingTable) {
	r.current.Store(newTable)
}