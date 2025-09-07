package service

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

// Test fixtures for service tests
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

func getTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}

func TestNewExampleService(t *testing.T) {
	mockRepo := &mocks.MockExampleRepository{}
	logger := zap.NewNop()

	service := NewExampleService(mockRepo, logger)

	assert.NotNil(t, service)
}

func TestExampleService_CreateExample(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputEmail  string
		inputAge    int
		setupMock   func(*mocks.MockExampleRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful creation",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			inputAge:   30,
			setupMock: func(m *mocks.MockExampleRepository) {
				// GetByEmail should return not found (email not exists)
				m.On("GetByEmail", mock.Anything, "john@example.com").
					Return(nil, repository.ErrExampleNotFound)
				// Create should succeed
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Example")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "email already exists",
			inputName:  "John Doe",
			inputEmail: "existing@example.com",
			inputAge:   30,
			setupMock: func(m *mocks.MockExampleRepository) {
				existingExample := validExampleWithCustomData("existing-id", "Existing User", "existing@example.com", 25)
				m.On("GetByEmail", mock.Anything, "existing@example.com").
					Return(existingExample, nil)
			},
			wantErr:     true,
			errContains: "already exists",
		},
		{
			name:       "invalid email format",
			inputName:  "John Doe",
			inputEmail: "invalid-email",
			inputAge:   30,
			setupMock: func(m *mocks.MockExampleRepository) {
				// No mock calls expected as validation should fail first
			},
			wantErr:     true,
			errContains: "invalid input",
		},
		{
			name:       "business logic validation fails - corporate email underage",
			inputName:  "Young User",
			inputEmail: "young@corp.com",
			inputAge:   16,
			setupMock: func(m *mocks.MockExampleRepository) {
				// No mock calls expected as business validation should fail first
			},
			wantErr:     true,
			errContains: "business logic validation failed",
		},
		{
			name:       "repository create fails",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			inputAge:   30,
			setupMock: func(m *mocks.MockExampleRepository) {
				m.On("GetByEmail", mock.Anything, "john@example.com").
					Return(nil, repository.ErrExampleNotFound)
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Example")).
					Return(repository.ErrExampleAlreadyExists)
			},
			wantErr:     true,
			errContains: "failed to save example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockExampleRepository{}
			logger := zap.NewNop()
			service := NewExampleService(mockRepo, logger)

			tt.setupMock(mockRepo)

			ctx := getTestContext()
			result, err := service.CreateExample(ctx, tt.inputName, tt.inputEmail, tt.inputAge)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.inputName, result.Name)
				assert.Equal(t, tt.inputEmail, result.Email)
				assert.Equal(t, tt.inputAge, result.Age)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExampleService_GetExampleByID(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		setupMock   func(*mocks.MockExampleRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful get",
			inputID: "test-id",
			setupMock: func(m *mocks.MockExampleRepository) {
				example := validExampleWithCustomData("test-id", "John Doe", "john@example.com", 30)
				m.On("GetByID", mock.Anything, "test-id").Return(example, nil)
			},
			wantErr: false,
		},
		{
			name:    "empty ID",
			inputID: "",
			setupMock: func(m *mocks.MockExampleRepository) {
				// No mock calls expected
			},
			wantErr:     true,
			errContains: "id cannot be empty",
		},
		{
			name:    "example not found",
			inputID: "non-existent-id",
			setupMock: func(m *mocks.MockExampleRepository) {
				m.On("GetByID", mock.Anything, "non-existent-id").
					Return(nil, repository.ErrExampleNotFound)
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockExampleRepository{}
			logger := zap.NewNop()
			service := NewExampleService(mockRepo, logger)

			tt.setupMock(mockRepo)

			ctx := getTestContext()
			result, err := service.GetExampleByID(ctx, tt.inputID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.inputID, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExampleService_UpdateExample(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		inputName   string
		inputEmail  string
		inputAge    int
		setupMock   func(*mocks.MockExampleRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful update",
			inputID:    "test-id",
			inputName:  "Updated Name",
			inputEmail: "updated@example.com",
			inputAge:   35,
			setupMock: func(m *mocks.MockExampleRepository) {
				existing := validExampleWithCustomData("test-id", "Original Name", "original@example.com", 30)
				m.On("GetByID", mock.Anything, "test-id").Return(existing, nil)
				m.On("GetByEmail", mock.Anything, "updated@example.com").
					Return(nil, repository.ErrExampleNotFound) // Email not in use
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Example")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "example not found",
			inputID:    "non-existent",
			inputName:  "Updated Name",
			inputEmail: "updated@example.com",
			inputAge:   35,
			setupMock: func(m *mocks.MockExampleRepository) {
				m.On("GetByID", mock.Anything, "non-existent").
					Return(nil, repository.ErrExampleNotFound)
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:       "email already in use by another example",
			inputID:    "test-id",
			inputName:  "Updated Name",
			inputEmail: "taken@example.com",
			inputAge:   35,
			setupMock: func(m *mocks.MockExampleRepository) {
				existing := validExampleWithCustomData("test-id", "Original Name", "original@example.com", 30)
				other := validExampleWithCustomData("other-id", "Other User", "taken@example.com", 25)
				m.On("GetByID", mock.Anything, "test-id").Return(existing, nil)
				m.On("GetByEmail", mock.Anything, "taken@example.com").Return(other, nil)
			},
			wantErr:     true,
			errContains: "email taken@example.com is already in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockExampleRepository{}
			logger := zap.NewNop()
			service := NewExampleService(mockRepo, logger)

			tt.setupMock(mockRepo)

			ctx := getTestContext()
			result, err := service.UpdateExample(ctx, tt.inputID, tt.inputName, tt.inputEmail, tt.inputAge)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.inputID, result.ID)
				assert.Equal(t, tt.inputName, result.Name)
				assert.Equal(t, tt.inputEmail, result.Email)
				assert.Equal(t, tt.inputAge, result.Age)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExampleService_DeleteExample(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		setupMock   func(*mocks.MockExampleRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful deletion",
			inputID: "test-id",
			setupMock: func(m *mocks.MockExampleRepository) {
				existing := validExampleWithCustomData("test-id", "John Doe", "john@example.com", 30)
				m.On("GetByID", mock.Anything, "test-id").Return(existing, nil)
				m.On("Delete", mock.Anything, "test-id").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "empty ID",
			inputID: "",
			setupMock: func(m *mocks.MockExampleRepository) {
				// No mock calls expected
			},
			wantErr:     true,
			errContains: "id cannot be empty",
		},
		{
			name:    "example not found",
			inputID: "non-existent",
			setupMock: func(m *mocks.MockExampleRepository) {
				m.On("GetByID", mock.Anything, "non-existent").
					Return(nil, repository.ErrExampleNotFound)
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockExampleRepository{}
			logger := zap.NewNop()
			service := NewExampleService(mockRepo, logger)

			tt.setupMock(mockRepo)

			ctx := getTestContext()
			err := service.DeleteExample(ctx, tt.inputID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExampleService_ListExamples(t *testing.T) {
	tests := []struct {
		name        string
		inputLimit  int
		inputOffset int
		setupMock   func(*mocks.MockExampleRepository)
		wantErr     bool
		errContains string
		expectLimit int
	}{
		{
			name:        "successful list",
			inputLimit:  5,
			inputOffset: 0,
			setupMock: func(m *mocks.MockExampleRepository) {
				examples := multipleValidExamples()[:3]
				m.On("List", mock.Anything, 5, 0).Return(examples, nil)
				m.On("Count", mock.Anything).Return(10, nil)
			},
			wantErr:     false,
			expectLimit: 5,
		},
		{
			name:        "zero limit uses default",
			inputLimit:  0,
			inputOffset: 0,
			setupMock: func(m *mocks.MockExampleRepository) {
				examples := multipleValidExamples()[:3]
				m.On("List", mock.Anything, 10, 0).Return(examples, nil) // Default limit is 10
				m.On("Count", mock.Anything).Return(10, nil)
			},
			wantErr:     false,
			expectLimit: 10,
		},
		{
			name:        "limit exceeds max uses max",
			inputLimit:  200,
			inputOffset: 0,
			setupMock: func(m *mocks.MockExampleRepository) {
				examples := multipleValidExamples()[:3]
				m.On("List", mock.Anything, 100, 0).Return(examples, nil) // Max limit is 100
				m.On("Count", mock.Anything).Return(10, nil)
			},
			wantErr:     false,
			expectLimit: 100,
		},
		{
			name:        "negative offset uses zero",
			inputLimit:  10,
			inputOffset: -5,
			setupMock: func(m *mocks.MockExampleRepository) {
				examples := multipleValidExamples()[:3]
				m.On("List", mock.Anything, 10, 0).Return(examples, nil) // Offset becomes 0
				m.On("Count", mock.Anything).Return(10, nil)
			},
			wantErr:     false,
			expectLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockExampleRepository{}
			logger := zap.NewNop()
			service := NewExampleService(mockRepo, logger)

			tt.setupMock(mockRepo)

			ctx := getTestContext()
			examples, total, err := service.ListExamples(ctx, tt.inputLimit, tt.inputOffset)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, examples)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, examples)
				assert.Greater(t, total, 0)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExampleService_ValidateExampleBusinessRules(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputEmail  string
		inputAge    int
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid example",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			inputAge:   30,
			wantErr:    false,
		},
		{
			name:        "name contains profanity",
			inputName:   "badword1",
			inputEmail:  "test@example.com",
			inputAge:    25,
			wantErr:     true,
			errContains: "inappropriate content",
		},
		{
			name:        "corporate email underage",
			inputName:   "Young User",
			inputEmail:  "young@corp.com",
			inputAge:    16,
			wantErr:     true,
			errContains: "corporate accounts require minimum age of 18",
		},
		{
			name:        "VIP domain underage",
			inputName:   "Young VIP",
			inputEmail:  "young@vip.com",
			inputAge:    19,
			wantErr:     true,
			errContains: "VIP accounts require minimum age of 21",
		},
		{
			name:       "corporate email valid age",
			inputName:  "Adult User",
			inputEmail: "adult@corp.com",
			inputAge:   25,
			wantErr:    false,
		},
		{
			name:       "VIP domain valid age",
			inputName:  "VIP User",
			inputEmail: "vip@vip.com",
			inputAge:   25,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockExampleRepository{}
			logger := zap.NewNop()
			service := NewExampleService(mockRepo, logger)

			ctx := getTestContext()
			err := service.ValidateExampleBusinessRules(ctx, tt.inputName, tt.inputEmail, tt.inputAge)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function tests
func TestGenerateExampleID(t *testing.T) {
	id := generateExampleID("John Doe", "john@example.com")
	expected := "ex_joh_8" // First 3 chars of email + length of name
	assert.Equal(t, expected, id)
}

func TestContainsProfanity(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"clean name", "John Doe", false},
		{"contains profanity", "badword1", true},
		{"case sensitive", "BADWORD1", false},        // Function is case sensitive
		{"partial match", "somebadword1text", false}, // Exact match only
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsProfanity(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsCorporateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"regular email", "user@gmail.com", false},
		{"corporate email", "user@corp.com", true},
		{"enterprise email", "user@enterprise.com", true},
		{"partial match", "user@mycorp.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCorporateEmail(tt.email)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsVIPDomain(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"regular email", "user@gmail.com", false},
		{"VIP email", "user@vip.com", true},
		{"premium email", "user@premium.com", true},
		{"partial match", "user@myvip.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVIPDomain(tt.email)
			assert.Equal(t, tt.want, got)
		})
	}
}
