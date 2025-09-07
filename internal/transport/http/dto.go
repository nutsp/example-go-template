package http

import (
	"time"

	"example-api-template/internal/domain"
	"example-api-template/internal/usecase"
	"example-api-template/pkg/validator"
)

// CreateExampleRequestDTO represents the HTTP request for creating an example
type CreateExampleRequestDTO struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=0,max=150"`
}

// UpdateExampleRequestDTO represents the HTTP request for updating an example
type UpdateExampleRequestDTO struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=0,max=150"`
}

// ExampleResponseDTO represents the HTTP response for an example
type ExampleResponseDTO struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Email        string                  `json:"email"`
	Age          int                     `json:"age"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	ExternalData *ExternalExampleDataDTO `json:"external_data,omitempty"`
	Enrichment   map[string]interface{}  `json:"enrichment,omitempty"`
}

// ExternalExampleDataDTO represents external API data in HTTP response
type ExternalExampleDataDTO struct {
	ExternalID   string            `json:"external_id"`
	Metadata     map[string]string `json:"metadata"`
	Score        float64           `json:"score"`
	LastModified time.Time         `json:"last_modified"`
}

// ListExamplesRequestDTO represents the HTTP request for listing examples
type ListExamplesRequestDTO struct {
	Limit  int `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// ListExamplesResponseDTO represents the HTTP response for listing examples
type ListExamplesResponseDTO struct {
	Examples   []*ExampleResponseDTO `json:"examples"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	HasNext    bool                  `json:"has_next"`
	HasPrev    bool                  `json:"has_prev"`
	TotalPages int                   `json:"total_pages"`
}

// ErrorResponseDTO represents an error response
type ErrorResponseDTO struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationErrorResponseDTO represents a validation error response
type ValidationErrorResponseDTO struct {
	Error   string                              `json:"error"`
	Message string                              `json:"message"`
	Code    string                              `json:"code"`
	Fields  []validator.ValidationFieldErrorDTO `json:"fields"`
}

// SuccessResponseDTO represents a success response without data
type SuccessResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HealthResponseDTO represents the health check response
type HealthResponseDTO struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// Conversion functions from domain/usecase to DTO

// ToCreateExampleRequest converts DTO to usecase request
func (dto *CreateExampleRequestDTO) ToCreateExampleRequest() usecase.CreateExampleRequest {
	return usecase.CreateExampleRequest{
		Name:  dto.Name,
		Email: dto.Email,
		Age:   dto.Age,
	}
}

// ToUpdateExampleRequest converts DTO to usecase request
func (dto *UpdateExampleRequestDTO) ToUpdateExampleRequest() usecase.UpdateExampleRequest {
	return usecase.UpdateExampleRequest{
		Name:  dto.Name,
		Email: dto.Email,
		Age:   dto.Age,
	}
}

// ToListExamplesRequest converts DTO to usecase request
func (dto *ListExamplesRequestDTO) ToListExamplesRequest() usecase.ListExamplesRequest {
	limit := dto.Limit
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := dto.Offset
	if offset < 0 {
		offset = 0
	}

	return usecase.ListExamplesRequest{
		Limit:  limit,
		Offset: offset,
	}
}

// FromExampleWithMetadata converts usecase response to DTO
func FromExampleWithMetadata(example *usecase.ExampleWithMetadata) *ExampleResponseDTO {
	dto := &ExampleResponseDTO{
		ID:        example.ID,
		Name:      example.Name,
		Email:     example.Email,
		Age:       example.Age,
		CreatedAt: example.CreatedAt,
		UpdatedAt: example.UpdatedAt,
	}

	if example.ExternalData != nil {
		dto.ExternalData = &ExternalExampleDataDTO{
			ExternalID:   example.ExternalData.ExternalID,
			Metadata:     example.ExternalData.Metadata,
			Score:        example.ExternalData.Score,
			LastModified: example.ExternalData.LastModified,
		}
	}

	if example.Enrichment != nil {
		dto.Enrichment = example.Enrichment
	}

	return dto
}

// FromExample converts domain example to DTO (without external data)
func FromExample(example *domain.Example) *ExampleResponseDTO {
	return &ExampleResponseDTO{
		ID:        example.ID,
		Name:      example.Name,
		Email:     example.Email,
		Age:       example.Age,
		CreatedAt: example.CreatedAt,
		UpdatedAt: example.UpdatedAt,
	}
}

// FromListExamplesResponse converts usecase response to DTO
func FromListExamplesResponse(response *usecase.ListExamplesResponse) *ListExamplesResponseDTO {
	examples := make([]*ExampleResponseDTO, len(response.Examples))
	for i, example := range response.Examples {
		examples[i] = FromExampleWithMetadata(example)
	}

	// Calculate pagination info
	totalPages := (response.Total + response.Limit - 1) / response.Limit
	hasNext := response.Offset+response.Limit < response.Total
	hasPrev := response.Offset > 0

	return &ListExamplesResponseDTO{
		Examples:   examples,
		Total:      response.Total,
		Limit:      response.Limit,
		Offset:     response.Offset,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		TotalPages: totalPages,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, message string) *ErrorResponseDTO {
	return &ErrorResponseDTO{
		Error:   err.Error(),
		Message: message,
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(message string, fields []validator.ValidationFieldErrorDTO) *ValidationErrorResponseDTO {
	return &ValidationErrorResponseDTO{
		Error:   "validation_failed",
		Message: message,
		Code:    "VALIDATION_ERROR",
		Fields:  fields,
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string) *SuccessResponseDTO {
	return &SuccessResponseDTO{
		Success: true,
		Message: message,
	}
}

// NewHealthResponse creates a new health response
func NewHealthResponse(version string, services map[string]string) *HealthResponseDTO {
	return &HealthResponseDTO{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
		Services:  services,
	}
}
