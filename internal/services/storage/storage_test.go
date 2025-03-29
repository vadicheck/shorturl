package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExistsURLError_Error(t *testing.T) {
	err := &ExistsURLError{
		OriginalURL: "https://example.com",
		ShortCode:   "abcd1234",
		Err:         ErrURLOrCodeExists,
	}

	expectedErrorMessage := "[https://example.com:abcd1234] url or code exists"
	assert.Equal(t, expectedErrorMessage, err.Error())
}

func TestErrURLOrCodeExists(t *testing.T) {
	assert.Equal(t, "url or code exists", ErrURLOrCodeExists.Error())
}
