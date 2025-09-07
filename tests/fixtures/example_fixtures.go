package fixtures

import (
	"context"
	"fmt"
	"time"

	"example-api-template/internal/domain"
	"example-api-template/internal/repository"
)

// Example domain fixtures

// ValidExample returns a valid example for testing
func ValidExample() *domain.Example {
	example, _ := domain.NewExample(
		"ex_test_123",
		"John Doe",
		"john.doe@example.com",
		30,
	)
	return example
}

// ValidExampleWithCustomData returns a valid example with custom data
func ValidExampleWithCustomData(id, name, email string, age int) *domain.Example {
	example, _ := domain.NewExample(id, name, email, age)
	return example
}

// MultipleValidExamples returns multiple valid examples for testing
func MultipleValidExamples() []*domain.Example {
	return []*domain.Example{
		ValidExampleWithCustomData("ex_001", "Alice Smith", "alice@example.com", 25),
		ValidExampleWithCustomData("ex_002", "Bob Johnson", "bob@example.com", 35),
		ValidExampleWithCustomData("ex_003", "Carol Williams", "carol@example.com", 28),
		ValidExampleWithCustomData("ex_004", "David Brown", "david@example.com", 42),
		ValidExampleWithCustomData("ex_005", "Eva Davis", "eva@example.com", 31),
	}
}

// UseCase fixtures - moved to individual test files to avoid import cycles

// External API fixtures

// ValidExternalExampleData returns valid external data
func ValidExternalExampleData() *repository.ExternalExampleData {
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

// ValidEnrichmentData returns valid enrichment data
func ValidEnrichmentData() map[string]interface{} {
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

// Invalid fixtures for negative testing

// InvalidExampleEmptyName returns an example with empty name
func InvalidExampleEmptyName() (string, string, string, int) {
	return "ex_invalid_001", "", "test@example.com", 25
}

// InvalidExampleInvalidEmail returns an example with invalid email
func InvalidExampleInvalidEmail() (string, string, string, int) {
	return "ex_invalid_002", "Test User", "invalid-email", 25
}

// InvalidExampleNegativeAge returns an example with negative age
func InvalidExampleNegativeAge() (string, string, string, int) {
	return "ex_invalid_003", "Test User", "test@example.com", -5
}

// InvalidExampleExcessiveAge returns an example with excessive age
func InvalidExampleExcessiveAge() (string, string, string, int) {
	return "ex_invalid_004", "Test User", "test@example.com", 200
}

// InvalidExampleLongName returns an example with overly long name
func InvalidExampleLongName() (string, string, string, int) {
	longName := "This is a very long name that exceeds the maximum allowed length for a person's name in our system"
	return "ex_invalid_005", longName, "test@example.com", 25
}

// Test data sets

// TestDataSet represents a set of test data
type TestDataSet struct {
	Name        string
	Examples    []*domain.Example
	Description string
}

// GetTestDataSets returns various test data sets
func GetTestDataSets() []TestDataSet {
	return []TestDataSet{
		{
			Name:        "empty_set",
			Examples:    []*domain.Example{},
			Description: "Empty data set for testing edge cases",
		},
		{
			Name:        "single_example",
			Examples:    []*domain.Example{ValidExample()},
			Description: "Single example for basic testing",
		},
		{
			Name:        "multiple_examples",
			Examples:    MultipleValidExamples(),
			Description: "Multiple examples for list operations",
		},
		{
			Name: "mixed_ages",
			Examples: []*domain.Example{
				ValidExampleWithCustomData("ex_young", "Young Person", "young@example.com", 18),
				ValidExampleWithCustomData("ex_middle", "Middle Aged", "middle@example.com", 45),
				ValidExampleWithCustomData("ex_senior", "Senior Person", "senior@example.com", 70),
			},
			Description: "Examples with different age ranges",
		},
		{
			Name: "different_domains",
			Examples: []*domain.Example{
				ValidExampleWithCustomData("ex_gmail", "Gmail User", "user@gmail.com", 25),
				ValidExampleWithCustomData("ex_yahoo", "Yahoo User", "user@yahoo.com", 30),
				ValidExampleWithCustomData("ex_corp", "Corp User", "user@corp.com", 35),
			},
			Description: "Examples with different email domains",
		},
	}
}

// Benchmark fixtures

// GenerateLargeDataSet generates a large data set for performance testing
func GenerateLargeDataSet(size int) []*domain.Example {
	examples := make([]*domain.Example, size)
	for i := 0; i < size; i++ {
		example, _ := domain.NewExample(
			fmt.Sprintf("ex_perf_%06d", i),
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			20+(i%50), // Age between 20 and 69
		)
		examples[i] = example
	}
	return examples
}

// Error scenarios

// ErrorScenarios returns common error scenarios for testing
func ErrorScenarios() map[string]error {
	return map[string]error{
		"not_found":                repository.ErrExampleNotFound,
		"already_exists":           repository.ErrExampleAlreadyExists,
		"external_api_unavailable": repository.ErrExternalAPIUnavailable,
		"external_api_timeout":     repository.ErrExternalAPITimeout,
		"invalid_external_data":    repository.ErrInvalidExternalData,
	}
}

// Context helpers

// GetTestContext returns a context with timeout for testing
func GetTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}

// GetTestContextWithCancel returns a context with cancel function
func GetTestContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
