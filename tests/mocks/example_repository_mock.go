package mocks

import (
	"context"

	"example-api-template/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockExampleRepository is a mock implementation of ExampleRepository
type MockExampleRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockExampleRepository) Create(ctx context.Context, example *domain.Example) error {
	args := m.Called(ctx, example)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockExampleRepository) GetByID(ctx context.Context, id string) (*domain.Example, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Example), args.Error(1)
}

// GetByEmail mocks the GetByEmail method
func (m *MockExampleRepository) GetByEmail(ctx context.Context, email string) (*domain.Example, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Example), args.Error(1)
}

// Update mocks the Update method
func (m *MockExampleRepository) Update(ctx context.Context, example *domain.Example) error {
	args := m.Called(ctx, example)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockExampleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List mocks the List method
func (m *MockExampleRepository) List(ctx context.Context, limit, offset int) ([]*domain.Example, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Example), args.Error(1)
}

// Count mocks the Count method
func (m *MockExampleRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
