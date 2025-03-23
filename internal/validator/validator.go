package validator

import (
	"github.com/vadicheck/shorturl/internal/models/shorten"

	"github.com/gobuffalo/validate"
)

type Validator struct{}

type CreateBatchURLValidator interface {
	CreateBatchShortURL(request *[]shorten.CreateBatchURLRequest) *validate.Errors
}

type DeleteURLsValidator interface {
	DeleteShortURLs(request *[]string) *validate.Errors
}

func (v *Validator) CreateBatchShortURL(data *[]shorten.CreateBatchURLRequest) *validate.Errors {
	var checks []validate.Validator

	checks = append(checks, &ArrayNotEmpty[shorten.CreateBatchURLRequest]{
		Name:  "Data",
		Array: *data,
	})

	errors := validate.Validate(checks...)

	return errors
}

func (v *Validator) DeleteShortURLs(data *[]string) *validate.Errors {
	var checks []validate.Validator

	checks = append(checks, &ArrayNotEmpty[string]{
		Name:  "Data",
		Array: *data,
	})

	errors := validate.Validate(checks...)

	return errors
}

func New() *Validator {
	return &Validator{}
}
