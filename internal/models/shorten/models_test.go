package shorten

import (
	"testing"
)

func TestNewError(t *testing.T) {
	errResponse := NewError("Test error")

	if errResponse.Error != "Test error" {
		t.Errorf("expected error message to be 'Test error', got '%s'", errResponse.Error)
	}
}
