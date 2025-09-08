package errs

type ErrorCode string

const (
	// Domain errors
	ErrorCodeExampleNotFound      ErrorCode = "example_not_found"
	ErrorCodeExampleAlreadyExists ErrorCode = "example_already_exists"
	ErrorCodeInvalidID            ErrorCode = "invalid_id"
	ErrorCodeInvalidEmail         ErrorCode = "invalid_email"
	ErrorCodeInvalidAge           ErrorCode = "invalid_age"
	ErrorCodeInvalidName          ErrorCode = "invalid_name"
	ErrorCodeInvalidInput         ErrorCode = "invalid_input"

	// Business rule errors
	ErrorCodeBusinessLogicFail      ErrorCode = "business_logic_fail"
	ErrorCodeCorporateEmailUnderage ErrorCode = "corporate_email_underage"
	ErrorCodeVIPDomainUnderage      ErrorCode = "vip_domain_underage"
	ErrorCodeProfanityDetected      ErrorCode = "profanity_detected"

	// System errors
	ErrorCodeDatabaseError        ErrorCode = "database_error"
	ErrorCodeExternalAPIError     ErrorCode = "external_api_error"
	ErrorCodeValidationError      ErrorCode = "validation_error"
	ErrorCodeInternalError        ErrorCode = "internal_error"
	ErrorCodeUnauthorized         ErrorCode = "unauthorized"
	ErrorCodeForbidden            ErrorCode = "forbidden"
	ErrorCodeBadRequest           ErrorCode = "bad_request"
	ErrorCodeMethodNotAllowed     ErrorCode = "method_not_allowed"
	ErrorCodeUnsupportedMediaType ErrorCode = "unsupported_media_type"
	ErrorCodeTooManyRequests      ErrorCode = "too_many_requests"
	ErrorCodeServiceUnavailable   ErrorCode = "service_unavailable"

	// Common errors
	ErrorCodeInvalidRequest   ErrorCode = "invalid_request"
	ErrorCodeValidationFailed ErrorCode = "validation_failed"

	// Example errors
	ErrorCodeExampleIDRequired    ErrorCode = "example_id_required"
	ErrorCodeExampleEmailRequired ErrorCode = "example_email_required"
)
