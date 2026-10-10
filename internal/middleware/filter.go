package middleware

import (
	"net/http"

	"github.com/themiyakasun/nexus-gateway/internal/filter"
)

func BloomFilterMiddleware(bf *filter.BloomFilter, next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		if !bf.Contains(r.URL.Path) {
			http.Error(w, "Not Found (Blocked by Nexus Gateway)", http.StatusNotFound)
		}

		next.ServeHTTP(w, r)
	})
}