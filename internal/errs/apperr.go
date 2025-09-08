package errs

import (
	"context"
	"errors"
	"net/http"

	"example-api-template/pkg/i18n"
)

type AppError struct {
	Message      string
	Code         ErrorCode
	Details      interface{}
	HTTPStatus   int
	TemplateData map[string]interface{}
	Err          error
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.Err
}

// GetHTTPStatus returns the HTTP status code
func (e *AppError) GetHTTPStatus() int {
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}
	return getDefaultHTTPStatus(e.Code)
}

// Localize localizes the message using Localizer and lang
func (e *AppError) Localize(localizer *i18n.Localizer, lang string) *AppError {
	message := localizer.LocalizeError(lang, string(e.Code), e.TemplateData)
	return &AppError{
		Code:         e.Code,
		Err:          e.Err,
		Details:      e.Details,
		TemplateData: e.TemplateData,
		HTTPStatus:   e.HTTPStatus,
		Message:      message,
	}
}

// LocalizeWithContext uses lang from context
func (e *AppError) LocalizeWithContext(localizer *i18n.Localizer, ctx context.Context) *AppError {
	lang := localizer.GetLanguageFromContext(ctx)
	return e.Localize(localizer, lang)
}

// New creates a simple AppError
func New(code ErrorCode, err error, details interface{}) *AppError {
	if err == nil {
		err = errors.New(string(code))
	}

	return &AppError{
		Code:       code,
		Err:        err,
		Details:    details,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// NewWithTemplate creates AppError พร้อม template data
func NewWithTemplate(code ErrorCode, err error, details interface{}, templateData map[string]interface{}) *AppError {
	return &AppError{
		Code:         code,
		Err:          err,
		Details:      details,
		HTTPStatus:   getDefaultHTTPStatus(code),
		TemplateData: templateData,
	}
}

// NewLocalized สร้างและ localize ทันที
func NewLocalized(localizer *i18n.Localizer, lang string, code ErrorCode, templateData map[string]interface{}, err error, details interface{}) *AppError {
	appErr := NewWithTemplate(code, err, details, templateData)
	return appErr.Localize(localizer, lang)
}

// NewLocalizedWithContext สร้างและ localize จาก context
func NewLocalizedWithContext(localizer *i18n.Localizer, ctx context.Context, code ErrorCode, templateData map[string]interface{}, err error, details interface{}) *AppError {
	appErr := NewWithTemplate(code, err, details, templateData)
	return appErr.LocalizeWithContext(localizer, ctx)
}

// Mapping ErrorCode → HTTP Status
func getDefaultHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrorCodeExampleNotFound:
		return http.StatusNotFound
	case ErrorCodeExampleAlreadyExists:
		return http.StatusConflict
	case ErrorCodeInvalidID, ErrorCodeInvalidEmail, ErrorCodeInvalidAge, ErrorCodeInvalidName, ErrorCodeInvalidInput, ErrorCodeBadRequest, ErrorCodeInvalidRequest, ErrorCodeValidationFailed:
		return http.StatusBadRequest
	case ErrorCodeBusinessLogicFail, ErrorCodeCorporateEmailUnderage, ErrorCodeVIPDomainUnderage, ErrorCodeProfanityDetected:
		return http.StatusUnprocessableEntity
	case ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case ErrorCodeUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
	case ErrorCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrorCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeExternalAPIError:
		return http.StatusBadGateway
	case ErrorCodeDatabaseError, ErrorCodeInternalError, ErrorCodeValidationError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
