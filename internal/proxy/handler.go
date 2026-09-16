package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)


func ReverseProxy (target string) (http.Handler, error) {
	parsedUrl, err := url.Parse(target)

	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(parsedUrl), nil
}