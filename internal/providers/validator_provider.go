package providers

import "github.com/go-playground/validator/v10"

// NewValidator provides a new instance of a Validator
func NewValidator() *validator.Validate {
	return validator.New()
}
