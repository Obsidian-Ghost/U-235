package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"strings"
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

// IsValidUUID User UUID Validator
func IsValidUUID(u string) bool {
	// Fast check for length and hyphens
	if len(u) != 36 || strings.Count(u, "-") != 4 {
		return false
	}

	// Parse using google/uuid package
	_, err := uuid.Parse(u)
	return err == nil
}
