package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"
)

// MockResponseWriter is a mock implementation of the http.ResponseWriter interface for testing purposes.
type MockResponseWriter struct {
	HeaderMap  http.Header
	Body       *bytes.Buffer
	StatusCode int
}

func (m *MockResponseWriter) Header() http.Header {
	return m.HeaderMap
}

func (m *MockResponseWriter) Write(p []byte) (int, error) {
	return m.Body.Write(p)
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		HeaderMap: make(http.Header),
		Body:      &bytes.Buffer{},
	}
}

// TestWriter tests the Writer implementation for compressing HTTP responses.
func TestWriter(t *testing.T) {
	mockWriter := NewMockResponseWriter()
	writer := NewCompressWriter(mockWriter)

	writer.WriteHeader(http.StatusOK)

	data := []byte("Hello, World!")
	_, err := writer.Write(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockWriter.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip Content-Encoding header, but got %s", mockWriter.Header().Get("Content-Encoding"))
	}

	gzipReader, err := gzip.NewReader(mockWriter.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("failed to decompress data: %v", err)
	}

	if string(decompressedData) != string(data) {
		t.Errorf("expected decompressed data to be %s, got %s", data, decompressedData)
	}
}

// TestReader tests the Reader implementation for decompressing HTTP requests.
func TestReader(t *testing.T) {
	data := []byte("Compressed request data")
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(data)
	if err != nil {
		t.Fatalf("failed to write data to gzip writer: %v", err)
	}
	gzipWriter.Close()

	reader, err := NewCompressReader(io.NopCloser(&buf))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read decompressed data: %v", err)
	}

	if string(decompressedData) != string(data) {
		t.Errorf("expected decompressed data to be %s, got %s", data, decompressedData)
	}
}
