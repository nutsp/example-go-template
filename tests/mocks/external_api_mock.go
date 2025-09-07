package mocks

import (
	"context"

	"example-api-template/internal/repository"

	"github.com/stretchr/testify/mock"
)

// MockExternalExampleAPI is a mock implementation of ExternalExampleAPI
type MockExternalExampleAPI struct {
	mock.Mock
}

// GetExampleData mocks the GetExampleData method
func (m *MockExternalExampleAPI) GetExampleData(ctx context.Context, exampleID string) (*repository.ExternalExampleData, error) {
	args := m.Called(ctx, exampleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ExternalExampleData), args.Error(1)
}

// ValidateExample mocks the ValidateExample method
func (m *MockExternalExampleAPI) ValidateExample(ctx context.Context, name, email string, age int) (bool, error) {
	args := m.Called(ctx, name, email, age)
	return args.Bool(0), args.Error(1)
}

// EnrichExample mocks the EnrichExample method
func (m *MockExternalExampleAPI) EnrichExample(ctx context.Context, exampleID string) (map[string]interface{}, error) {
	args := m.Called(ctx, exampleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// NotifyExampleCreated mocks the NotifyExampleCreated method
func (m *MockExternalExampleAPI) NotifyExampleCreated(ctx context.Context, exampleID, email string) error {
	args := m.Called(ctx, exampleID, email)
	return args.Error(0)
}
