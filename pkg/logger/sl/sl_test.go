package sl

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErr(t *testing.T) {
	t.Run("Convert error to slog.Attr", func(t *testing.T) {
		err := errors.New("test error")
		expected := slog.Attr{
			Key:   "error",
			Value: slog.StringValue("test error"),
		}

		result := Err(err)
		assert.Equal(t, expected, result)
	})
}
