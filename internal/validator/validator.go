// Package validator provides validation logic for different operations,
// specifically for batch URL creation and URL deletion. It uses the
// gobuffalo/validate package to validate data before processing it.
package validator

import (
	"github.com/gobuffalo/validate"
	"github.com/vadicheck/shorturl/internal/models/shorten"
)

// Validator is a struct that provides methods for validating different
// operations related to short URL creation and deletion.
type Validator struct{}

// CreateBatchURLValidator is an interface that defines the method for validating
// batch URL creation requests.
type CreateBatchURLValidator interface {
	// CreateBatchShortURL validates the request for creating a batch of short URLs.
	// It returns a validation error object if any validation errors occur.
	CreateBatchShortURL(request *[]shorten.CreateBatchURLRequest) *validate.Errors
}

// DeleteURLsValidator is an interface that defines the method for validating
// URL deletion requests.
type DeleteURLsValidator interface {
	// DeleteShortURLs validates the request to delete short URLs.
	// It returns a validation error object if any validation errors occur.
	DeleteShortURLs(request *[]string) *validate.Errors
}

// CreateBatchShortURL validates the data for batch URL creation. It checks
// that the data array is not empty. If the array is empty, it returns validation errors.
func (v *Validator) CreateBatchShortURL(data *[]shorten.CreateBatchURLRequest) *validate.Errors {
	var checks []validate.Validator

	// Add the ArrayNotEmpty validator for the data array.
	checks = append(checks, &ArrayNotEmpty[shorten.CreateBatchURLRequest]{
		Name:  "Data",
		Array: *data,
	})

	// Validate the checks and return any errors.
	errors := validate.Validate(checks...)

	return errors
}

// DeleteShortURLs validates the data for URL deletion. It checks that the data array
// (containing the list of URLs to delete) is not empty. If the array is empty,
// it returns validation errors.
func (v *Validator) DeleteShortURLs(data *[]string) *validate.Errors {
	var checks []validate.Validator

	// Add the ArrayNotEmpty validator for the data array.
	checks = append(checks, &ArrayNotEmpty[string]{
		Name:  "Data",
		Array: *data,
	})

	// Validate the checks and return any errors.
	errors := validate.Validate(checks...)

	return errors
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{}
}
