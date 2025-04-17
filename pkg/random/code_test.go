package random

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 0", 0},
		{"length 1", 1},
		{"length 10", 10},
		{"length 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomString(tt.length)

			require.NoError(t, err)
			assert.Equal(t, tt.length, len(result))
		})
	}
}
