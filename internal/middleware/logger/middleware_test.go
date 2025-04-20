package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	loggedHandler := New()(handler)

	req := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	rr := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("unexpected status code: got %d, want %d", status, http.StatusOK)
	}

	if body := rr.Body.String(); body != "OK" {
		t.Errorf("unexpected response body: got %q, want %q", body, "OK")
	}
}
