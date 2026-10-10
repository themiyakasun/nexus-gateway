package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from Nexus Gateway! This is a long string that can be compressed"))
	})

	wrapped := GzipMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()

	wrapped.ServeHTTP(recorder, req)
	resp := recorder.Result()

	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %s", resp.Header.Get("Content-Encoding"))
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	unzippedContent, _ := io.ReadAll(gzReader)
	if string(unzippedContent) != "Hello from Nexus Gateway! This is a long string that can be compressed" {
		t.Errorf("Unzipped content mismatch: got %s", string(unzippedContent))
	}

	reqPlain := httptest.NewRequest("GET", "/", nil)
	recorderPlain := httptest.NewRecorder()

	wrapped.ServeHTTP(recorderPlain, reqPlain)
	respPlain := recorderPlain.Result()

	if respPlain.Header.Get("Content-Encoding") == "gzip" {
		t.Errorf("Did not expect Content-Encoding to be gzip for plain client")
	}
}