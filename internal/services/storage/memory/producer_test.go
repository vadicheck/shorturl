package memory

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/models"
)

func TestNewProducer(t *testing.T) {
	var buf bytes.Buffer

	producer, err := NewProducer(&buf)

	require.NoError(t, err)
	require.NotNil(t, producer)
}

func TestWriteURL_Success(t *testing.T) {
	var buf bytes.Buffer
	producer, err := NewProducer(&buf)

	require.NoError(t, err)

	url := &models.URL{
		Code: "abcd1234",
		URL:  "https://example.com",
	}

	err = producer.WriteURL(url)
	require.NoError(t, err)

	expectedJSON := `{"id":0,"code":"abcd1234","url":"https://example.com","user_id":"","is_deleted":false}` + "\n"

	assert.Equal(t, expectedJSON, buf.String())
}

func TestWriteURL_Error(t *testing.T) {
	errorWriter := errorWriter{}
	producer, err := NewProducer(errorWriter)

	require.NoError(t, err)

	url := &models.URL{
		Code: "abcd1234",
		URL:  "https://example.com",
	}

	err = producer.WriteURL(url)
	assert.Error(t, err)
}

type errorWriter struct{}

func (errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
