package validator

import (
	"testing"

	"github.com/vadicheck/shorturl/internal/models/shorten"
)

// TestCreateBatchShortURL tests the CreateBatchShortURL method of the Validator.
func TestCreateBatchShortURL(t *testing.T) {
	validator := New()

	validData := []shorten.CreateBatchURLRequest{
		{CorrelationID: "12345", OriginalURL: "http://example.com"},
	}
	errors := validator.CreateBatchShortURL(&validData)
	if errors.HasAny() {
		t.Errorf("expected no errors, but got: %v", errors)
	}

	invalidData := []shorten.CreateBatchURLRequest{}
	errors = validator.CreateBatchShortURL(&invalidData)
	if !errors.HasAny() {
		t.Errorf("expected errors for empty array, but got none")
	}
}

// TestDeleteShortURLs tests the DeleteShortURLs method of the Validator.
func TestDeleteShortURLs(t *testing.T) {
	validator := New()

	validData := []string{"shortUrl1", "shortUrl2"}
	errors := validator.DeleteShortURLs(&validData)
	if errors.HasAny() {
		t.Errorf("expected no errors, but got: %v", errors)
	}

	invalidData := []string{}
	errors = validator.DeleteShortURLs(&invalidData)
	if !errors.HasAny() {
		t.Errorf("expected errors for empty array, but got none")
	}
}
