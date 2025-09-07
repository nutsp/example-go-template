package mocks

import (
	"context"

	"example-api-template/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockExampleService is a mock implementation of ExampleService
type MockExampleService struct {
	mock.Mock
}

// CreateExample mocks the CreateExample method
func (m *MockExampleService) CreateExample(ctx context.Context, name, email string, age int) (*domain.Example, error) {
	args := m.Called(ctx, name, email, age)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Example), args.Error(1)
}

// GetExampleByID mocks the GetExampleByID method
func (m *MockExampleService) GetExampleByID(ctx context.Context, id string) (*domain.Example, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Example), args.Error(1)
}

// GetExampleByEmail mocks the GetExampleByEmail method
func (m *MockExampleService) GetExampleByEmail(ctx context.Context, email string) (*domain.Example, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Example), args.Error(1)
}

// UpdateExample mocks the UpdateExample method
func (m *MockExampleService) UpdateExample(ctx context.Context, id, name, email string, age int) (*domain.Example, error) {
	args := m.Called(ctx, id, name, email, age)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Example), args.Error(1)
}

// DeleteExample mocks the DeleteExample method
func (m *MockExampleService) DeleteExample(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ListExamples mocks the ListExamples method
func (m *MockExampleService) ListExamples(ctx context.Context, limit, offset int) ([]*domain.Example, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Example), args.Int(1), args.Error(2)
}

// ValidateExampleBusinessRules mocks the ValidateExampleBusinessRules method
func (m *MockExampleService) ValidateExampleBusinessRules(ctx context.Context, name, email string, age int) error {
	args := m.Called(ctx, name, email, age)
	return args.Error(0)
}
