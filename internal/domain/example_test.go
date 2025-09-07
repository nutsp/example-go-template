package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExample(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		exampleName string
		email       string
		age         int
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid example",
			id:          "test-id",
			exampleName: "John Doe",
			email:       "john@example.com",
			age:         30,
			wantErr:     false,
		},
		{
			name:        "empty name",
			id:          "test-id",
			exampleName: "",
			email:       "john@example.com",
			age:         30,
			wantErr:     true,
			errMsg:      "name cannot be empty",
		},
		{
			name:        "name too long",
			id:          "test-id",
			exampleName: "This is a very long name that exceeds the maximum allowed length of 100 characters for testing purposes",
			email:       "john@example.com",
			age:         30,
			wantErr:     true,
			errMsg:      "name cannot exceed 100 characters",
		},
		{
			name:        "empty email",
			id:          "test-id",
			exampleName: "John Doe",
			email:       "",
			age:         30,
			wantErr:     true,
			errMsg:      "email cannot be empty",
		},
		{
			name:        "invalid email format",
			id:          "test-id",
			exampleName: "John Doe",
			email:       "invalid-email",
			age:         30,
			wantErr:     true,
			errMsg:      "invalid email format",
		},
		{
			name:        "negative age",
			id:          "test-id",
			exampleName: "John Doe",
			email:       "john@example.com",
			age:         -1,
			wantErr:     true,
			errMsg:      "age cannot be negative",
		},
		{
			name:        "age too high",
			id:          "test-id",
			exampleName: "John Doe",
			email:       "john@example.com",
			age:         151,
			wantErr:     true,
			errMsg:      "age cannot exceed 150",
		},
		{
			name:        "boundary age 0",
			id:          "test-id",
			exampleName: "Baby Doe",
			email:       "baby@example.com",
			age:         0,
			wantErr:     false,
		},
		{
			name:        "boundary age 150",
			id:          "test-id",
			exampleName: "Old Doe",
			email:       "old@example.com",
			age:         150,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			example, err := NewExample(tt.id, tt.exampleName, tt.email, tt.age)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, example)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, example)
				assert.Equal(t, tt.id, example.ID)
				assert.Equal(t, tt.exampleName, example.Name)
				assert.Equal(t, tt.email, example.Email)
				assert.Equal(t, tt.age, example.Age)
				assert.WithinDuration(t, time.Now(), example.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), example.UpdatedAt, time.Second)
				assert.Equal(t, example.CreatedAt, example.UpdatedAt)
			}
		})
	}
}

func TestExample_Update(t *testing.T) {
	// Create a valid example first
	example, err := NewExample("test-id", "John Doe", "john@example.com", 30)
	require.NoError(t, err)

	originalCreatedAt := example.CreatedAt
	originalUpdatedAt := example.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name     string
		newName  string
		newEmail string
		newAge   int
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid update",
			newName:  "Jane Doe",
			newEmail: "jane@example.com",
			newAge:   25,
			wantErr:  false,
		},
		{
			name:     "empty name",
			newName:  "",
			newEmail: "jane@example.com",
			newAge:   25,
			wantErr:  true,
			errMsg:   "name cannot be empty",
		},
		{
			name:     "invalid email",
			newName:  "Jane Doe",
			newEmail: "invalid-email",
			newAge:   25,
			wantErr:  true,
			errMsg:   "invalid email format",
		},
		{
			name:     "negative age",
			newName:  "Jane Doe",
			newEmail: "jane@example.com",
			newAge:   -1,
			wantErr:  true,
			errMsg:   "age cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh example for each test
			testExample, _ := NewExample("test-id", "John Doe", "john@example.com", 30)
			testExample.CreatedAt = originalCreatedAt
			testExample.UpdatedAt = originalUpdatedAt

			time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt will be different

			err := testExample.Update(tt.newName, tt.newEmail, tt.newAge)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				// Fields should not be updated on error
				assert.Equal(t, "John Doe", testExample.Name)
				assert.Equal(t, "john@example.com", testExample.Email)
				assert.Equal(t, 30, testExample.Age)
				assert.Equal(t, originalUpdatedAt, testExample.UpdatedAt)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newName, testExample.Name)
				assert.Equal(t, tt.newEmail, testExample.Email)
				assert.Equal(t, tt.newAge, testExample.Age)
				assert.Equal(t, originalCreatedAt, testExample.CreatedAt)      // CreatedAt should not change
				assert.True(t, testExample.UpdatedAt.After(originalUpdatedAt)) // UpdatedAt should change
			}
		})
	}
}

func TestValidateExample(t *testing.T) {
	tests := []struct {
		name    string
		exName  string
		email   string
		age     int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid data",
			exName:  "John Doe",
			email:   "john@example.com",
			age:     30,
			wantErr: false,
		},
		{
			name:    "empty name",
			exName:  "",
			email:   "john@example.com",
			age:     30,
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "empty email",
			exName:  "John Doe",
			email:   "",
			age:     30,
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "negative age",
			exName:  "John Doe",
			email:   "john@example.com",
			age:     -1,
			wantErr: true,
			errMsg:  "age cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExample(tt.exName, tt.email, tt.age)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "valid email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
			want:  true,
		},
		{
			name:  "valid email with numbers",
			email: "user123@example.com",
			want:  true,
		},
		{
			name:  "valid email with dots",
			email: "user.name@example.com",
			want:  true,
		},
		{
			name:  "invalid email no @",
			email: "userexample.com",
			want:  false,
		},
		{
			name:  "invalid email no domain",
			email: "user@",
			want:  false,
		},
		{
			name:  "invalid email no user",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "invalid email no TLD",
			email: "user@example",
			want:  false,
		},
		{
			name:  "empty email",
			email: "",
			want:  false,
		},
		{
			name:  "invalid email multiple @",
			email: "user@@example.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidEmail(tt.email)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExample_String(t *testing.T) {
	example, err := NewExample("test-123", "John Doe", "john@example.com", 30)
	require.NoError(t, err)

	str := example.String()
	expected := "Example{ID: test-123, Name: John Doe, Email: john@example.com, Age: 30}"
	assert.Equal(t, expected, str)
}

func TestExample_CreatedAtUpdatedAt(t *testing.T) {
	example, err := NewExample("test-id", "John Doe", "john@example.com", 30)
	require.NoError(t, err)

	// CreatedAt and UpdatedAt should be the same initially
	assert.Equal(t, example.CreatedAt, example.UpdatedAt)

	// Wait and update
	time.Sleep(10 * time.Millisecond)
	err = example.Update("Jane Doe", "jane@example.com", 25)
	require.NoError(t, err)

	// UpdatedAt should be after CreatedAt
	assert.True(t, example.UpdatedAt.After(example.CreatedAt))
}

// Benchmark tests
func BenchmarkNewExample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewExample("test-id", "John Doe", "john@example.com", 30)
	}
}

func BenchmarkExample_Update(b *testing.B) {
	example, _ := NewExample("test-id", "John Doe", "john@example.com", 30)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = example.Update("Jane Doe", "jane@example.com", 25)
	}
}

func BenchmarkIsValidEmail(b *testing.B) {
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isValidEmail(email)
	}
}
