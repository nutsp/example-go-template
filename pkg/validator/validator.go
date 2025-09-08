package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationFieldErrorDTO represents a field validation error
type ValidationFieldErrorDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
}

// Validator wraps the go-playground validator with additional functionality
type Validator interface {
	ValidateStruct(s interface{}) ([]ValidationFieldErrorDTO, error)
	ValidateVar(field interface{}, tag string) error
	RegisterValidation(tag string, fn validator.Func) error
}

// customValidator implements the Validator interface
type customValidator struct {
	validator *validator.Validate
}

// New creates a new validator instance
func New() Validator {
	validate := validator.New()

	// Use JSON tag names for validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validations
	cv := &customValidator{validator: validate}
	cv.registerCustomValidations()

	return cv
}

// ValidateStruct validates a struct and returns validation errors
func (cv *customValidator) ValidateStruct(s interface{}) ([]ValidationFieldErrorDTO, error) {
	var validationErrors []ValidationFieldErrorDTO

	err := cv.validator.Struct(s)
	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, fe := range ve {
				validationErrors = append(validationErrors, ValidationFieldErrorDTO{
					Field:   fe.Field(),
					Message: cv.getErrorMessage(fe),
					Tag:     fe.Tag(),
					Value:   fmt.Sprintf("%v", fe.Value()),
				})
			}
		}
	}

	return validationErrors, err
}

// ValidateVar validates a single variable
func (cv *customValidator) ValidateVar(field interface{}, tag string) error {
	return cv.validator.Var(field, tag)
}

// RegisterValidation registers a custom validation function
func (cv *customValidator) RegisterValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// registerCustomValidations registers custom validation functions
func (cv *customValidator) registerCustomValidations() {
	// Register custom email validation (stricter than default)
	cv.validator.RegisterValidation("strict_email", validateStrictEmail)

	// Register name validation (no numbers, special chars)
	cv.validator.RegisterValidation("valid_name", validateName)

	// Register age validation (reasonable range)
	cv.validator.RegisterValidation("valid_age", validateAge)

	// Register no profanity validation
	cv.validator.RegisterValidation("no_profanity", validateNoProfanity)
}

// getErrorMessage returns a human-readable error message for validation errors
func (cv *customValidator) getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "strict_email":
		return fmt.Sprintf("%s must be a valid email address with proper domain", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", fe.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", fe.Field())
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", fe.Field())
	case "valid_name":
		return fmt.Sprintf("%s must contain only letters and spaces", fe.Field())
	case "valid_age":
		return fmt.Sprintf("%s must be between 0 and 150", fe.Field())
	case "no_profanity":
		return fmt.Sprintf("%s contains inappropriate content", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Custom validation functions

// validateStrictEmail validates email with stricter rules
func validateStrictEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Basic email validation
	if err := validator.New().Var(email, "email"); err != nil {
		return false
	}

	// Additional checks
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]

	// Check for common invalid domains
	invalidDomains := []string{
		"test.com",
		"example.com",
		"localhost",
		"temp.com",
	}

	for _, invalid := range invalidDomains {
		if domain == invalid {
			return false
		}
	}

	// Must have at least one dot in domain
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// validateName validates that name contains only letters and spaces
func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	// Allow only letters, spaces, hyphens, and apostrophes
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			char == ' ' || char == '-' || char == '\'') {
			return false
		}
	}

	// Name should not start or end with space
	if strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") {
		return false
	}

	// Name should not have multiple consecutive spaces
	if strings.Contains(name, "  ") {
		return false
	}

	return true
}

// validateAge validates age is in reasonable range
func validateAge(fl validator.FieldLevel) bool {
	age := fl.Field().Int()
	return age >= 0 && age <= 150
}

// validateNoProfanity validates that text doesn't contain profanity
func validateNoProfanity(fl validator.FieldLevel) bool {
	text := strings.ToLower(fl.Field().String())

	// Simple profanity filter - in real app, use proper service
	profanityWords := []string{
		"badword1",
		"badword2",
		"inappropriate",
		"offensive",
	}

	for _, word := range profanityWords {
		if strings.Contains(text, word) {
			return false
		}
	}

	return true
}

// Utility functions for common validations

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	validate := validator.New()
	return validate.Var(email, "required,email")
}

// ValidateStrictEmail validates email with strict rules
func ValidateStrictEmail(email string) error {
	validate := validator.New()
	validate.RegisterValidation("strict_email", validateStrictEmail)
	return validate.Var(email, "required,strict_email")
}

// ValidateName validates name format
func ValidateName(name string) error {
	validate := validator.New()
	validate.RegisterValidation("valid_name", validateName)
	return validate.Var(name, "required,min=1,max=100,valid_name")
}

// ValidateAge validates age range
func ValidateAge(age int) error {
	validate := validator.New()
	validate.RegisterValidation("valid_age", validateAge)
	return validate.Var(age, "required,valid_age")
}

// ValidateRequired validates that a field is not empty
func ValidateRequired(field interface{}) error {
	validate := validator.New()
	return validate.Var(field, "required")
}

// ValidateUUID validates UUID format
func ValidateUUID(uuid string) error {
	validate := validator.New()
	return validate.Var(uuid, "required,uuid")
}

// ValidateURL validates URL format
func ValidateURL(url string) error {
	validate := validator.New()
	return validate.Var(url, "required,url")
}

// ValidateRange validates that a number is within a range
func ValidateRange(value, min, max int) error {
	validate := validator.New()
	return validate.Var(value, fmt.Sprintf("gte=%d,lte=%d", min, max))
}

// ValidateOneOf validates that a value is one of the allowed values
func ValidateOneOf(value string, allowedValues []string) error {
	validate := validator.New()
	return validate.Var(value, fmt.Sprintf("oneof=%s", strings.Join(allowedValues, " ")))
}
