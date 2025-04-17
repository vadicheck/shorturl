package memory

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsumer(t *testing.T) {
	reader := bytes.NewReader([]byte(`{"Code":"abcd1234","URL":"https://example.com"}`))

	consumer, err := NewConsumer(reader)

	require.NoError(t, err)
	require.NotNil(t, consumer)
	assert.Equal(t, reader, *consumer.reader)
}

func TestReadURL_Success(t *testing.T) {
	data := `{"Code":"abcd1234","URL":"https://example.com"}`
	reader := bytes.NewReader([]byte(data))
	consumer, err := NewConsumer(reader)

	require.NoError(t, err)

	url, err := consumer.ReadURL()
	require.NoError(t, err)

	assert.Equal(t, "abcd1234", url.Code)
	assert.Equal(t, "https://example.com", url.URL)
}

func TestReadURL_Error(t *testing.T) {
	data := `{"Code":"abcd1234","URL":"https://example.com"`
	reader := bytes.NewReader([]byte(data))
	consumer, err := NewConsumer(reader)

	require.NoError(t, err)

	url, err := consumer.ReadURL()
	assert.Error(t, err)
	assert.Nil(t, url)
}

func TestLoad_Success(t *testing.T) {
	data := `{"Code":"abcd1234","URL":"https://example.com"}
	{"Code":"efgh5678","URL":"https://example.org"}`
	reader := bytes.NewReader([]byte(data))
	consumer, err := NewConsumer(reader)

	require.NoError(t, err)

	urls, err := consumer.Load()
	require.NoError(t, err)

	assert.Len(t, urls, 2)
	assert.Equal(t, "abcd1234", urls["abcd1234"].Code)
	assert.Equal(t, "https://example.com", urls["abcd1234"].URL)
	assert.Equal(t, "efgh5678", urls["efgh5678"].Code)
	assert.Equal(t, "https://example.org", urls["efgh5678"].URL)
}

func TestLoad_Error(t *testing.T) {
	data := `{"Code":"abcd1234","URL":"https://example.com"}
	{"Code":"efgh5678","URL":"https://example.org"`
	reader := bytes.NewReader([]byte(data))
	consumer, err := NewConsumer(reader)

	require.NoError(t, err)

	urls, err := consumer.Load()
	assert.Error(t, err)
	assert.Nil(t, urls)
}

func TestLoad_EOF(t *testing.T) {
	data := `{"Code":"abcd1234","URL":"https://example.com"}`
	reader := bytes.NewReader([]byte(data))
	consumer, err := NewConsumer(reader)

	require.NoError(t, err)

	urls, err := consumer.Load()
	require.NoError(t, err)

	assert.Len(t, urls, 1)
	assert.Equal(t, "abcd1234", urls["abcd1234"].Code)
	assert.Equal(t, "https://example.com", urls["abcd1234"].URL)
}
