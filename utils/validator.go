package utils

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the default validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewValidator initializes a new validator
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}
