package http

import (
	"errors"
	"net/http"
	"strconv"

	"example-api-template/internal/errs"
	"example-api-template/internal/usecase"
	"example-api-template/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Constants for validation and limits
const (
	DefaultLimit = 10
	MaxLimit     = 100
	MinAge       = 0
	MaxAge       = 150
	MinNameLen   = 1
	MaxNameLen   = 100
)

// Error messages
const (
	ErrMsgMissingID = "missing id"
)

// ExampleHandler handles HTTP requests for examples
type ExampleHandler struct {
	useCase   usecase.ExampleUseCase
	validator validator.Validator
}

// NewExampleHandler creates a new example handler
func NewExampleHandler(
	useCase usecase.ExampleUseCase,
	validator validator.Validator,
) *ExampleHandler {
	return &ExampleHandler{
		useCase:   useCase,
		validator: validator,
	}
}

// RegisterRoutes registers all example routes
func (h *ExampleHandler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")

	// Example routes
	examples := api.Group("/examples")
	examples.POST("", h.CreateExample)
	examples.GET("", h.ListExamples)
	examples.GET("/:id", h.GetExample)
	examples.PUT("/:id", h.UpdateExample)
	examples.DELETE("/:id", h.DeleteExample)
	examples.GET("/email/:email", h.GetExampleByEmail)
	examples.POST("/validate", h.ValidateAndCreateExample)

	// Health check
	api.GET("/health", h.HealthCheck)
}

// CreateExample creates a new example
// @Summary Create a new example
// @Description Create a new example with the provided data
// @Tags examples
// @Accept json
// @Produce json
// @Param example body CreateExampleRequestDTO true "Example data"
// @Success 201 {object} ExampleResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 422 {object} ValidationErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples [post]
func (h *ExampleHandler) CreateExample(c echo.Context) error {
	var req CreateExampleRequestDTO
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.ErrorCodeInvalidRequest, err, nil)
	}

	// Validate request
	if validationErrors, err := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		return errs.New(errs.ErrorCodeValidationFailed, err, validationErrors)
	}

	// Call use case
	example, err := h.useCase.CreateExample(c.Request().Context(), req.ToCreateExampleRequest())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, FromExampleWithMetadata(example))
}

// GetExample retrieves an example by ID
// @Summary Get an example by ID
// @Description Get an example by its ID
// @Tags examples
// @Produce json
// @Param id path string true "Example ID"
// @Success 200 {object} ExampleResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 404 {object} ErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples/{id} [get]
func (h *ExampleHandler) GetExample(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return errs.New(errs.ErrorCodeExampleIDRequired, errors.New(ErrMsgMissingID), nil)
	}

	example, err := h.useCase.GetExample(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FromExampleWithMetadata(example))
}

// GetExampleByEmail retrieves an example by email
// @Summary Get an example by email
// @Description Get an example by its email address
// @Tags examples
// @Produce json
// @Param email path string true "Example email"
// @Success 200 {object} ExampleResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 404 {object} ErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples/email/{email} [get]
func (h *ExampleHandler) GetExampleByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return errs.New(errs.ErrorCodeExampleEmailRequired, errors.New("missing email"), nil)
	}

	example, err := h.useCase.GetExampleByEmail(c.Request().Context(), email)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FromExampleWithMetadata(example))
}

// UpdateExample updates an existing example
// @Summary Update an example
// @Description Update an existing example with the provided data
// @Tags examples
// @Accept json
// @Produce json
// @Param id path string true "Example ID"
// @Param example body UpdateExampleRequestDTO true "Updated example data"
// @Success 200 {object} ExampleResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 404 {object} ErrorResponseDTO
// @Failure 422 {object} ValidationErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples/{id} [put]
func (h *ExampleHandler) UpdateExample(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return errs.New(errs.ErrorCodeExampleIDRequired, errors.New(ErrMsgMissingID), nil)
	}

	var req UpdateExampleRequestDTO
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.ErrorCodeInvalidRequest, err, nil)
	}

	// Validate request
	if validationErrors, err := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		return errs.New(errs.ErrorCodeValidationFailed, err, validationErrors)
	}

	example, err := h.useCase.UpdateExample(c.Request().Context(), id, req.ToUpdateExampleRequest())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FromExampleWithMetadata(example))
}

// DeleteExample deletes an example
// @Summary Delete an example
// @Description Delete an example by its ID
// @Tags examples
// @Produce json
// @Param id path string true "Example ID"
// @Success 200 {object} SuccessResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 404 {object} ErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples/{id} [delete]
func (h *ExampleHandler) DeleteExample(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return errs.New(errs.ErrorCodeExampleIDRequired, errors.New(ErrMsgMissingID), nil)
	}

	if err := h.useCase.DeleteExample(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// ListExamples retrieves a paginated list of examples
// @Summary List examples
// @Description Get a paginated list of examples
// @Tags examples
// @Produce json
// @Param limit query int false "Number of examples to return (max 100)" default(10)
// @Param offset query int false "Number of examples to skip" default(0)
// @Success 200 {object} ListExamplesResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples [get]
func (h *ExampleHandler) ListExamples(c echo.Context) error {
	var req ListExamplesRequestDTO

	// Parse query parameters with proper error handling
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		} else {
			return errs.New(errs.ErrorCodeInvalidRequest,
				errors.New("invalid limit parameter"),
				map[string]string{"limit": "must be a valid integer"})
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		} else {
			return errs.New(errs.ErrorCodeInvalidRequest,
				errors.New("invalid offset parameter"),
				map[string]string{"offset": "must be a valid integer"})
		}
	}

	// Set defaults if not provided
	if req.Limit <= 0 {
		req.Limit = DefaultLimit
	}
	if req.Limit > MaxLimit {
		req.Limit = MaxLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Validate request
	if validationErrors, err := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		return errs.New(errs.ErrorCodeValidationFailed, err, validationErrors)
	}

	response, err := h.useCase.ListExamples(c.Request().Context(), req.ToListExamplesRequest())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FromListExamplesResponse(response))
}

// ValidateAndCreateExample creates an example with external validation
// @Summary Create an example with external validation
// @Description Create a new example with external API validation
// @Tags examples
// @Accept json
// @Produce json
// @Param example body CreateExampleRequestDTO true "Example data"
// @Success 201 {object} ExampleResponseDTO
// @Failure 400 {object} ErrorResponseDTO
// @Failure 422 {object} ValidationErrorResponseDTO
// @Failure 500 {object} ErrorResponseDTO
// @Router /api/v1/examples/validate [post]
func (h *ExampleHandler) ValidateAndCreateExample(c echo.Context) error {
	var req CreateExampleRequestDTO
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.ErrorCodeInvalidRequest, err, nil)
	}

	// Validate request
	if validationErrors, err := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		return errs.New(errs.ErrorCodeValidationFailed, err, validationErrors)
	}

	example, err := h.useCase.ValidateAndCreateExample(c.Request().Context(), req.ToCreateExampleRequest())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, FromExampleWithMetadata(example))
}

// HealthCheck returns the health status of the service
// @Summary Health check
// @Description Get the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponseDTO
// @Router /api/v1/health [get]
func (h *ExampleHandler) HealthCheck(c echo.Context) error {
	services := map[string]string{
		"database":     "healthy",
		"external_api": "healthy",
		"cache":        "not_configured",
	}

	response := NewHealthResponse("1.0.0", services)
	return c.JSON(http.StatusOK, response)
}
