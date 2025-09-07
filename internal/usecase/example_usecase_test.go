package usecase

import (
	"context"
	"testing"
	"time"

	"example-api-template/internal/domain"
	"example-api-template/internal/repository"
	"example-api-template/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test fixtures for usecase tests
func validExample() *domain.Example {
	example, _ := domain.NewExample(
		"ex_test_123",
		"John Doe",
		"john.doe@example.com",
		30,
	)
	return example
}

func validExampleWithCustomData(id, name, email string, age int) *domain.Example {
	example, _ := domain.NewExample(id, name, email, age)
	return example
}

func multipleValidExamples() []*domain.Example {
	return []*domain.Example{
		validExampleWithCustomData("ex_001", "Alice Smith", "alice@example.com", 25),
		validExampleWithCustomData("ex_002", "Bob Johnson", "bob@example.com", 35),
		validExampleWithCustomData("ex_003", "Carol Williams", "carol@example.com", 28),
	}
}

func validCreateExampleRequest() CreateExampleRequest {
	return CreateExampleRequest{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
	}
}

func validUpdateExampleRequest() UpdateExampleRequest {
	return UpdateExampleRequest{
		Name:  "John Smith",
		Email: "john.smith@example.com",
		Age:   31,
	}
}

func validExternalExampleData() *repository.ExternalExampleData {
	return &repository.ExternalExampleData{
		ExternalID: "ext_test_123",
		Metadata: map[string]string{
			"source":    "test_api",
			"version":   "1.0",
			"processed": time.Now().Format(time.RFC3339),
		},
		Score:        0.85,
		LastModified: time.Now(),
	}
}

func validEnrichmentData() map[string]interface{} {
	return map[string]interface{}{
		"external_id":  "ext_test_123",
		"risk_score":   0.1,
		"verification": "completed",
		"location_data": map[string]string{
			"country": "US",
			"region":  "CA",
		},
		"preferences": map[string]bool{
			"marketing_emails": true,
			"notifications":    false,
		},
	}
}

func getTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}

func TestNewExampleUseCase(t *testing.T) {
	mockService := &mocks.MockExampleService{}
	mockExternalAPI := &mocks.MockExternalExampleAPI{}
	logger := zap.NewNop()

	useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

	assert.NotNil(t, useCase)
}

func TestExampleUseCase_CreateExample(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateExampleRequest
		setupService  func(*mocks.MockExampleService)
		setupExternal func(*mocks.MockExternalExampleAPI)
		wantErr       bool
		errContains   string
	}{
		{
			name:    "successful creation",
			request: validCreateExampleRequest(),
			setupService: func(m *mocks.MockExampleService) {
				example := validExample()
				m.On("CreateExample", mock.Anything, "John Doe", "john.doe@example.com", 30).
					Return(example, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// NotifyExampleCreated is called in goroutine, so we need to account for that
				m.On("NotifyExampleCreated", mock.Anything, mock.AnythingOfType("string"), "john.doe@example.com").
					Return(nil).Maybe() // Maybe because it's async
			},
			wantErr: false,
		},
		{
			name: "service creation fails",
			request: CreateExampleRequest{
				Name:  "Invalid User",
				Email: "invalid@example.com",
				Age:   25,
			},
			setupService: func(m *mocks.MockExampleService) {
				m.On("CreateExample", mock.Anything, "Invalid User", "invalid@example.com", 25).
					Return(nil, repository.ErrExampleAlreadyExists)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// No external API calls expected when service fails
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)
			tt.setupExternal(mockExternalAPI)

			ctx := getTestContext()
			result, err := useCase.CreateExample(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.Age, result.Age)
			}

			mockService.AssertExpectations(t)
			// Give some time for async notification to complete
			time.Sleep(10 * time.Millisecond)
			mockExternalAPI.AssertExpectations(t)
		})
	}
}

func TestExampleUseCase_GetExample(t *testing.T) {
	tests := []struct {
		name           string
		inputID        string
		setupService   func(*mocks.MockExampleService)
		setupExternal  func(*mocks.MockExternalExampleAPI)
		wantErr        bool
		errContains    string
		expectEnriched bool
	}{
		{
			name:    "successful get with enrichment",
			inputID: "test-id",
			setupService: func(m *mocks.MockExampleService) {
				example := validExampleWithCustomData("test-id", "John Doe", "john@example.com", 30)
				m.On("GetExampleByID", mock.Anything, "test-id").Return(example, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				externalData := validExternalExampleData()
				enrichment := validEnrichmentData()
				m.On("GetExampleData", mock.Anything, "test-id").Return(externalData, nil)
				m.On("EnrichExample", mock.Anything, "test-id").Return(enrichment, nil)
			},
			wantErr:        false,
			expectEnriched: true,
		},
		{
			name:    "successful get with partial enrichment failure",
			inputID: "test-id",
			setupService: func(m *mocks.MockExampleService) {
				example := validExampleWithCustomData("test-id", "John Doe", "john@example.com", 30)
				m.On("GetExampleByID", mock.Anything, "test-id").Return(example, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// External data fails, enrichment succeeds
				m.On("GetExampleData", mock.Anything, "test-id").
					Return(nil, repository.ErrExternalAPIUnavailable)
				enrichment := validEnrichmentData()
				m.On("EnrichExample", mock.Anything, "test-id").Return(enrichment, nil)
			},
			wantErr:        false,
			expectEnriched: false, // Only partial enrichment
		},
		{
			name:    "service fails",
			inputID: "non-existent",
			setupService: func(m *mocks.MockExampleService) {
				m.On("GetExampleByID", mock.Anything, "non-existent").
					Return(nil, repository.ErrExampleNotFound)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// No external calls expected when service fails
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)
			tt.setupExternal(mockExternalAPI)

			ctx := getTestContext()
			result, err := useCase.GetExample(ctx, tt.inputID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.inputID, result.ID)

				if tt.expectEnriched {
					assert.NotNil(t, result.ExternalData)
					assert.NotNil(t, result.Enrichment)
				}
			}

			mockService.AssertExpectations(t)
			mockExternalAPI.AssertExpectations(t)
		})
	}
}

func TestExampleUseCase_ValidateAndCreateExample(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateExampleRequest
		setupService  func(*mocks.MockExampleService)
		setupExternal func(*mocks.MockExternalExampleAPI)
		wantErr       bool
		errContains   string
	}{
		{
			name:    "successful validation and creation",
			request: validCreateExampleRequest(),
			setupService: func(m *mocks.MockExampleService) {
				example := validExample()
				m.On("CreateExample", mock.Anything, "John Doe", "john.doe@example.com", 30).
					Return(example, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// External validation succeeds
				m.On("ValidateExample", mock.Anything, "John Doe", "john.doe@example.com", 30).
					Return(true, nil)
				// Enrichment calls
				externalData := validExternalExampleData()
				enrichment := validEnrichmentData()
				m.On("GetExampleData", mock.Anything, mock.AnythingOfType("string")).
					Return(externalData, nil)
				m.On("EnrichExample", mock.Anything, mock.AnythingOfType("string")).
					Return(enrichment, nil)
				// Async notification
				m.On("NotifyExampleCreated", mock.Anything, mock.AnythingOfType("string"), "john.doe@example.com").
					Return(nil).Maybe()
			},
			wantErr: false,
		},
		{
			name: "external validation fails",
			request: CreateExampleRequest{
				Name:  "invalid",
				Email: "invalid@example.com",
				Age:   25,
			},
			setupService: func(m *mocks.MockExampleService) {
				// No service calls expected when external validation fails
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				m.On("ValidateExample", mock.Anything, "invalid", "invalid@example.com", 25).
					Return(false, nil)
			},
			wantErr:     true,
			errContains: "rejected by external validation",
		},
		{
			name: "external validation error",
			request: CreateExampleRequest{
				Name:  "Test User",
				Email: "test@example.com",
				Age:   25,
			},
			setupService: func(m *mocks.MockExampleService) {
				// No service calls expected when external validation errors
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				m.On("ValidateExample", mock.Anything, "Test User", "test@example.com", 25).
					Return(false, repository.ErrExternalAPIUnavailable)
			},
			wantErr:     true,
			errContains: "external validation failed",
		},
		{
			name:    "validation succeeds but service fails",
			request: validCreateExampleRequest(),
			setupService: func(m *mocks.MockExampleService) {
				m.On("CreateExample", mock.Anything, "John Doe", "john.doe@example.com", 30).
					Return(nil, repository.ErrExampleAlreadyExists)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				m.On("ValidateExample", mock.Anything, "John Doe", "john.doe@example.com", 30).
					Return(true, nil)
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)
			tt.setupExternal(mockExternalAPI)

			ctx := getTestContext()
			result, err := useCase.ValidateAndCreateExample(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.Age, result.Age)
			}

			mockService.AssertExpectations(t)
			// Give some time for async notification to complete
			time.Sleep(10 * time.Millisecond)
			mockExternalAPI.AssertExpectations(t)
		})
	}
}

func TestExampleUseCase_ListExamples(t *testing.T) {
	tests := []struct {
		name          string
		request       ListExamplesRequest
		setupService  func(*mocks.MockExampleService)
		setupExternal func(*mocks.MockExternalExampleAPI)
		wantErr       bool
		errContains   string
		expectedLimit int
	}{
		{
			name: "successful list with enrichment",
			request: ListExamplesRequest{
				Limit:  5,
				Offset: 0,
			},
			setupService: func(m *mocks.MockExampleService) {
				examples := multipleValidExamples()[:3]
				m.On("ListExamples", mock.Anything, 5, 0).Return(examples, 10, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// Each example will be enriched
				externalData := validExternalExampleData()
				enrichment := validEnrichmentData()
				m.On("GetExampleData", mock.Anything, mock.AnythingOfType("string")).
					Return(externalData, nil).Times(3)
				m.On("EnrichExample", mock.Anything, mock.AnythingOfType("string")).
					Return(enrichment, nil).Times(3)
			},
			wantErr:       false,
			expectedLimit: 5,
		},
		{
			name: "zero limit uses default",
			request: ListExamplesRequest{
				Limit:  0,
				Offset: 0,
			},
			setupService: func(m *mocks.MockExampleService) {
				examples := multipleValidExamples()[:3]
				m.On("ListExamples", mock.Anything, 10, 0).Return(examples, 10, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				externalData := validExternalExampleData()
				enrichment := validEnrichmentData()
				m.On("GetExampleData", mock.Anything, mock.AnythingOfType("string")).
					Return(externalData, nil).Times(3)
				m.On("EnrichExample", mock.Anything, mock.AnythingOfType("string")).
					Return(enrichment, nil).Times(3)
			},
			wantErr:       false,
			expectedLimit: 10,
		},
		{
			name: "service fails",
			request: ListExamplesRequest{
				Limit:  5,
				Offset: 0,
			},
			setupService: func(m *mocks.MockExampleService) {
				m.On("ListExamples", mock.Anything, 5, 0).
					Return(nil, 0, repository.ErrExampleNotFound)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// No external calls expected when service fails
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)
			tt.setupExternal(mockExternalAPI)

			ctx := getTestContext()
			result, err := useCase.ListExamples(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedLimit, result.Limit)
				assert.GreaterOrEqual(t, result.Total, len(result.Examples))
			}

			mockService.AssertExpectations(t)
			mockExternalAPI.AssertExpectations(t)
		})
	}
}

func TestExampleUseCase_UpdateExample(t *testing.T) {
	tests := []struct {
		name          string
		inputID       string
		request       UpdateExampleRequest
		setupService  func(*mocks.MockExampleService)
		setupExternal func(*mocks.MockExternalExampleAPI)
		wantErr       bool
		errContains   string
	}{
		{
			name:    "successful update",
			inputID: "test-id",
			request: validUpdateExampleRequest(),
			setupService: func(m *mocks.MockExampleService) {
				example := validExampleWithCustomData("test-id", "John Smith", "john.smith@example.com", 31)
				m.On("UpdateExample", mock.Anything, "test-id", "John Smith", "john.smith@example.com", 31).
					Return(example, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				externalData := validExternalExampleData()
				enrichment := validEnrichmentData()
				m.On("GetExampleData", mock.Anything, "test-id").Return(externalData, nil)
				m.On("EnrichExample", mock.Anything, "test-id").Return(enrichment, nil)
			},
			wantErr: false,
		},
		{
			name:    "service update fails",
			inputID: "non-existent",
			request: validUpdateExampleRequest(),
			setupService: func(m *mocks.MockExampleService) {
				m.On("UpdateExample", mock.Anything, "non-existent", "John Smith", "john.smith@example.com", 31).
					Return(nil, repository.ErrExampleNotFound)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// No external calls expected when service fails
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)
			tt.setupExternal(mockExternalAPI)

			ctx := getTestContext()
			result, err := useCase.UpdateExample(ctx, tt.inputID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.inputID, result.ID)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.Age, result.Age)
			}

			mockService.AssertExpectations(t)
			mockExternalAPI.AssertExpectations(t)
		})
	}
}

func TestExampleUseCase_DeleteExample(t *testing.T) {
	tests := []struct {
		name         string
		inputID      string
		setupService func(*mocks.MockExampleService)
		wantErr      bool
		errContains  string
	}{
		{
			name:    "successful deletion",
			inputID: "test-id",
			setupService: func(m *mocks.MockExampleService) {
				m.On("DeleteExample", mock.Anything, "test-id").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "service deletion fails",
			inputID: "non-existent",
			setupService: func(m *mocks.MockExampleService) {
				m.On("DeleteExample", mock.Anything, "non-existent").
					Return(repository.ErrExampleNotFound)
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)

			ctx := getTestContext()
			err := useCase.DeleteExample(ctx, tt.inputID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestExampleUseCase_GetExampleByEmail(t *testing.T) {
	tests := []struct {
		name          string
		inputEmail    string
		setupService  func(*mocks.MockExampleService)
		setupExternal func(*mocks.MockExternalExampleAPI)
		wantErr       bool
		errContains   string
	}{
		{
			name:       "successful get by email",
			inputEmail: "john@example.com",
			setupService: func(m *mocks.MockExampleService) {
				example := validExampleWithCustomData("test-id", "John Doe", "john@example.com", 30)
				m.On("GetExampleByEmail", mock.Anything, "john@example.com").Return(example, nil)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				externalData := validExternalExampleData()
				enrichment := validEnrichmentData()
				m.On("GetExampleData", mock.Anything, "test-id").Return(externalData, nil)
				m.On("EnrichExample", mock.Anything, "test-id").Return(enrichment, nil)
			},
			wantErr: false,
		},
		{
			name:       "service fails",
			inputEmail: "nonexistent@example.com",
			setupService: func(m *mocks.MockExampleService) {
				m.On("GetExampleByEmail", mock.Anything, "nonexistent@example.com").
					Return(nil, repository.ErrExampleNotFound)
			},
			setupExternal: func(m *mocks.MockExternalExampleAPI) {
				// No external calls expected when service fails
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockExampleService{}
			mockExternalAPI := &mocks.MockExternalExampleAPI{}
			logger := zap.NewNop()
			useCase := NewExampleUseCase(mockService, mockExternalAPI, logger)

			tt.setupService(mockService)
			tt.setupExternal(mockExternalAPI)

			ctx := getTestContext()
			result, err := useCase.GetExampleByEmail(ctx, tt.inputEmail)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.inputEmail, result.Email)
			}

			mockService.AssertExpectations(t)
			mockExternalAPI.AssertExpectations(t)
		})
	}
}
