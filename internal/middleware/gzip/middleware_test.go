package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware_RequestDecompression(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte(`{"input":"test"}`))
	zw.Close()

	handler := New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error reading body: %v", err)
		}
		if string(body) != `{"input":"test"}` {
			t.Fatalf("unexpected request body: %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rec.Code)
	}
}

func TestGzipMiddleware_NoCompression(t *testing.T) {
	handler := New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("plain response"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	if ce := res.Header.Get("Content-Encoding"); ce != "" {
		t.Fatalf("expected no Content-Encoding, got %s", ce)
	}

	body, _ := io.ReadAll(res.Body)
	if string(body) != "plain response" {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestIsCompressibleContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"JSON type", "application/json", true},
		{"JSON with charset", "application/json; charset=utf-8", true},
		{"HTML type", "text/html", true},
		{"HTML with charset", "text/html; charset=utf-8", true},
		{"Plain text", "text/plain", false},
		{"XML type", "application/xml", false},
		{"Empty", "", false},
		{"Unrelated type", "image/png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCompressibleContentType(tt.contentType)
			if got != tt.want {
				t.Errorf("isCompressibleContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}
