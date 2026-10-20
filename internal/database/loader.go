package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/themiyakasun/nexus-gateway/internal/cache"
	"github.com/themiyakasun/nexus-gateway/internal/proxy"
)

func RebuildRoutingTable(ctx context.Context, db *sql.DB, lruCache *cache.LRUCache) (*proxy.RoutingTable, error) {
	query := `
		SELECT r.id, r.path_prefix, r.strip_prefix, r.auth_required, s.id, s.url
		FROM routes r
		JOIN services s ON r.service_id = s.id
		WHERE s.is_active = true
		ORDER BY LENGTH(r.path_prefix) DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []proxy.RouteEntry

	for rows.Next(){
		var re proxy.RouteEntry
		if err := rows.Scan(&re.ID, &re.PathPrefix, &re.StripPrefix, &re.AuthRequired, &re.ServiceID, &re.ServiceURL); err != nil {
			return nil, err
		}

		pool := &proxy.ServerPool{
			Ring: proxy.NewHashRing(50),
			Cache: lruCache,
		}
		upstream, err := proxy.NewUpstream(re.ServiceURL, lruCache)
		if err != nil {
			log.Printf("Warning: failed to initialize upstream %s: %v", re.ServiceURL, err)
			continue
		}
		pool.Ring.AddUpstream(upstream)
		re.Pool = pool

		entries = append(entries, re)
	}

	return proxy.NewRoutingTable(entries), nil
	
}