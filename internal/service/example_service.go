package service

import (
	"context"
	"errors"
	"fmt"

	"example-api-template/internal/domain"
	"example-api-template/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrBusinessLogicFail = errors.New("business logic validation failed")
)

// ExampleService defines the interface for example business logic
type ExampleService interface {
	CreateExample(ctx context.Context, name, email string, age int) (*domain.Example, error)
	GetExampleByID(ctx context.Context, id string) (*domain.Example, error)
	GetExampleByEmail(ctx context.Context, email string) (*domain.Example, error)
	UpdateExample(ctx context.Context, id, name, email string, age int) (*domain.Example, error)
	DeleteExample(ctx context.Context, id string) error
	ListExamples(ctx context.Context, limit, offset int) ([]*domain.Example, int, error)
	ValidateExampleBusinessRules(ctx context.Context, name, email string, age int) error
}

// exampleService implements ExampleService
type exampleService struct {
	repo   repository.ExampleRepository
	logger *zap.Logger
}

// NewExampleService creates a new example service
func NewExampleService(repo repository.ExampleRepository, logger *zap.Logger) ExampleService {
	return &exampleService{
		repo:   repo,
		logger: logger,
	}
}

// CreateExample creates a new example with business logic validation
func (s *exampleService) CreateExample(ctx context.Context, name, email string, age int) (*domain.Example, error) {
	logger := s.logger.With(
		zap.String("operation", "CreateExample"),
		zap.String("email", email),
		zap.String("name", name),
		zap.Int("age", age),
	)

	logger.Info("Creating new example")

	// Business logic validation
	if err := s.ValidateExampleBusinessRules(ctx, name, email, age); err != nil {
		logger.Error("Business validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrBusinessLogicFail, err)
	}

	// Generate ID (in real app, might use UUID)
	id := generateExampleID(name, email)

	// Create domain entity
	example, err := domain.NewExample(id, name, email, age)
	if err != nil {
		logger.Error("Failed to create domain entity", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Check if example with same email already exists
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		logger.Error("Example with email already exists", zap.String("email", email))
		return nil, fmt.Errorf("%w: example with email %s already exists", ErrBusinessLogicFail, email)
	}

	// Save to repository
	if err := s.repo.Create(ctx, example); err != nil {
		logger.Error("Failed to save example", zap.Error(err))
		return nil, fmt.Errorf("failed to save example: %w", err)
	}

	logger.Info("Example created successfully", zap.String("id", example.ID))
	return example, nil
}

// GetExampleByID retrieves an example by ID
func (s *exampleService) GetExampleByID(ctx context.Context, id string) (*domain.Example, error) {
	logger := s.logger.With(
		zap.String("operation", "GetExampleByID"),
		zap.String("id", id),
	)

	if id == "" {
		return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	example, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get example", zap.Error(err))
		return nil, err
	}

	logger.Info("Example retrieved successfully")
	return example, nil
}

// GetExampleByEmail retrieves an example by email
func (s *exampleService) GetExampleByEmail(ctx context.Context, email string) (*domain.Example, error) {
	logger := s.logger.With(
		zap.String("operation", "GetExampleByEmail"),
		zap.String("email", email),
	)

	if email == "" {
		return nil, fmt.Errorf("%w: email cannot be empty", ErrInvalidInput)
	}

	example, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		logger.Error("Failed to get example by email", zap.Error(err))
		return nil, err
	}

	logger.Info("Example retrieved successfully by email")
	return example, nil
}

// UpdateExample updates an existing example
func (s *exampleService) UpdateExample(ctx context.Context, id, name, email string, age int) (*domain.Example, error) {
	logger := s.logger.With(
		zap.String("operation", "UpdateExample"),
		zap.String("id", id),
		zap.String("email", email),
	)

	logger.Info("Updating example")

	if id == "" {
		return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	// Business logic validation
	if err := s.ValidateExampleBusinessRules(ctx, name, email, age); err != nil {
		logger.Error("Business validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrBusinessLogicFail, err)
	}

	// Get existing example
	example, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get existing example", zap.Error(err))
		return nil, err
	}

	// Check if email is being changed and conflicts with another example
	if example.Email != email {
		if existing, err := s.repo.GetByEmail(ctx, email); err == nil && existing.ID != id {
			logger.Error("Email already in use by another example", zap.String("email", email))
			return nil, fmt.Errorf("%w: email %s is already in use", ErrBusinessLogicFail, email)
		}
	}

	// Update the domain entity
	if err := example.Update(name, email, age); err != nil {
		logger.Error("Failed to update domain entity", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Save to repository
	if err := s.repo.Update(ctx, example); err != nil {
		logger.Error("Failed to update example", zap.Error(err))
		return nil, fmt.Errorf("failed to update example: %w", err)
	}

	logger.Info("Example updated successfully")
	return example, nil
}

// DeleteExample deletes an example by ID
func (s *exampleService) DeleteExample(ctx context.Context, id string) error {
	logger := s.logger.With(
		zap.String("operation", "DeleteExample"),
		zap.String("id", id),
	)

	if id == "" {
		return fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	// Check if example exists before deletion
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		logger.Error("Example not found for deletion", zap.Error(err))
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete example", zap.Error(err))
		return fmt.Errorf("failed to delete example: %w", err)
	}

	logger.Info("Example deleted successfully")
	return nil
}

// ListExamples retrieves a paginated list of examples
func (s *exampleService) ListExamples(ctx context.Context, limit, offset int) ([]*domain.Example, int, error) {
	logger := s.logger.With(
		zap.String("operation", "ListExamples"),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Validate pagination parameters
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	if offset < 0 {
		offset = 0
	}

	examples, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to list examples", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list examples: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		logger.Error("Failed to count examples", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count examples: %w", err)
	}

	logger.Info("Examples listed successfully",
		zap.Int("count", len(examples)),
		zap.Int("total", total),
	)
	return examples, total, nil
}

// ValidateExampleBusinessRules validates business-specific rules
func (s *exampleService) ValidateExampleBusinessRules(ctx context.Context, name, email string, age int) error {
	// Business rule: No profanity in names
	if containsProfanity(name) {
		return errors.New("name contains inappropriate content")
	}

	// Business rule: Corporate emails have different age restrictions
	if isCorporateEmail(email) && age < 18 {
		return errors.New("corporate accounts require minimum age of 18")
	}

	// Business rule: VIP domains get special treatment
	if isVIPDomain(email) && age < 21 {
		return errors.New("VIP accounts require minimum age of 21")
	}

	return nil
}

// Helper functions for business logic

func generateExampleID(name, email string) string {
	// Simple ID generation - in real app, use UUID or similar
	return fmt.Sprintf("ex_%s_%d", email[:3], len(name))
}

func containsProfanity(name string) bool {
	// Simple profanity check - in real app, use proper filter
	profanity := []string{"badword1", "badword2"}
	for _, word := range profanity {
		if name == word {
			return true
		}
	}
	return false
}

func isCorporateEmail(email string) bool {
	// Check if email belongs to corporate domain
	corporateDomains := []string{"@corp.com", "@enterprise.com"}
	for _, domain := range corporateDomains {
		if len(email) >= len(domain) && email[len(email)-len(domain):] == domain {
			return true
		}
	}
	return false
}

func isVIPDomain(email string) bool {
	// Check if email belongs to VIP domain
	vipDomains := []string{"@vip.com", "@premium.com"}
	for _, domain := range vipDomains {
		if len(email) >= len(domain) && email[len(email)-len(domain):] == domain {
			return true
		}
	}
	return false
}
