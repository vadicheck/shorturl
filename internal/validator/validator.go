package validator

import (
	"github.com/vadicheck/shorturl/internal/models/shorten"

	"github.com/gobuffalo/validate"
)

type validator struct{}

func (v *validator) CreateBatchShortURL(data *[]shorten.CreateBatchURLRequest) *validate.Errors {
	var checks []validate.Validator

	checks = append(checks, &ArrayNotEmpty[shorten.CreateBatchURLRequest]{
		Name:  "Data",
		Array: *data,
	})

	errors := validate.Validate(checks...)

	return errors
}

type CreateBatchURLValidator interface {
	CreateBatchShortURL(request *[]shorten.CreateBatchURLRequest) *validate.Errors
}

func New() CreateBatchURLValidator {
	return &validator{}
}
