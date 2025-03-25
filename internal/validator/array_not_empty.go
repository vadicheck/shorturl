// Package validator provides validation logic for various types, including
// custom validations for array types. It utilizes the gobuffalo/validate package
// for managing validation errors.
package validator

import (
	"strings"

	"github.com/gobuffalo/validate"
)

// ArrayNotEmpty is a custom validator that checks whether a given array is empty.
// It implements the validation logic for the 'array' field to ensure that it
// contains at least one element. If the array is empty, an error will be added
// to the validation errors.
type ArrayNotEmpty[T any] struct {
	// Name is the name of the field being validated (e.g., "myArray").
	Name string

	// Array is the array to be validated.
	Array []T

	// Message is the error message to display if the array is empty.
	// If not set, a default message "Array is empty" is used.
	Message string
}

// IsValid checks if the array is empty. If the array is empty, it adds an error
// to the provided validation errors object. If the Message field is empty, the
// default message "Array is empty" is used.
func (v *ArrayNotEmpty[T]) IsValid(errors *validate.Errors) {
	lengthArray := len(v.Array)

	// Set the default message if none is provided.
	if v.Message == "" {
		v.Message = "Array is empty"
	}

	// If the array is empty, add an error to the validation errors.
	if lengthArray == 0 {
		errors.Add(strings.ToLower(v.Name), v.Message)
	}
}
