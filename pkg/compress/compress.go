// Package compress provides an implementation for compressing and decompressing HTTP requests using gzip.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

// Writer implements the http.ResponseWriter interface and compresses responses using gzip.
type Writer struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter creates a new Writer for compressing HTTP responses.
func NewCompressWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the HTTP response headers.
func (c *Writer) Header() http.Header {
	return c.w.Header()
}

// Write compresses and writes data to the HTTP response.
func (c *Writer) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader writes an HTTP status code to the response.
// If the status code is less than 300, it sets the Content-Encoding: gzip header.
func (c *Writer) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close finalizes the gzip.Writer and releases resources.
func (c *Writer) Close() error {
	return c.zw.Close()
}

// Reader implements the io.ReadCloser interface and decompresses gzip-encoded data.
type Reader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader creates a new Reader for decompressing HTTP requests.
// Returns an error if initialization fails.
func NewCompressReader(r io.ReadCloser) (*Reader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &Reader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads and decompresses data from the stream.
func (c Reader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes the gzip.Reader and the underlying stream.
func (c *Reader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
