package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"example-api-template/internal/errs"
	"example-api-template/internal/repository"
	"example-api-template/internal/service"
	"example-api-template/internal/usecase"
	"example-api-template/pkg/i18n"
	"example-api-template/pkg/validator"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ExampleHandler handles HTTP requests for examples
type ExampleHandler struct {
	useCase   usecase.ExampleUseCase
	validator validator.Validator
	logger    *zap.Logger
	localizer *i18n.Localizer
}

// NewExampleHandler creates a new example handler
func NewExampleHandler(
	useCase usecase.ExampleUseCase,
	validator validator.Validator,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) *ExampleHandler {
	return &ExampleHandler{
		useCase:   useCase,
		validator: validator,
		logger:    logger,
		localizer: localizer,
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

	response := FromExampleWithMetadata(example)

	return c.JSON(http.StatusCreated, response)
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
	logger := h.logger.With(
		zap.String("handler", "GetExample"),
		zap.String("id", c.Param("id")),
	)

	id := c.Param("id")
	if id == "" {
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(
		// 	errors.New("missing id"), "Example ID is required"))
	}

	example, err := h.useCase.GetExample(c.Request().Context(), id)
	if err != nil {
		return h.handleError(c, err, "Failed to get example", logger)
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
	logger := h.logger.With(
		zap.String("handler", "GetExampleByEmail"),
		zap.String("email", c.Param("email")),
	)

	email := c.Param("email")
	if email == "" {
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(
		// 	errors.New("missing email"), "Email is required"))
	}

	example, err := h.useCase.GetExampleByEmail(c.Request().Context(), email)
	if err != nil {
		return h.handleError(c, err, "Failed to get example by email", logger)
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
	logger := h.logger.With(
		zap.String("handler", "UpdateExample"),
		zap.String("id", c.Param("id")),
	)

	id := c.Param("id")
	if id == "" {
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(
		// 	errors.New("missing id"), "Example ID is required"))
	}

	var req UpdateExampleRequestDTO
	if err := c.Bind(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(err, "Invalid request format"))
	}

	// Validate request
	if validationErrors, _ := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		// logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		// return c.JSON(http.StatusUnprocessableEntity, NewValidationErrorResponse(
		// 	"Validation failed", validationErrors))
	}

	example, err := h.useCase.UpdateExample(c.Request().Context(), id, req.ToUpdateExampleRequest())
	if err != nil {
		return h.handleError(c, err, "Failed to update example", logger)
	}

	logger.Info("Example updated successfully", zap.String("id", example.ID))
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
	logger := h.logger.With(
		zap.String("handler", "DeleteExample"),
		zap.String("id", c.Param("id")),
	)

	id := c.Param("id")
	if id == "" {
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(
		// errors.New("missing id"), "Example ID is required"))
	}

	if err := h.useCase.DeleteExample(c.Request().Context(), id); err != nil {
		return h.handleError(c, err, "Failed to delete example", logger)
	}

	logger.Info("Example deleted successfully", zap.String("id", id))
	return c.JSON(http.StatusOK, NewSuccessResponse("Example deleted successfully"))
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
	logger := h.logger.With(
		zap.String("handler", "ListExamples"),
	)

	var req ListExamplesRequestDTO

	// Parse query parameters
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	// Validate request
	if validationErrors, _ := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		// message := h.localizer.LocalizeWithContext(c.Request().Context(), "common.validation_failed", nil)
		// return c.JSON(http.StatusBadRequest, NewValidationErrorResponse(
		// message, validationErrors))
	}

	response, err := h.useCase.ListExamples(c.Request().Context(), req.ToListExamplesRequest())
	if err != nil {
		return h.handleError(c, err, "Failed to list examples", logger)
	}

	logger.Info("Examples listed successfully",
		zap.Int("count", len(response.Examples)),
		zap.Int("total", response.Total))

	// message := h.localizer.LocalizeWithContext(c.Request().Context(), "example.list_retrieved", nil)

	listResponse := FromListExamplesResponse(response)
	// listResponse.Message = message

	return c.JSON(http.StatusOK, listResponse)
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
	logger := h.logger.With(
		zap.String("handler", "ValidateAndCreateExample"),
	)

	var req CreateExampleRequestDTO
	if err := c.Bind(&req); err != nil {
		// logger.Error("Failed to bind request", zap.Error(err))
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(err, "Invalid request format"))
	}

	// Validate request
	if validationErrors, _ := h.validator.ValidateStruct(&req); len(validationErrors) > 0 {
		logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		return c.JSON(http.StatusUnprocessableEntity, NewValidationErrorResponse(
			"Validation failed", validationErrors))
	}

	example, err := h.useCase.ValidateAndCreateExample(c.Request().Context(), req.ToCreateExampleRequest())
	if err != nil {
		return h.handleError(c, err, "Failed to validate and create example", logger)
	}

	logger.Info("Example validated and created successfully", zap.String("id", example.ID))
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

// handleError handles different types of errors and returns appropriate HTTP responses
func (h *ExampleHandler) handleError(c echo.Context, err error, message string, logger *zap.Logger) error {
	logger.Error(message, zap.Error(err))

	// Handle specific error types
	if errors.Is(err, repository.ErrExampleNotFound) {
		// return c.JSON(http.StatusNotFound, NewErrorResponse(err, "Example not found"))
	}

	if errors.Is(err, repository.ErrExampleAlreadyExists) {
		// return c.JSON(http.StatusConflict, NewErrorResponse(err, "Example already exists"))
	}

	if errors.Is(err, service.ErrInvalidInput) {
		// return c.JSON(http.StatusBadRequest, NewErrorResponse(err, "Invalid input"))
	}

	if errors.Is(err, service.ErrBusinessLogicFail) {
		// return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(err, "Business logic validation failed"))
	}

	if errors.Is(err, usecase.ErrUseCaseValidation) {
		// return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(err, "Use case validation failed"))
	}

	if errors.Is(err, usecase.ErrExternalService) {
		// return c.JSON(http.StatusBadGateway, NewErrorResponse(err, "External service error"))
	}

	// Handle external API specific errors
	if errors.Is(err, repository.ErrExternalAPIUnavailable) {
		// return c.JSON(http.StatusServiceUnavailable, NewErrorResponse(err, "External API unavailable"))
	}

	if errors.Is(err, repository.ErrExternalAPITimeout) {
		// return c.JSON(http.StatusGatewayTimeout, NewErrorResponse(err, "External API timeout"))
	}

	// Check for context errors
	if strings.Contains(err.Error(), "context deadline exceeded") {
		// return c.JSON(http.StatusGatewayTimeout, NewErrorResponse(err, "Request timeout"))
	}

	if strings.Contains(err.Error(), "context canceled") {
		// return c.JSON(http.StatusRequestTimeout, NewErrorResponse(err, "Request canceled"))
	}

	// Default to internal server error
	// return c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "Internal server error"))
	return errs.New(errs.ErrorCodeInternalError, err, nil)
}
