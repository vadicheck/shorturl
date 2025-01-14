package validator

import (
	"strings"

	"github.com/gobuffalo/validate"
)

type ArrayNotEmpty[T any] struct {
	Name    string
	Array   []T
	Message string
}

func (v *ArrayNotEmpty[T]) IsValid(errors *validate.Errors) {
	lengthArray := len(v.Array)

	if v.Message == "" {
		v.Message = "Array is empty"
	}

	if lengthArray == 0 {
		errors.Add(strings.ToLower(v.Name), v.Message)
	}
}
