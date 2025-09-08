package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"example-api-template/internal/domain"
	"example-api-template/internal/repository"
	"example-api-template/internal/service"

	"go.uber.org/zap"
)

var (
	ErrUseCaseValidation = errors.New("use case validation failed")
	ErrExternalService   = errors.New("external service error")
)

// CreateExampleRequest represents the input for creating an example
type CreateExampleRequest struct {
	Name  string
	Email string
	Age   int
}

// UpdateExampleRequest represents the input for updating an example
type UpdateExampleRequest struct {
	Name  string
	Email string
	Age   int
}

// ExampleWithMetadata represents an example with additional metadata
type ExampleWithMetadata struct {
	*domain.Example
	ExternalData *repository.ExternalExampleData
	Enrichment   map[string]interface{}
}

// ListExamplesRequest represents pagination parameters
type ListExamplesRequest struct {
	Limit  int
	Offset int
}

// ListExamplesResponse represents the paginated response
type ListExamplesResponse struct {
	Examples []*ExampleWithMetadata
	Total    int
	Limit    int
	Offset   int
}

// ExampleUseCase defines the interface for example use cases
type ExampleUseCase interface {
	CreateExample(ctx context.Context, req CreateExampleRequest) (*ExampleWithMetadata, error)
	GetExample(ctx context.Context, id string) (*ExampleWithMetadata, error)
	GetExampleByEmail(ctx context.Context, email string) (*ExampleWithMetadata, error)
	UpdateExample(ctx context.Context, id string, req UpdateExampleRequest) (*ExampleWithMetadata, error)
	DeleteExample(ctx context.Context, id string) error
	ListExamples(ctx context.Context, req ListExamplesRequest) (*ListExamplesResponse, error)
	ValidateAndCreateExample(ctx context.Context, req CreateExampleRequest) (*ExampleWithMetadata, error)
}

// exampleUseCase implements ExampleUseCase
type exampleUseCase struct {
	service     service.ExampleService
	externalAPI repository.ExternalExampleAPI
	logger      *zap.Logger
	timeout     time.Duration
}

// NewExampleUseCase creates a new example use case
func NewExampleUseCase(
	service service.ExampleService,
	externalAPI repository.ExternalExampleAPI,
	logger *zap.Logger,
) ExampleUseCase {
	return &exampleUseCase{
		service:     service,
		externalAPI: externalAPI,
		logger:      logger,
		timeout:     30 * time.Second, // Default timeout for external API calls
	}
}

// CreateExample creates a new example with external validation
func (uc *exampleUseCase) CreateExample(ctx context.Context, req CreateExampleRequest) (*ExampleWithMetadata, error) {
	logger := uc.logger.With(
		zap.String("layer", "UseCase"),
		zap.String("operation", "CreateExample"),
		zap.String("email", req.Email),
		zap.String("name", req.Name),
		zap.Int("age", req.Age),
	)

	// Create example using service
	example, err := uc.service.CreateExample(ctx, req.Name, req.Email, req.Age)
	if err != nil {
		logger.Error("Service failed to create example", zap.Error(err))
		return nil, err
	}

	// Notify external API about new example creation (fire and forget)
	go func() {
		notifyCtx, cancel := context.WithTimeout(context.Background(), uc.timeout)
		defer cancel()

		if err := uc.externalAPI.NotifyExampleCreated(notifyCtx, example.ID, example.Email); err != nil {
			logger.Warn("Failed to notify external API", zap.Error(err))
		}
	}()

	// Return example with metadata
	return &ExampleWithMetadata{
		Example: example,
	}, nil
}

// GetExample retrieves an example with external data
func (uc *exampleUseCase) GetExample(ctx context.Context, id string) (*ExampleWithMetadata, error) {
	logger := uc.logger.With(
		zap.String("operation", "GetExample"),
		zap.String("id", id),
	)

	// Get example from service
	example, err := uc.service.GetExampleByID(ctx, id)
	if err != nil {
		logger.Error("Service failed to get example", zap.Error(err))
		return nil, err
	}

	// Enrich with external data
	return uc.enrichExample(ctx, example, logger)
}

// GetExampleByEmail retrieves an example by email with external data
func (uc *exampleUseCase) GetExampleByEmail(ctx context.Context, email string) (*ExampleWithMetadata, error) {
	logger := uc.logger.With(
		zap.String("operation", "GetExampleByEmail"),
		zap.String("email", email),
	)

	// Get example from service
	example, err := uc.service.GetExampleByEmail(ctx, email)
	if err != nil {
		logger.Error("Service failed to get example by email", zap.Error(err))
		return nil, err
	}

	// Enrich with external data
	return uc.enrichExample(ctx, example, logger)
}

// UpdateExample updates an example
func (uc *exampleUseCase) UpdateExample(ctx context.Context, id string, req UpdateExampleRequest) (*ExampleWithMetadata, error) {
	logger := uc.logger.With(
		zap.String("operation", "UpdateExample"),
		zap.String("id", id),
	)

	logger.Info("Updating example via use case")

	// Update example using service
	example, err := uc.service.UpdateExample(ctx, id, req.Name, req.Email, req.Age)
	if err != nil {
		logger.Error("Service failed to update example", zap.Error(err))
		return nil, err
	}

	// Enrich with external data
	return uc.enrichExample(ctx, example, logger)
}

// DeleteExample deletes an example
func (uc *exampleUseCase) DeleteExample(ctx context.Context, id string) error {
	logger := uc.logger.With(
		zap.String("operation", "DeleteExample"),
		zap.String("id", id),
	)

	logger.Info("Deleting example via use case")

	if err := uc.service.DeleteExample(ctx, id); err != nil {
		logger.Error("Service failed to delete example", zap.Error(err))
		return err
	}

	logger.Info("Example deleted successfully")
	return nil
}

// ListExamples retrieves a paginated list of examples with external data
func (uc *exampleUseCase) ListExamples(ctx context.Context, req ListExamplesRequest) (*ListExamplesResponse, error) {
	logger := uc.logger.With(
		zap.String("operation", "ListExamples"),
		zap.Int("limit", req.Limit),
		zap.Int("offset", req.Offset),
	)

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10 // Default limit
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}

	// Get examples from service
	examples, total, err := uc.service.ListExamples(ctx, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Service failed to list examples", zap.Error(err))
		return nil, err
	}

	// Enrich examples with external data (with timeout)
	enrichedExamples := make([]*ExampleWithMetadata, len(examples))
	for i, example := range examples {
		enriched, err := uc.enrichExample(ctx, example, logger)
		if err != nil {
			// Log error but continue with basic example data
			logger.Warn("Failed to enrich example", zap.String("id", example.ID), zap.Error(err))
			enriched = &ExampleWithMetadata{Example: example}
		}
		enrichedExamples[i] = enriched
	}

	return &ListExamplesResponse{
		Examples: enrichedExamples,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// ValidateAndCreateExample creates an example with external validation
func (uc *exampleUseCase) ValidateAndCreateExample(ctx context.Context, req CreateExampleRequest) (*ExampleWithMetadata, error) {
	logger := uc.logger.With(
		zap.String("operation", "ValidateAndCreateExample"),
		zap.String("email", req.Email),
	)

	logger.Info("Creating example with external validation")

	// Validate with external API first
	externalCtx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	isValid, err := uc.externalAPI.ValidateExample(externalCtx, req.Name, req.Email, req.Age)
	if err != nil {
		logger.Error("External validation failed",
			zap.String("name", req.Name),
			zap.String("email", req.Email),
			zap.Int("age", req.Age),
			zap.Error(err))
		return nil, fmt.Errorf("%w: external validation failed for user %s (%s): %v", ErrExternalService, req.Name, req.Email, err)
	}

	if !isValid {
		logger.Warn("External validation rejected example",
			zap.String("name", req.Name),
			zap.String("email", req.Email),
			zap.Int("age", req.Age))
		return nil, fmt.Errorf("%w: example %s (%s) rejected by external validation", ErrUseCaseValidation, req.Name, req.Email)
	}

	// Create example using service
	example, err := uc.service.CreateExample(ctx, req.Name, req.Email, req.Age)
	if err != nil {
		logger.Error("Service failed to create example", zap.Error(err))
		return nil, err
	}

	// Enrich with external data
	enriched, err := uc.enrichExample(ctx, example, logger)
	if err != nil {
		// Log error but return basic example
		logger.Warn("Failed to enrich created example", zap.Error(err))
		return &ExampleWithMetadata{Example: example}, nil
	}

	// Notify external API about new example creation (fire and forget)
	go func() {
		notifyCtx, cancel := context.WithTimeout(context.Background(), uc.timeout)
		defer cancel()

		if err := uc.externalAPI.NotifyExampleCreated(notifyCtx, example.ID, example.Email); err != nil {
			logger.Warn("Failed to notify external API", zap.Error(err))
		}
	}()

	return enriched, nil
}

// enrichExample enriches an example with external data
func (uc *exampleUseCase) enrichExample(ctx context.Context, example *domain.Example, logger *zap.Logger) (*ExampleWithMetadata, error) {
	enriched := &ExampleWithMetadata{
		Example: example,
	}

	// Create timeout context for external API calls
	externalCtx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// Use goroutines to parallelize external API calls
	var wg sync.WaitGroup
	var externalData *repository.ExternalExampleData
	var enrichmentData map[string]interface{}
	var extErr, enrichErr error

	wg.Add(2)

	// Get external data in parallel
	go func() {
		defer wg.Done()
		externalData, extErr = uc.externalAPI.GetExampleData(externalCtx, example.ID)
		if extErr != nil {
			logger.Warn("Failed to get external data", zap.String("id", example.ID), zap.Error(extErr))
		}
	}()

	// Get enrichment data in parallel
	go func() {
		defer wg.Done()
		enrichmentData, enrichErr = uc.externalAPI.EnrichExample(externalCtx, example.ID)
		if enrichErr != nil {
			logger.Warn("Failed to get enrichment data", zap.String("id", example.ID), zap.Error(enrichErr))
		}
	}()

	// Wait for both calls to complete
	wg.Wait()

	// Set the data if successful
	if extErr == nil && externalData != nil {
		enriched.ExternalData = externalData
	}
	if enrichErr == nil && enrichmentData != nil {
		enriched.Enrichment = enrichmentData
	}

	return enriched, nil
}
